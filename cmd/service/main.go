package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/twelvepills-936/tgapp-/internal/bot"
	"github.com/twelvepills-936/tgapp-/internal/migrations"
	"github.com/twelvepills-936/tgapp-/internal/repository"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	"github.com/twelvepills-936/tgapp-/internal/service"
	"github.com/twelvepills-936/tgapp-/internal/usecase"
	"github.com/twelvepills-936/tgapp-/pkg/adminapi"
	"github.com/twelvepills-936/tgapp-/pkg/ai"
	api "github.com/twelvepills-936/tgapp-/pkg/api"
	"github.com/twelvepills-936/tgapp-/pkg/app"
	"github.com/twelvepills-936/tgapp-/pkg/applinks"
	"github.com/twelvepills-936/tgapp-/pkg/config"
	"github.com/twelvepills-936/tgapp-/pkg/cors"
	"github.com/twelvepills-936/tgapp-/pkg/feedbackapi"
	"github.com/twelvepills-936/tgapp-/pkg/generate"
	"github.com/twelvepills-936/tgapp-/pkg/health"
	"github.com/twelvepills-936/tgapp-/pkg/logger"
	"github.com/twelvepills-936/tgapp-/pkg/mediadownload"
	"github.com/twelvepills-936/tgapp-/pkg/paymentsapi"
	"github.com/twelvepills-936/tgapp-/pkg/prompthistory"
	"github.com/twelvepills-936/tgapp-/pkg/ratelimit"
	"github.com/twelvepills-936/tgapp-/pkg/siteapi"
	"github.com/twelvepills-936/tgapp-/pkg/swagger"
	"github.com/twelvepills-936/tgapp-/pkg/tokenguard"
	"github.com/twelvepills-936/tgapp-/pkg/walletapi"
	"github.com/twelvepills-936/tgapp-/pkg/yookassa"
)

func main() {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	cfg := app.LoadConfigFromEnv()

	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	addConfig := config.LoadConfig()
	slog.InfoContext(ctx, "cors configured",
		slog.Bool("allow_all", config.CORSAllowsAll(addConfig.CORS.AllowedOrigins)),
		slog.Int("origins_count", len(addConfig.CORS.AllowedOrigins)),
	)

	tgBot, err := bot.New()
	if err != nil {
		slog.WarnContext(ctx, "failed to init bot", logger.ErrorAttr(err))
	} else if tgBot != nil && tgBot.Active() {
		if bot.BotPollingEnabled() {
			if err := tgBot.PreparePolling(ctx); err != nil {
				slog.WarnContext(ctx, "telegram polling disabled", logger.ErrorAttr(err))
			} else {
				go tgBot.StartPolling(ctx)
			}
		} else if webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL"); webhookURL != "" {
			if err := tgBot.SetWebhook(ctx, webhookURL); err != nil {
				slog.WarnContext(ctx, "telegram webhook not configured", logger.ErrorAttr(err))
			}
		}
	}

	pool, err := repository.NewPostgres(ctx, repoModels.ConfigPostgres(addConfig.Postgres))
	if err != nil {
		slog.ErrorContext(ctx, "failed to init postgres", logger.ErrorAttr(err))
		return
	}
	defer pool.Close()
	if err := migrations.ApplyAdminPanel(ctx, pool); err != nil {
		slog.WarnContext(ctx, "admin panel schema bootstrap failed", logger.ErrorAttr(err))
	}
	repo := repository.NewRepository(pool)

	yookassaClient := yookassa.New(addConfig.YooKassa.ShopID, addConfig.YooKassa.SecretKey)
	if !yookassaClient.Enabled() {
		slog.WarnContext(ctx, "yookassa is not configured (YOOKASSA_SHOP_ID/YOOKASSA_SECRET_KEY missing); checkout endpoint will return 503")
	}

	// Create single instances of usecase and service
	uc := usecase.NewUseCase(
		repo,
		addConfig.JWT,
		usecase.WithYooKassa(yookassaClient, addConfig.YooKassa.ReturnURL),
		usecase.WithTelegramBotToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
	)
	if err := uc.BootstrapAdmin(ctx); err != nil {
		slog.WarnContext(ctx, "admin bootstrap skipped", logger.ErrorAttr(err))
	}
	svc := service.NewService(uc)

	// Register gRPC services BEFORE starting the server
	api.RegisterUsersServer(application.GrpcServer, svc)
	api.RegisterCyberMateServer(application.GrpcServer, svc)

	err = application.Init(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to init app", logger.ErrorAttr(err))
		return
	}
	err = api.RegisterUsersHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register users handler", logger.ErrorAttr(err))
		return
	}

	err = api.RegisterCyberMateHandler(ctx, application.ServeMux, application.GrpcConn)
	if err != nil {
		slog.ErrorContext(ctx, "failed to register cybermate handler", logger.ErrorAttr(err))
		return
	}

	aiSvc := ai.NewService(config.LoadAIConfig())
	promptHistory := prompthistory.NewStore(pool)
	tokenGuard := tokenguard.New(pool)

	// A small, deliberately conservative set of rate limits on endpoints
	// that are either unauthenticated or brute-forceable: admin login
	// (credential stuffing), account registration (bonus/abuse farming),
	// web auth (credential stuffing), and the unauthenticated media-prepare
	// endpoint (memory DoS via repeated large uploads). Generation and other
	// initData-guarded endpoints already require a verified Telegram
	// identity per request, which is a stronger control than IP-based
	// limiting alone.
	rateLimitRules := []ratelimit.Rule{
		{Method: http.MethodPost, Path: "/api/admin/auth/login", Limit: 10, Window: 5 * time.Minute},
		{Method: http.MethodPost, Path: "/v1/register", Limit: 20, Window: 5 * time.Minute},
		{Method: http.MethodPost, Path: "/v1/site/auth/login", Limit: 10, Window: 5 * time.Minute},
		{Method: http.MethodPost, Path: "/v1/site/auth/register", Limit: 10, Window: 5 * time.Minute},
		{Method: http.MethodPost, Path: "/v1/media/download/prepare", Limit: 30, Window: 5 * time.Minute},
		{Method: http.MethodPost, Path: "/v1/payments/yookassa/webhook", Limit: 120, Window: time.Minute},
	}

	httpHandler := bot.HTTPWrap(cors.Wrap(
		ratelimit.Wrap(health.Wrap(
			walletapi.Wrap(
				adminapi.Wrap(
					feedbackapi.Wrap(
						paymentsapi.Wrap(
							mediadownload.Wrap(
								generate.Wrap(
									prompthistory.Wrap(
										applinks.Wrap(
											siteapi.Wrap(
												swagger.Wrap(application.ServeMux, addConfig.App.SwaggerEnabled),
												uc,
												addConfig.JWT,
											),
											addConfig.App,
											uc,
											tokenGuard,
										),
										promptHistory,
										tokenGuard,
									),
									aiSvc,
									promptHistory,
									tokenGuard,
								),
							),
							uc,
							tokenGuard,
						),
						uc,
						tokenGuard,
					),
					uc,
					addConfig.JWT,
					tgBot,
				),
				pool,
				tokenGuard,
			),
		), rateLimitRules),
		addConfig.CORS,
	), tgBot)
	application.SetHTTPHandler(httpHandler)

	err = application.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run app", logger.ErrorAttr(err))
		return
	}
}
