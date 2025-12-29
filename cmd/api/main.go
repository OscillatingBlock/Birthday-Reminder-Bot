package main

import (
	"bdayBot/config"
	"bdayBot/internal/server"
	"bdayBot/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"gopkg.in/telebot.v4"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	// "github.com/joho/godotenv"
)

func main() {
	// err := godotenv.Load("/etc/bdayBot/bdayBot.env")
	// if err != nil {
	// 	slog.Error("Error loading .env file")
	// }

	cfgPath := os.Getenv("BDAYBOT_CONFIG")
	if cfgPath == "" {
		log.Fatal("BDAYBOT_CONFIG is not set")
	}
	fmt.Printf("BDAYBOT_CONFIG=%s", cfgPath)
	RawConfig, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("error while loading config", "err", err)
	}
	cfg, err := config.ParseConfig(RawConfig)
	if err != nil {
		slog.Error("error while parsing config", "err", err)
	}
	logger, err := logger.NewLogger(cfg)
	pref := telebot.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		logger.Errorf("failed to get bot: %w", err)
	}

	dsn := os.Getenv("DSN")
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	if err := db.PingContext(context.Background()); err != nil {
		slog.Error("❌ failed to connect to Postgres: %v", "err", err)
	}

	slog.Info("✅ Connected to Postgres successfully")

	defer db.Close()

	location, err := time.LoadLocation(cfg.TimeZone.TimeZone)
	if err != nil {
		logger.Warnf("failed to load timezone '%s', using Asia/Kolkata: %v", cfg.TimeZone.TimeZone, err)
		location, _ = time.LoadLocation("Asia/Kolkata")
	}
	logger.Infof("location set as: %v", location.String())
	server := server.NewServer(bot, logger, cfg, db, location)
	server.Run()
}
