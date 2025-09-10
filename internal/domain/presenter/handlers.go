package presenter

import (
	"errors"
	"post-analyzer/internal/domain/validation"

	"github.com/jackc/pgx/v5/pgconn"
)

var validationHandler = func(e error) (string, bool) {

	switch {

	case errors.Is(e, validation.ErrArgNumber):
		return "В бот нужно передать 2 опции: канал и время отправки сообщения.", true

	case errors.Is(e, validation.ErrTimeFormat):
		return "Время отправки сообщения принимается в формате ЧЧ:ММ.", true

	case errors.Is(e, validation.ErrTimeValue):
		return "Некорректное время отправки сообщения.", true

	case errors.Is(e, validation.ErrShortUsername):
		return "Слишком короткий юзернейм для канала. Проверьте его и отправьте запрос заново.", true

	case errors.Is(e, validation.ErrCharactersInName):
		return "Юзернейм содержит некорректные символы. Проверьте его и отправьте запрос заново.", true

	case errors.Is(e, validation.ErrChannelNotFound):
		return "Не удалось найти канал с заданным юзернеймом. Проверьте его и повторите запрос.", true
	}

	return "", false
}

var repositoryHandler = func(e error) (string, bool) {

	var pgErr *pgconn.PgError
	if errors.As(e, &pgErr) && pgErr.Code == "23505" {
		return "Вы уже получаете уведомления о новостях из этого канала в данное время.", true
	}

	return "", false
}
