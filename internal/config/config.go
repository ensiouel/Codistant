package config

import (
	"log/slog"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Bot      Bot
	Ollama   Ollama
	Postgres Postgres
}

func (conf Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Group("bot",
			slog.String("token", "<hidden>")),
		slog.Group("postgres",
			slog.String("user", conf.Postgres.User),
			slog.String("password", "<hidden>"),
			slog.String("db", conf.Postgres.DB),
		),
		slog.Group("ollama",
			slog.String("host", conf.Ollama.Host),
			slog.String("models", strings.Join(conf.Ollama.Models, ",")),
		),
	)
}

type Bot struct {
	Token string `env:"BOT_TOKEN" env-required:"true"`
}

type Ollama struct {
	Host   string   `env:"OLLAMA_HOST" env-required:"true"`
	Models []string `env:"OLLAMA_MODELS" env-required:"true"`
}

type Postgres struct {
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB" env-required:"true"`
}

func Read() (Config, error) {
	var conf Config
	err := cleanenv.ReadEnv(&conf)
	if err != nil {
		return Config{}, err
	}
	return conf, nil
}
