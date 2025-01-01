package bot

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"codistant/internal/assistant"
	"codistant/internal/config"
	assistantpkg "codistant/pkg/assistant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	ollamaapi "github.com/ollama/ollama/api"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

var limiter = rate.NewLimiter(rate.Every(1*time.Second), 1)

type Bot struct {
	API              *telebot.Bot
	AssistantService assistantpkg.Service
	RiverClient      *river.Client[pgx.Tx]
	ollama           *ollamaapi.Client
	conf             config.Config
}

func NewBot(conf config.Config) *Bot {
	api, err := telebot.NewBot(telebot.Settings{
		Token:  conf.Bot.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	ollama := ollamaapi.NewClient(&url.URL{
		Scheme: "http",
		Host:   conf.Ollama.Host,
	}, http.DefaultClient)

	assistantService := assistant.NewService(ollama)
	err = assistantService.Pull(context.Background(), conf.Ollama.Models)
	if err != nil {
		panic(err)
	}

	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@postgres:5432/codistant?sslmode=disable")
	if err != nil {
		panic(err)
	}

	dbDriver := riverpgxv5.New(db)
	{
		migrator, err := rivermigrate.New(dbDriver, nil)
		if err != nil {
			panic(err)
		}
		for _, version := range migrator.AllVersions() {
			_, err = migrator.Migrate(
				context.Background(),
				rivermigrate.DirectionUp,
				&rivermigrate.MigrateOpts{
					TargetVersion: version.Version,
				},
			)
			if err != nil {
				panic(err)
			}
		}
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, &ChatWorker{
		BotAPI:           api,
		AssistantService: assistantService,
	})

	riverClient, err := river.NewClient(
		dbDriver,
		&river.Config{
			Queues: map[string]river.QueueConfig{
				river.QueueDefault: {
					MaxWorkers: 3,
				},
			},
			Workers: workers,
		},
	)
	if err != nil {
		panic(err)
	}

	return &Bot{
		API:              api,
		AssistantService: assistantService,
		RiverClient:      riverClient,
		ollama:           ollama,
		conf:             conf,
	}
}

func (bot *Bot) Start(ctx context.Context) {
	go func() {
		slog.Info("River client successfully starter")
		err := bot.RiverClient.Start(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	bot.API.Use(
		middleware.AutoRespond(),
	)

	bot.API.Handle(telebot.OnText, func(c telebot.Context) error {
		job, err := bot.RiverClient.Insert(ctx, ChatArgs{
			ChatID:  c.Chat().ID,
			Content: c.Text(),
		}, &river.InsertOpts{
			MaxAttempts: 1,
		})
		if err != nil {
			return err
		}

		if job.UniqueSkippedAsDuplicate {
			return c.Reply("Дождитесь ответа на предыдущее сообщение.")
		}

		return nil
	})

	bot.API.Handle("/clear", func(c telebot.Context) error {
		return nil
	})

	slog.Info("Telegram bot successfully starter")

	bot.API.Start()
}

func (bot *Bot) Stop() {
	bot.API.Stop()
}
