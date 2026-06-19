package prompthistory

import "time"

// Item is a single prompt history record for the Mini App.
type Item struct {
	ID         int64     `json:"id"`
	TelegramID string    `json:"telegramId"`
	Prompt     string    `json:"prompt"`
	Response   string    `json:"response,omitempty"`
	Category   string    `json:"category"`
	Model      string    `json:"model,omitempty"`
	SessionID  string    `json:"sessionId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type saveRequest struct {
	TelegramID  string `json:"telegramId"`
	InitDataRaw string `json:"initDataRaw"`
	Prompt      string `json:"prompt"`
	Response   string `json:"response,omitempty"`
	Category   string `json:"category"`
	Model      string `json:"model,omitempty"`
	SessionID  string `json:"sessionId,omitempty"`
}

type listResponse struct {
	Items []Item `json:"items"`
}

type saveResponse struct {
	Item Item `json:"item"`
}
