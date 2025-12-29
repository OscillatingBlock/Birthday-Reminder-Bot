package user

import (
	"context"
)

type UserUsecase interface {
	RegisterOrUpdateBirthday(ctx context.Context, params RegisterBirthdayParams) error

	GetTodaysBirthdayDeliveries(ctx context.Context, date, month int) ([]*BirthdayDelivery, error)

	OptOutFromGroup(ctx context.Context, telegramID int64, groupID int64) error

	OptInToGroup(ctx context.Context, telegramID int64, groupID int64) error

	DeactivateUser(ctx context.Context, telegramID int64) error

	ReactivateUser(ctx context.Context, telegramID int64) error

	ListBirthdaysInGroup(ctx context.Context, groupID int64) ([]BirthdayInfo, error)
}

type RegisterBirthdayParams struct {
	TelegramID    int64   `validate:"required"`
	Username      *string `validate:"omitempty"`
	FirstName     string  `validate:"required"`
	LastName      *string `validate:"omitempty"`
	BirthdayDay   int     `validate:"required,min=1,max=31"`
	BirthdayMonth int     `validate:"required,min=1,max=12"`
	BirthdayYear  *int    `validate:"omitempty,min=1900,max=9999"`
	Timezone      *string `validate:"omitempty"`
	GroupID       int64   `validate:"required"`
}
type BirthdayInfo struct {
	Name         string
	BirthdayDate string
	Mention      string
}
