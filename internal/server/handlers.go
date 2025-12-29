package server

import (
	"bdayBot/internal/scheduler"
	"bdayBot/internal/user/handler/http"
	"bdayBot/internal/user/models"
	"bdayBot/internal/user/repository"
	"bdayBot/internal/user/usecase"

	"gopkg.in/telebot.v4/middleware"

	"context"
	"gopkg.in/telebot.v4"
)

func (s *Server) MapHandlers() {
	repo := repository.NewUserRepository(s.db, *s.cfg)
	uc := usecase.NewUserUsecase(repo, s.cfg, s.logger)
	handler := handler.NewUserHandler(uc, s.logger, s.cfg)

	s.bot.Handle("/birthday", handler.SetBirthday)
	s.bot.Handle("/disable", handler.OptOutFromGroup)
	s.bot.Handle("/enable", handler.OptInToGroup)
	s.bot.Handle("/pause", handler.DeactivateUser)
	s.bot.Handle("/resume", handler.ReactivateUser)
	s.bot.Handle("/ping", handler.CheckAlive)
	s.bot.Handle("/birthdays", handler.HandleListBirthdays)
	s.bot.Handle("/help", handler.Help)
	s.bot.Handle("/version", handler.CheckVersion)

	s.bot.Use(middleware.Logger())

	ctx := context.Background()
	if _, err := s.db.NewCreateTable().Model((*models.User)(nil)).IfNotExists().Exec(ctx); err != nil {
		s.logger.Error(err)
	}
	if _, err := s.db.NewCreateTable().Model((*models.RegisteredGroup)(nil)).IfNotExists().Exec(ctx); err != nil {
		s.logger.Error(err)
	}

	scheduler := scheduler.NewBirthdayScheduler(uc, s.bot, *s.logger, s.location)
	scheduler.Start(ctx)
	s.bot.Handle("/testrun", func(ctx telebot.Context) error {
		if ctx.Sender().ID == int64(398116518) || ctx.Sender().ID == int64(6659775172) {
			scheduler.RunOnce(context.Background())
			ctx.Send("sending test wishes")
		}
		return nil
	})
}
