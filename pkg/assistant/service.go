package assistant

import (
	"context"
)

type ChatResponse struct {
	Content string
	Done    bool
}

type ChatResponseFunc func(ChatResponse) error

type Service interface {
	Chat(ctx context.Context, chatID int64, content string, fn ChatResponseFunc) error
}
