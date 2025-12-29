package usecase

import (
	"bdayBot/config"
	"bdayBot/internal/user"
	"bdayBot/internal/user/mocks"
	appErrors "bdayBot/pkg/errors"
	"bdayBot/pkg/logger"
	"errors"

	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_RegisterOrUpdateBirthday(t *testing.T) {
	username := "aayushAgain"
	lastname := "sharma"
	year := 2006
	params := user.RegisterBirthdayParams{
		TelegramID:    int64(1),
		Username:      &username,
		FirstName:     "aayush",
		LastName:      &lastname,
		BirthdayDay:   27,
		BirthdayMonth: 5,
		BirthdayYear:  &year,
		Timezone:      nil,
		GroupID:       int64(69),
	}

	cfg := &config.Config{}
	logger, _ := logger.NewLogger(cfg)

	t.Run("happy path- user registererd", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockUserRepository(ctrl)

		uc := UserUsecase{
			repo:   mockRepo,
			config: cfg,
			logger: logger,
		}
		mockRepo.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any()).Return(nil)

		err := uc.RegisterOrUpdateBirthday(t.Context(), params)
		require.NoError(t, err)
	})

	t.Run("happy path- valid GroupID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockUserRepository(ctrl)

		uc := UserUsecase{
			repo:   mockRepo,
			config: cfg,
			logger: logger,
		}
		mockRepo.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		params.GroupID = -999
		err := uc.RegisterOrUpdateBirthday(t.Context(), params)
		require.NoError(t, err)
	})

	t.Run("sad path- db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockUserRepository(ctrl)

		uc := UserUsecase{
			repo:   mockRepo,
			config: cfg,
			logger: logger,
		}
		mockRepo.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to create user"))

		err := uc.RegisterOrUpdateBirthday(t.Context(), params)
		require.Error(t, err)
		var appErr *appErrors.AppError
		if !errors.As(err, &appErr) {
			t.Errorf("expected appError, got %v", err)
		}
		if appErr.Code != appErrors.CodeInternal {
			t.Errorf("expected code %s, got %s", appErrors.CodeInternal, appErr.Code)
		}
	})
}

func Test_GetTodaysBirthdays(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := logger.NewLogger(cfg)

	t.Run("happy path- no errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockUserRepository(ctrl)

		uc := UserUsecase{
			repo:   mockRepo,
			config: cfg,
			logger: logger,
		}

		var bdays []*user.BirthdayDelivery
		for i := range 4 {
			u := "testUser"
			year := 200 + i
			bday := user.BirthdayDelivery{
				UserID:       uuid.New(),
				TelegramID:   int64(i),
				Username:     &u,
				FirstName:    "orange",
				BirthdayYear: &year,
				GroupID:      int64(0 - i),
			}
			bdays = append(bdays, &bday)
		}

		mockRepo.EXPECT().GetTodaysBirthdays(gomock.Any(), 27, 5).Return(bdays, nil)

		gotBdays, err := uc.GetTodaysBirthdayDeliveries(t.Context(), 27, 5)
		require.NoError(t, err)
		assert.Equal(t, gotBdays, bdays)
	})

	t.Run("sad path- db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockUserRepository(ctrl)
		uc := UserUsecase{
			repo:   mockRepo,
			config: cfg,
			logger: logger,
		}

		mockRepo.EXPECT().GetTodaysBirthdays(gomock.Any(), 27, 5).Return(nil, fmt.Errorf("failed to start db"))
		gotBdays, err := uc.GetTodaysBirthdayDeliveries(t.Context(), 27, 5)
		require.Error(t, err)
		assert.Nil(t, gotBdays)

		var appErr *appErrors.AppError
		if !errors.As(err, &appErr) {
			t.Errorf("expected appError, got %v", err)
		}
		if appErr.Code != appErrors.CodeInternal {
			t.Errorf("expected code %s, got %s", appErrors.CodeInternal, appErr.Code)
		}
	})
}
