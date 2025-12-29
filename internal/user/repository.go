//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks
package user

import (
	"context"

	"bdayBot/internal/user/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateOrUpdate(ctx context.Context, user *models.User) error

	FindByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)

	// Group operations now use telegramID
	AddGroup(ctx context.Context, telegramID int64, groupID int64) error

	RemoveGroup(ctx context.Context, telegramID int64, groupID int64) error

	GetGroupsForUser(ctx context.Context, telegramID int64) ([]int64, error)

	GetTodaysBirthdays(ctx context.Context, day, month int) ([]*BirthdayDelivery, error)

	DeactivateUser(ctx context.Context, telegramID int64) error

	ReactivateUser(ctx context.Context, telegramID int64) error

	GetBirthdaysInGroup(ctx context.Context, groupID int64) ([]BirthdayInGroup, error)
}

type BirthdayDelivery struct {
	UserID       uuid.UUID `bun:"id"`
	TelegramID   int64     `bun:"telegram_id"`
	Username     *string   `bun:"username"`
	FirstName    string    `bun:"first_name"`
	LastName     *string   `bun:"last_name"`
	BirthdayYear *int      `bun:"birthday_year"`
	GroupID      int64     `bun:"group_id"`
}

// DTO for the result
type BirthdayInGroup struct {
	TelegramID    int64
	Username      *string
	FirstName     string
	LastName      *string
	BirthdayDay   int
	BirthdayMonth int
	BirthdayYear  *int
}
