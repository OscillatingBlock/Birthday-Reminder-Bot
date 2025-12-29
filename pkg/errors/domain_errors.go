package errors

var (
	ErrUserNotFound         = NotFound("user not found")
	ErrInvalidBirthdayDate  = InvalidArg("invalid birthday date: day must be 1-31, month 1-12")
	ErrInvalidBirthdayDay   = InvalidArg("birthday day must be between 1 and 31")
	ErrInvalidBirthdayMonth = InvalidArg("birthday month must be between 1 and 12")
	ErrInvalidBirthdayYear  = InvalidArg("invalid birth year")
	ErrInvalidTimezone      = InvalidArg("invalid timezone name")
	ErrGroupNotRegistered   = NotFound("group not registered for birthday reminders")
	ErrUserDeactivated      = FailedPrecondition("user has deactivated birthday reminders")

	ErrNoBirthdaysToday = NotFound("no birthdays today")
)

func ErrDatabaseFailed(cause error) error {
	return Wrap(CodeInternal, "database operation failed", cause)
}

func ErrSendMessageFailed(cause error) error {
	return Wrap(CodeInternal, "failed to send birthday message", cause)
}

func ErrValidationFailed(cause error) error {
	return Wrap(CodeInvalidArgument, "request validation failed", cause)
}
