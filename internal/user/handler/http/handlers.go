package handler

import (
	"bdayBot/config"
	"bdayBot/internal/user"
	"bdayBot/pkg/logger"
	"bdayBot/pkg/utils"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

type UserHandler struct {
	usecase user.UserUsecase
	logger  *logger.Logger
	config  *config.Config
}

func NewUserHandler(usecase user.UserUsecase, logger *logger.Logger, config *config.Config) *UserHandler {
	return &UserHandler{
		usecase: usecase,
		logger:  logger,
		config:  config,
	}
}

func (h *UserHandler) SetBirthday(c telebot.Context) error {
	if c.Chat().Type != telebot.ChatPrivate {
		return c.Send("For privacy, please set your birthday in a private chat with me. Start a DM and use /birthday there.")
	}

	args := c.Args()
	if len(args) == 0 {
		return c.Send("Please provide your birthday.\nExamples:\n/birthday 21-12\n/birthday 21-12-1990")
	}

	input := strings.Join(args, " ")
	day, month, yearPtr, err := parseFlexibleDate(input)
	if err != nil {
		return c.Send("Invalid date format.\nUse: /birthday DD-MM or DD-MM-YYYY\nExample: /setbirthday 21-12")
	}

	var usernamePtr *string
	if c.Sender().Username != "" {
		usernamePtr = &c.Sender().Username
	}

	var lastNamePtr *string
	if c.Sender().LastName != "" {
		lastNamePtr = &c.Sender().LastName
	}

	params := user.RegisterBirthdayParams{
		TelegramID:    c.Sender().ID,
		Username:      usernamePtr,
		FirstName:     c.Sender().FirstName,
		LastName:      lastNamePtr,
		BirthdayDay:   day,
		BirthdayMonth: month,
		BirthdayYear:  yearPtr,
		GroupID:       c.Chat().ID,
	}

	if err := utils.ValidateStruct(context.Background(), params); err != nil {
		c.Send("Please provide your birthday.\nExamples:\n/birthday 21-12\n/birthday 21-12-1990")
	}

	if err := h.usecase.RegisterOrUpdateBirthday(context.Background(), params); err != nil {
		h.logger.Errorf("failed to register birthday for user %d: %v", c.Sender().ID, err)
		return c.Send("❌ Failed to save birthday. Please try again.")
	}

	yearStr := ""
	if yearPtr != nil {
		yearStr = fmt.Sprintf("-%.4d", *yearPtr)
	}

	return c.Send(fmt.Sprintf("✅ Birthday saved as %d %s%s!, use /enable in groups where you want to set reminder for your bday",
		day, time.Month(month).String(), yearStr))
}

func parseFlexibleDate(input string) (day, month int, year *int, err error) {
	input = strings.TrimSpace(input)

	parts := strings.Split(input, "-")
	if len(parts) != 2 && len(parts) != 3 {
		return 0, 0, nil, fmt.Errorf("invalid format")
	}

	day, err = strconv.Atoi(parts[0])
	if err != nil || day < 1 || day > 31 {
		return 0, 0, nil, fmt.Errorf("invalid day")
	}

	month, err = strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, nil, fmt.Errorf("invalid month")
	}

	if len(parts) == 3 {
		y, err := strconv.Atoi(parts[2])
		if err != nil || y < 1900 || y > time.Now().Year()+100 {
			return 0, 0, nil, fmt.Errorf("invalid year")
		}
		return day, month, &y, nil
	}

	return day, month, nil, nil
}

func (h *UserHandler) OptOutFromGroup(c telebot.Context) error {
	if c.Chat().Type == telebot.ChatPrivate {
		return c.Send("This command only works in groups.")
	}
	err := h.usecase.OptOutFromGroup(context.Background(), c.Sender().ID, c.Chat().ID)
	if err != nil {
		h.logger.Errorf("failed to opt out from group: %v", err)
	}
	c.Send("Sure, I'll not remind your birthday in this group")
	return nil
}

func (h *UserHandler) OptInToGroup(c telebot.Context) error {
	if c.Chat().Type == telebot.ChatPrivate {
		return c.Send("This command only works in groups.")
	}

	err := h.usecase.OptInToGroup(context.Background(), c.Sender().ID, c.Chat().ID)
	if err != nil {
		h.logger.Errorf("failed to opt in group: %v", err)
	}
	c.Send("Okay,I'll remind your birthday in this group")
	return nil
}

func (h *UserHandler) DeactivateUser(c telebot.Context) error {
	if c.Chat().Type == telebot.ChatPrivate {
		return c.Send("This command only works in groups.")
	}

	err := h.usecase.DeactivateUser(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("failed to deactivate user: %v", err)
	}
	c.Send(c.Send("All birthday reminders have been disabled. Use /resume to turn them back on."))
	return nil
}

func (h *UserHandler) ReactivateUser(c telebot.Context) error {
	err := h.usecase.ReactivateUser(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("failed to reactivate user: %v", err)
	}
	c.Send("Birthday reminders are now enabled in all your registered groups!")
	return nil
}

func (h *UserHandler) CheckAlive(c telebot.Context) error {

	c.Send("I'm alive madafaka!")
	return nil
}

func (h *UserHandler) CheckSystemTime(c telebot.Context) error {
	location, err := time.LoadLocation(h.config.TimeZone.TimeZone)
	if err != nil {
		location, _ = time.LoadLocation("Asia/Kolkata")
	}

	now := time.Now().In(location).Round(time.Second)
	return c.Send(fmt.Sprintf("Current system time: %s", now))
}

func (h *UserHandler) HandleListBirthdays(c telebot.Context) error {
	if c.Chat().Type == telebot.ChatPrivate {
		return c.Send("This command works only in groups.")
	}

	list, err := h.usecase.ListBirthdaysInGroup(context.Background(), c.Chat().ID)
	if err != nil {
		h.logger.Errorf("failed to list birthdays in group %d: %v", c.Chat().ID, err)
		return c.Send("Sorry, couldn't fetch the birthday list right now.")
	}

	if len(list) == 0 {
		return c.Send("No birthdays registered in this group yet.\nUse /setBirthday to add yours!")
	}

	var msg strings.Builder
	msg.WriteString("<b>Birthdays in this group:</b>\n\n")

	for _, b := range list {
		msg.WriteString(fmt.Sprintf("• %s — %s\n", b.Mention, b.BirthdayDate))
	}

	msg.WriteString("\n🎉 I'll remind everyone on their special day!")

	return c.Send(msg.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func (h *UserHandler) Help(c telebot.Context) error {
	help := `
	BIRTHDAY BOT COMMANDS🎂

SETUP & INFORMATION
/birthday DD-MM   : Register or update birth date (via DM)
/birthdays                : List registered birthdays in current group

NOTIFICATION CONTROL
/enable   : Activate reminders for current channel
/disable  : Deactivate reminders for current channel
/pause    : Suspend all automated reminders
/resume : Reinstate all automated reminders

GENERAL
/help : Display this menu

NOTES
Schedule: Notifications sent daily at 00:00 IST.
`

	c.Send(help, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
	return nil
}

func (h *UserHandler) CheckVersion(c telebot.Context) error {
	return c.Send("current version: v1.0.3")
}
