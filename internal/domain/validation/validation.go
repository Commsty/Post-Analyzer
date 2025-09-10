package validation

import (
	"context"
	"errors"
	"fmt"
	"post-analyzer/internal/adapters/telegram/user"
	"post-analyzer/internal/domain/entity"
	"regexp"
	"strings"
)

var (
	ErrArgNumber = errors.New("invalid number of arguments")

	ErrTimeFormat = errors.New("invalid time format")
	ErrTimeValue  = errors.New("invalid time value")

	ErrShortUsername    = errors.New("channel username too short")
	ErrCharactersInName = errors.New("forrbidden characters in username")

	ErrChannelNotFound = errors.New("channel @")
	ErrExternal        = errors.New("external service error")
)

type Validator func(ctx context.Context, command string, sub *entity.Subscription) error

func ArgsValidator(next Validator) Validator {
	return func(ctx context.Context, command string, sub *entity.Subscription) error {

		command = strings.TrimSpace(command)
		if len(strings.Split(command, " ")) != 2 {
			return ErrArgNumber
		}

		if next != nil {
			return next(ctx, command, sub)
		}
		return nil
	}
}

func TimeValidator(next Validator) Validator {

	return func(ctx context.Context, command string, sub *entity.Subscription) error {

		time := strings.Split(command, " ")[1]

		var hour, minute int
		if _, err := fmt.Sscanf(time, "%d:%d", &hour, &minute); err != nil {
			return ErrTimeFormat
		}

		if (hour < 0 || hour > 23) || (minute < 0 || minute > 59) {
			return ErrTimeValue
		}

		sub.SendingTime = time

		if next != nil {
			return next(ctx, command, sub)
		}
		return nil
	}
}

func ChannelNameValidator(next Validator) Validator {

	return func(ctx context.Context, command string, sub *entity.Subscription) error {

		name := strings.Split(command, " ")[0]

		name = strings.TrimPrefix(name, "https://")
		name = strings.TrimPrefix(name, "http://")
		name = strings.TrimPrefix(name, "t.me/")
		name = strings.TrimPrefix(name, "telegram.me/")
		name = strings.TrimPrefix(name, "@")
		name = strings.Split(name, "?")[0]
		name = strings.Split(name, "/")[0]

		if len(name) < 5 {
			return ErrShortUsername
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
			return ErrCharactersInName
		}

		sub.ChannelUsername = name

		if next != nil {
			return next(ctx, command, sub)
		}
		return nil
	}
}

func ChannelValidator(next Validator, client user.TelegramService) Validator {

	return func(ctx context.Context, command string, sub *entity.Subscription) error {

		channel, err := client.ChannelInfo(ctx, sub.ChannelUsername)
		if err != nil {
			if errors.Is(err, user.ErrChannelNotFound) {
				return fmt.Errorf("%w%s: %s", ErrChannelNotFound, sub.ChannelUsername, err)
			}
			return ErrExternal
		}

		sub.ChannelID = channel.ID

		if next != nil {
			return next(ctx, command, sub)
		}
		return nil
	}

}
