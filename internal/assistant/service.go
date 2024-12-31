package assistant

import (
	"context"
	"strings"

	assistantpkg "codistant/pkg/assistant"
	"codistant/pkg/optional"
	ollamaapi "github.com/ollama/ollama/api"
)

type Service struct {
	ollama   *ollamaapi.Client
	contexts map[int64]*assistantpkg.Context
}

func NewService(ollama *ollamaapi.Client) *Service {
	return &Service{
		ollama:   ollama,
		contexts: make(map[int64]*assistantpkg.Context),
	}
}

func (svc *Service) Context(ctx context.Context, chatID int64) (*assistantpkg.Context, error) {
	c, ok := svc.contexts[chatID]
	if !ok {
		c = &assistantpkg.Context{
			ChatID:   chatID,
			Model:    "llama3.2:1b",
			Messages: make([]assistantpkg.Message, 0),
		}
		svc.contexts[chatID] = c
	}
	return c, nil
}

func (svc *Service) Chat(ctx context.Context, chatID int64, content string, fn assistantpkg.ChatResponseFunc) error {
	c, err := svc.Context(ctx, chatID)
	if err != nil {
		return err
	}

	c.AddMessage(assistantpkg.RoleUser, content)

	builder := &strings.Builder{}
	builder.Grow(1024)

	err = svc.ollama.Chat(ctx, &ollamaapi.ChatRequest{
		Model:    c.Model,
		Messages: messagesToOllamaMessages(c.Messages),
		Stream:   optional.Pointer(false),
	}, func(response ollamaapi.ChatResponse) error {
		builder.WriteString(response.Message.Content)
		return fn(assistantpkg.ChatResponse{
			Content: response.Message.Content,
			Done:    response.Done,
		})
	})
	if err != nil {
		return err
	}

	c.AddMessage(assistantpkg.RoleAssistant, builder.String())

	return nil
}

func messagesToOllamaMessages(messages []assistantpkg.Message) []ollamaapi.Message {
	result := make([]ollamaapi.Message, len(messages))
	for i, message := range messages {
		result[i] = ollamaapi.Message{
			Role:    message.Role,
			Content: message.Content,
		}
	}
	return result
}
