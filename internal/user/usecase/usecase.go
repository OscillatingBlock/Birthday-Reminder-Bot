package usecase

import (
	"bdayBot/config"
	"bdayBot/internal/user"
	"bdayBot/internal/user/models"
	appErrors "bdayBot/pkg/errors"
	"bdayBot/pkg/logger"

	"context"
	"errors"
	"fmt"
	"time"
)

type UserUsecase struct {
	repo   user.UserRepository
	config *config.Config
	logger *logger.Logger
}

func NewUserUsecase(repo user.UserRepository, config *config.Config, logger *logger.Logger) user.UserUsecase {
	return &UserUsecase{
		repo:   repo,
		config: config,
		logger: logger,
	}
}

func (uc *UserUsecase) RegisterOrUpdateBirthday(ctx context.Context, params user.RegisterBirthdayParams) error {
	if params.Timezone != nil && !IsValidTimezone(*params.Timezone) {
		return appErrors.ErrInvalidTimezone
	}
	user := &models.User{
		TelegramID:    params.TelegramID,
		Username:      params.Username,
		FirstName:     params.FirstName,
		LastName:      params.LastName,
		BirthdayDay:   params.BirthdayDay,
		BirthdayMonth: params.BirthdayMonth,
		BirthdayYear:  params.BirthdayYear,
		Timezone:      params.Timezone,
	}

	err := uc.repo.CreateOrUpdate(ctx, user)
	if err != nil {
		uc.logger.Errorf("failed to create or update user: %v", err)
		return appErrors.ErrDatabaseFailed(err)
	}

	if params.GroupID < 0 {
		if err := uc.repo.AddGroup(ctx, params.TelegramID, params.GroupID); err != nil {
			uc.logger.Errorf("failed to register group: %v", err)
			// Non-critical — don't fail
		}
	}
	return nil
}

func IsValidTimezone(name string) bool {
	if name == "" {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}

func (uc *UserUsecase) GetTodaysBirthdayDeliveries(ctx context.Context, day, month int) ([]*user.BirthdayDelivery, error) {
	bdays, err := uc.repo.GetTodaysBirthdays(ctx, day, month)
	if err != nil {
		uc.logger.Errorf("failed to get bdays: %w", err)
		return nil, appErrors.Internal("failed to get bdays")
	}
	return bdays, nil
}

func (uc *UserUsecase) OptOutFromGroup(ctx context.Context, telegramID int64, groupID int64) error {
	err := uc.repo.RemoveGroup(ctx, telegramID, groupID)
	if err != nil {
		uc.logger.Errorf("failed to remvoe group: %w", err)
		return appErrors.Internal("failed to opt out from group")
	}
	return nil
}

func (uc *UserUsecase) OptInToGroup(ctx context.Context, telegramID int64, groupID int64) error {
	err := uc.repo.AddGroup(ctx, telegramID, groupID)
	if err != nil {
		uc.logger.Errorf("failed to add group: %w", err)
		return appErrors.Internal("failed to opt in group")
	}
	return nil
}

func (uc *UserUsecase) DeactivateUser(ctx context.Context, telegramID int64) error {
	if err := uc.repo.DeactivateUser(ctx, telegramID); err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return appErrors.ErrUserNotFound
		}
		uc.logger.Errorf("failed to deactivate user %d: %v", telegramID, err)
		return appErrors.Internal("failed to deactivate user")
	}
	return nil
}

func (uc *UserUsecase) ReactivateUser(ctx context.Context, telegramID int64) error {
	if err := uc.repo.ReactivateUser(ctx, telegramID); err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return appErrors.ErrUserNotFound
		}
		uc.logger.Errorf("failed to reactivate user %d: %v", telegramID, err)
		return appErrors.Internal("failed to reactivate user")
	}
	return nil
}

func (uc *UserUsecase) ListBirthdaysInGroup(ctx context.Context, groupID int64) ([]user.BirthdayInfo, error) {
	list, err := uc.repo.GetBirthdaysInGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var result []user.BirthdayInfo
	for _, b := range list {
		// Format name
		name := b.FirstName
		if b.LastName != nil && *b.LastName != "" {
			name += " " + *b.LastName
		}
		if b.Username != nil && *b.Username != "" {
			name += " (" + *b.Username + ")"
		}

		// Format date
		monthName := time.Month(b.BirthdayMonth).String()
		dateStr := fmt.Sprintf("%d %s", b.BirthdayDay, monthName)
		if b.BirthdayYear != nil {
			dateStr += fmt.Sprintf(" %d", *b.BirthdayYear)
		}

		// Mention (for clickable name if username exists)
		mention := name
		if b.Username != nil && *b.Username != "" {
			mention = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", b.TelegramID, name)
		}

		result = append(result, user.BirthdayInfo{
			Name:         name,
			BirthdayDate: dateStr,
			Mention:      mention,
		})
	}

	return result, nil
}
