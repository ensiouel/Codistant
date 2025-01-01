package bot

import (
	"context"

	assistantpkg "codistant/pkg/assistant"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"gopkg.in/telebot.v4"
)

type ChatArgs struct {
	ChatID  int64  `json:"chat_id" river:"unique"`
	Content string `json:"prompt"`
}

func (ChatArgs) Kind() string { return "chat" }

func (args ChatArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 1,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}

type ChatWorker struct {
	BotAPI           telebot.API
	AssistantService assistantpkg.Service
	river.WorkerDefaults[ChatArgs]
}

func (w *ChatWorker) Work(ctx context.Context, job *river.Job[ChatArgs]) error {
	origin, err := w.BotAPI.Send(telebot.ChatID(job.Args.ChatID), "...")
	if err != nil {
		return err
	}

	manager := NewMessageManager(w.BotAPI, origin)
	return w.AssistantService.Chat(ctx, job.Args.ChatID, job.Args.Content, func(response assistantpkg.ChatResponse) error {
		return manager.Edit(response.Content)
	})
}
