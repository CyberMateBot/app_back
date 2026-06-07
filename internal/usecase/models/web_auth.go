package models

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrWebAccountAlreadyExists = errors.New("ErrWebAccountAlreadyExists")
	ErrWebAccountNotFound      = errors.New("ErrWebAccountNotFound")
	ErrInvalidCredentials      = errors.New("ErrInvalidCredentials")
)

type RegisterWebAccountInput struct {
	Email       string
	Password    string
	DisplayName string
}

func (i *RegisterWebAccountInput) Validate() error {
	i.Email = strings.TrimSpace(strings.ToLower(i.Email))
	i.DisplayName = strings.TrimSpace(i.DisplayName)

	if i.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if len(i.Email) > 320 {
		return fmt.Errorf("%w: email too long", ErrInvalidInput)
	}
	if i.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	if len(i.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 chars", ErrInvalidInput)
	}
	if len(i.Password) > 200 {
		return fmt.Errorf("%w: password too long", ErrInvalidInput)
	}
	if len(i.DisplayName) > 200 {
		return fmt.Errorf("%w: display_name too long", ErrInvalidInput)
	}
	return nil
}

type LoginWebAccountInput struct {
	Email    string
	Password string
}

func (i *LoginWebAccountInput) Validate() error {
	i.Email = strings.TrimSpace(strings.ToLower(i.Email))
	if i.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if i.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	return nil
}

type AuthTokensOutput struct {
	AccessToken string
}

type GetWebAccountOutput struct {
	Data WebAccount
}

type WebAccount struct {
	ID          int64
	Email       string
	DisplayName string
}

type CreateWebPromptInput struct {
	WebAccountID int64
	Prompt       string
	Category     string
	Model        string
}

func (i *CreateWebPromptInput) Validate() error {
	if i.WebAccountID <= 0 {
		return fmt.Errorf("%w: web_account_id is required", ErrInvalidInput)
	}
	i.Prompt = strings.TrimSpace(i.Prompt)
	if i.Prompt == "" {
		return fmt.Errorf("%w: prompt is required", ErrInvalidInput)
	}
	if len(i.Prompt) > 20000 {
		return fmt.Errorf("%w: prompt too long", ErrInvalidInput)
	}
	if len(i.Category) > 100 {
		return fmt.Errorf("%w: category too long", ErrInvalidInput)
	}
	if len(i.Model) > 100 {
		return fmt.Errorf("%w: model too long", ErrInvalidInput)
	}
	return nil
}

type CreateWebPromptOutput struct {
	ID int64
}

type ListWebPromptsInput struct {
	WebAccountID int64
	Limit        int32
	Offset       int32
}

func (i *ListWebPromptsInput) Validate() error {
	if i.WebAccountID <= 0 {
		return fmt.Errorf("%w: web_account_id is required", ErrInvalidInput)
	}
	return nil
}

type ListWebPromptsOutput struct {
	Items []WebPrompt
}

type WebPrompt struct {
	ID       int64
	Prompt   string
	Category string
	Model    string
}

