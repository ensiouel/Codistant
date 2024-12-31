package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"codistant/internal/bot"
	"codistant/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf, err := config.Read()
	if err != nil {
		slog.Error("Failed to read config", slog.Any("error", err))
		return
	}

	codistant := bot.NewBot(conf)
	go codistant.Start(ctx)

	slog.Info("Codistant successfully started", slog.Any("config", conf))

	<-ctx.Done()

	slog.Info("Codistant gracefully shutdown")
}
