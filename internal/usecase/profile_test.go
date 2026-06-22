package usecase

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"
    "net/url"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/twelvepills-936/tgapp-/internal"
    repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
    ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
    "github.com/twelvepills-936/tgapp-/pkg/config"
)

type fakeTx struct{}

func (f *fakeTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (f *fakeTx) Commit(ctx context.Context) error { return nil }
func (f *fakeTx) Rollback(ctx context.Context) error { return nil }
func (f *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) { return 0, nil }
func (f *fakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) { return nil, nil }
func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (f *fakeTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row { return nil }
func (f *fakeTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, fn func(pgx.Row) error) (pgconn.CommandTag, error) { return pgconn.CommandTag{}, nil }
func (f *fakeTx) Conn() *pgx.Conn { return nil }

type fakeRepoProfile struct{
    fakeAdminRepoStubs
    exists map[string]repoModels.Profile
    addCalls [][2]int64
    creditCalls []struct {
        profileID int64
        amount    int64
        reason    string
    }
    registrationBonus           int64
    hasRegistrationBonusSetting bool
}

func (f *fakeRepoProfile) DBBeginTransaction(ctx context.Context) (pgx.Tx, error) { return &fakeTx{}, nil }
func (f *fakeRepoProfile) ReadUser(ctx context.Context, id int64, dbTx pgx.Tx) (repoModels.User, error) { return repoModels.User{}, nil }
func (f *fakeRepoProfile) CreateProfile(ctx context.Context, tx pgx.Tx, p repoModels.Profile) (int64, error) { f.exists[p.TelegramID] = p; return 1, nil }
func (f *fakeRepoProfile) GetProfileByTelegramID(ctx context.Context, tx pgx.Tx, telegramID string) (repoModels.Profile, error) {
    if p, ok := f.exists[telegramID]; ok { return p, nil }
    return repoModels.Profile{}, pgx.ErrNoRows
}
func (f *fakeRepoProfile) CreateWalletForUser(ctx context.Context, tx pgx.Tx, profileID int64) (int64, error) { return 1, nil }
func (f *fakeRepoProfile) GetAdminSettings(ctx context.Context, tx pgx.Tx) (map[string]json.RawMessage, error) {
    bonus := int64(10)
    if f.hasRegistrationBonusSetting {
        bonus = f.registrationBonus
    }
    raw, err := json.Marshal(bonus)
    if err != nil {
        return nil, err
    }
    return map[string]json.RawMessage{
        "registration_bonus": raw,
    }, nil
}
func (f *fakeRepoProfile) CreditProfileTokens(ctx context.Context, tx pgx.Tx, profileID, adminID int64, amount int64, reason string) (repoModels.TokenOperationResult, error) {
    f.creditCalls = append(f.creditCalls, struct {
        profileID int64
        amount    int64
        reason    string
    }{profileID, amount, reason})
    return repoModels.TokenOperationResult{ProfileID: profileID, BalanceAfter: amount}, nil
}
func (f *fakeRepoProfile) AddReferral(ctx context.Context, tx pgx.Tx, referrerProfileID int64, refereeProfileID int64) error {
    f.addCalls = append(f.addCalls, [2]int64{referrerProfileID, refereeProfileID})
    return nil
}
func (f *fakeRepoProfile) ListReferralsByReferrerProfileID(ctx context.Context, tx pgx.Tx, referrerProfileID int64) ([]repoModels.Referral, error) {
    return nil, nil
}
func (f *fakeRepoProfile) UpdateProfileTheme(ctx context.Context, tx pgx.Tx, telegramID, theme string) error {
    p, ok := f.exists[telegramID]
    if !ok {
        return pgx.ErrNoRows
    }
    p.UITheme = theme
    f.exists[telegramID] = p
    return nil
}

// Web site methods (unused in this test)
func (f *fakeRepoProfile) CreateWebAccount(ctx context.Context, tx pgx.Tx, a repoModels.WebAccount) (int64, error) { return 0, errors.New("not impl") }
func (f *fakeRepoProfile) GetWebAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (repoModels.WebAccount, error) { return repoModels.WebAccount{}, errors.New("not impl") }
func (f *fakeRepoProfile) GetWebAccountByID(ctx context.Context, tx pgx.Tx, id int64) (repoModels.WebAccount, error) { return repoModels.WebAccount{}, errors.New("not impl") }
func (f *fakeRepoProfile) CreateWebPrompt(ctx context.Context, tx pgx.Tx, p repoModels.WebPrompt) (int64, error) { return 0, errors.New("not impl") }
func (f *fakeRepoProfile) ListWebPrompts(ctx context.Context, tx pgx.Tx, webAccountID int64, limit int32, offset int32) ([]repoModels.WebPrompt, error) {
    return nil, errors.New("not impl")
}

func TestRegisterByTelegram_CreatesProfile(t *testing.T) {
    var _ internal.Repository = (*fakeRepoProfile)(nil)
    repo := &fakeRepoProfile{
        exists:                      map[string]repoModels.Profile{},
        registrationBonus:           10,
        hasRegistrationBonusSetting: true,
    }
    uc := NewUseCase(repo, config.ConfigJWT{})

    values := url.Values{}
    values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan","photo_url":"","language_code":"ru"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    out, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if out.ProfileID == 0 { t.Fatalf("expected profile id > 0") }
    if len(repo.creditCalls) != 1 {
        t.Fatalf("expected one registration bonus credit, got %+v", repo.creditCalls)
    }
    if repo.creditCalls[0].profileID != 1 || repo.creditCalls[0].amount != 10 || repo.creditCalls[0].reason != "registration_bonus" {
        t.Fatalf("unexpected credit call: %+v", repo.creditCalls[0])
    }
}

func TestRegisterByTelegram_SkipsZeroRegistrationBonus(t *testing.T) {
    repo := &fakeRepoProfile{
        exists:                      map[string]repoModels.Profile{},
        registrationBonus:           0,
        hasRegistrationBonusSetting: true,
    }
    uc := NewUseCase(repo, config.ConfigJWT{})

    values := url.Values{}
    values.Set("user", `{"id":456,"first_name":"Anna","username":"anna"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    _, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(repo.creditCalls) != 0 {
        t.Fatalf("expected no credit when bonus is zero, got %+v", repo.creditCalls)
    }
}

func TestRegisterByTelegram_AlreadyExists(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{"123":{TelegramID:"123"}}}
    uc := NewUseCase(repo, config.ConfigJWT{})

    values := url.Values{}
    values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    _, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
    if !errors.Is(err, ucModels.ErrProfileAlreadyRegistered) {
        t.Fatalf("expected ErrProfileAlreadyRegistered, got %v", err)
    }
}

func TestRegisterByTelegram_InvalidInitData_NoUserParam(t *testing.T) {
	repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}}
	uc := NewUseCase(repo, config.ConfigJWT{})

	values := url.Values{}
	values.Set("foo", "bar")
	initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

	_, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{InitDataRaw: initRaw})
	if !errors.Is(err, ucModels.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetUserByTelegramID_NotFound(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{}}
    uc := NewUseCase(repo, config.ConfigJWT{})
    _, err := uc.GetUserByTelegramID(context.Background(), "not-exists")
    if !errors.Is(err, ucModels.ErrProfileNotFound) {
        t.Fatalf("expected ErrProfileNotFound, got %v", err)
    }
}

func TestRegisterByTelegram_LinksReferral(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{
        "777000": {ID: 10, TelegramID: "777000"},
    }}
    uc := NewUseCase(repo, config.ConfigJWT{})

    values := url.Values{}
    values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    _, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{
        InitDataRaw: initRaw,
        StartParam:  "ref_777000",
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(repo.addCalls) != 1 || repo.addCalls[0] != [2]int64{10, 1} {
        t.Fatalf("unexpected addCalls: %+v", repo.addCalls)
    }
}

func TestRegisterByTelegram_AlreadyExistsStillLinksReferral(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{
        "123":    {ID: 1, TelegramID: "123"},
        "777000": {ID: 10, TelegramID: "777000"},
    }}
    uc := NewUseCase(repo, config.ConfigJWT{})

    values := url.Values{}
    values.Set("user", `{"id":123,"first_name":"Ivan","username":"ivan"}`)
    initRaw := base64.StdEncoding.EncodeToString([]byte(values.Encode()))

    _, err := uc.RegisterByTelegram(context.Background(), ucModels.RegisterByTelegramInput{
        InitDataRaw: initRaw,
        StartParam:  "ref_777000",
    })
    if !errors.Is(err, ucModels.ErrProfileAlreadyRegistered) {
        t.Fatalf("expected ErrProfileAlreadyRegistered, got %v", err)
    }
    if len(repo.addCalls) != 1 || repo.addCalls[0] != [2]int64{10, 1} {
        t.Fatalf("unexpected addCalls: %+v", repo.addCalls)
    }
}

func TestUpdateProfileTheme_OK(t *testing.T) {
    repo := &fakeRepoProfile{exists: map[string]repoModels.Profile{
        "42": {TelegramID: "42", UITheme: ucModels.ThemeLight},
    }}
    uc := NewUseCase(repo, config.ConfigJWT{})

    out, err := uc.UpdateProfileTheme(context.Background(), ucModels.UpdateProfileThemeInput{
        TelegramID: "42",
        Theme:      ucModels.ThemeDark,
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if out.Theme != ucModels.ThemeDark {
        t.Fatalf("expected dark, got %s", out.Theme)
    }
}


