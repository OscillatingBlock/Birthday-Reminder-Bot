package server

import (
	"bdayBot/config"
	"bdayBot/pkg/logger"
	"time"

	"github.com/uptrace/bun"
	"gopkg.in/telebot.v4"
)

type Server struct {
	bot      *telebot.Bot
	logger   *logger.Logger
	cfg      *config.Config
	db       *bun.DB
	location *time.Location
}

func NewServer(bot *telebot.Bot, logger *logger.Logger, cfg *config.Config, db *bun.DB, location *time.Location) *Server {
	return &Server{bot: bot, logger: logger, cfg: cfg, db: db, location: location}
}

func (s *Server) Run() {
	s.MapHandlers()
	s.bot.Start()
}
