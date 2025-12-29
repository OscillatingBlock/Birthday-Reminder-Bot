package scheduler

import (
	"bdayBot/internal/user"
	"bdayBot/pkg/logger"
	"fmt"

	"context"
	"time"

	"gopkg.in/telebot.v4"
)

type BirthdayScheduler interface {
	Start(ctx context.Context)
	RunOnce(ctx context.Context) error
}

type birthdayScheduler struct {
	usecase  user.UserUsecase
	bot      *telebot.Bot
	logger   logger.Logger
	location *time.Location
}

func NewBirthdayScheduler(usecase user.UserUsecase, bot *telebot.Bot,
	logger logger.Logger, location *time.Location) BirthdayScheduler {

	return &birthdayScheduler{usecase: usecase, bot: bot, logger: logger, location: location}
}

func (s *birthdayScheduler) Start(parentCtx context.Context) {
	go func() {
		s.waitUntilNextMidnight()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-parentCtx.Done():
				s.logger.Info("birthday scheduler stopped")
				return
			case <-ticker.C:
				s.runJob(context.Background())
			}
		}
	}()
}

func (s *birthdayScheduler) RunOnce(ctx context.Context) error {
	return s.runJob(ctx)
}

func (s *birthdayScheduler) runJob(ctx context.Context) error {
	now := time.Now()
	bdays, err := s.usecase.GetTodaysBirthdayDeliveries(ctx, now.Day(),
		int(now.Month()))
	if err != nil {
		s.logger.Errorf("failed to get todays birthdays: %w", err)
		return err
	}
	if len(bdays) == 0 {
		s.logger.Infof("No birthdays today, date: %w", now.Format("02-01-2006"))
		return nil
	}
	s.logger.Infof("Sending %d birthday wishes for %s", len(bdays), now.Format("02-01-2006"))
	for i := range bdays {
		name := ""
		if bdays[i].Username != nil && *bdays[i].Username != "" {
			name = "@" + *bdays[i].Username
		} else if bdays[i].FirstName != "" {
			name = bdays[i].FirstName
		}
		bdayWish := fmt.Sprintf("🎉 Happy Birthday bitch, %s! 🎂", name)
		if _, err := s.bot.Send(&telebot.Chat{ID: bdays[i].GroupID}, bdayWish); err != nil {
			s.logger.Error("failed to send bday wish to user: %s, on date: %s, :%w",
				bdays[i].Username, now.Format("02-01-2006"), err)
		}
	}
	return nil
}

func (s *birthdayScheduler) waitUntilNextMidnight() {
	now := time.Now().In(s.location)
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location)
	if now.After(nextMidnight) {
		nextMidnight = nextMidnight.Add(24 * time.Hour)
	}

	duration := nextMidnight.Sub(now)
	s.logger.Infof("next birthday check at midnight in %v", duration.Round(time.Second))
	time.Sleep(duration)
}
