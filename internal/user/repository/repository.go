package repository

import (
	"bdayBot/config"
	"bdayBot/internal/user"
	"bdayBot/internal/user/models"
	appErrors "bdayBot/pkg/errors"

	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	db     *bun.DB
	config config.Config
}

func NewUserRepository(db *bun.DB, cfg config.Config) user.UserRepository {
	return UserRepository{db: db, config: cfg}
}

func (r UserRepository) CreateOrUpdate(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().
		Model(user).
		On("CONFLICT (telegram_id) DO UPDATE").
		Set("username = EXCLUDED.username").
		Set("first_name = EXCLUDED.first_name").
		Set("last_name = EXCLUDED.last_name").
		Set("birthday_day = EXCLUDED.birthday_day").
		Set("birthday_month = EXCLUDED.birthday_month").
		Set("birthday_year = EXCLUDED.birthday_year").
		Set("timezone = EXCLUDED.timezone").
		Set("is_active = EXCLUDED.is_active").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}
	return nil
}

func (r UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().
		Model(user).
		Where("telegram_id = ?", telegramID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by telegram id: %w", err)
	}
	return user, nil
}

func (r UserRepository) AddGroup(
	ctx context.Context,
	telegramID int64,
	groupID int64,
) error {

	var userID uuid.UUID

	err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Column("id").
		Where("telegram_id = ?", telegramID).
		Scan(ctx, &userID)

	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	_, err = r.db.NewInsert().
		Model(&models.RegisteredGroup{
			UserID:  userID,
			GroupID: groupID,
		}).
		On("CONFLICT (user_id, group_id) DO NOTHING").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add group: %w", err)
	}

	return nil
}

func (r UserRepository) RemoveGroup(
	ctx context.Context,
	telegramID int64,
	groupID int64,
) error {

	var userID uuid.UUID

	err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Column("id").
		Where("telegram_id = ?", telegramID).
		Scan(ctx, &userID)

	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	_, err = r.db.NewDelete().
		Model((*models.RegisteredGroup)(nil)).
		Where("user_id = ?", userID).
		Where("group_id = ?", groupID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to remove group: %w", err)
	}

	return nil
}

func (r UserRepository) GetGroupsForUser(ctx context.Context, telegramID int64) ([]int64, error) {
	var groupIDs []int64
	// SELECT ug.group_id
	// FROM user_groups AS ug
	// INNER JOIN users AS u ON u.id = ug.user_id
	// WHERE u.telegram_id = $1;

	err := r.db.NewSelect().
		ColumnExpr("ug.group_id").
		TableExpr("registered_groups ug").
		Join("INNER JOIN users u ON u.id = ug.user_id").
		Where("u.telegram_id = ?", telegramID).
		Scan(ctx, &groupIDs)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int64{}, nil
		}
		return nil, fmt.Errorf("failed to get groups for user: %w", err)
	}
	return groupIDs, nil
}

func (r UserRepository) GetBirthdaysInGroup(ctx context.Context, groupID int64) ([]user.BirthdayInGroup, error) {
	var results []user.BirthdayInGroup

	err := r.db.NewSelect().
		Column("u.telegram_id", "u.username", "u.first_name", "u.last_name",
			"u.birthday_day", "u.birthday_month", "u.birthday_year").
		TableExpr("users AS u").
		Join("INNER JOIN registered_groups rg ON rg.user_id = u.id").
		Where("rg.group_id = ? AND u.is_active = true", groupID).
		OrderExpr("u.birthday_month ASC, u.birthday_day ASC").
		Scan(ctx, &results)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.BirthdayInGroup{}, nil
		}
		return nil, fmt.Errorf("failed to get birthdays in group: %w", err)
	}

	return results, nil
}

func (r UserRepository) GetTodaysBirthdays(ctx context.Context, day, month int) ([]*user.BirthdayDelivery, error) {
	var deliveries []*user.BirthdayDelivery

	// SELECT
	//     users.id,
	//     users.telegram_id,
	//     users.username,
	//     users.first_name,
	//     users.last_name,
	//     users.birthday_year,
	//     ug.group_id
	// FROM users
	// INNER JOIN user_groups AS ug ON ug.user_id = users.id
	// WHERE
	//     users.birthday_day = $1
	//     AND users.birthday_month = $2
	//     AND users.is_active = true;

	err := r.db.NewSelect().
		Column("users.id").
		Column("users.telegram_id").
		Column("users.username").
		Column("users.first_name").
		Column("users.last_name").
		Column("users.birthday_year").
		ColumnExpr("ug.group_id").
		Table("users").
		Join("INNER JOIN registered_groups ug ON ug.user_id = users.id").
		Where("users.birthday_day = ? AND users.birthday_month = ? AND users.is_active = true", day, month).
		Scan(ctx, &deliveries)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get today's birthdays: %w", err)
	}
	return deliveries, nil
}

func (r UserRepository) DeactivateUser(ctx context.Context, telegramID int64) error {
	res, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = false").
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("telegram_id = ?", telegramID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

func (r UserRepository) ReactivateUser(ctx context.Context, telegramID int64) error {
	res, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = true").
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("telegram_id = ?", telegramID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}
