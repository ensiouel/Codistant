package assistant

type Role = string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

type Message struct {
	ID      int    `json:"id"`
	ChatID  int64  `json:"chat_id"`
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Context struct {
	ChatID   int64     `json:"chat_id"`
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

func (c *Context) AddMessage(role Role, content string) {
	message := Message{
		ID:      len(c.Messages) + 1,
		ChatID:  c.ChatID,
		Role:    role,
		Content: content,
	}
	c.Messages = append(c.Messages, message)
}
