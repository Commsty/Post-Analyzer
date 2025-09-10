package presenter

var handlers []handler

func register(h handler) {
	handlers = append(handlers, h)
}

func init() {
	register(validationHandler)
	register(repositoryHandler)
}

func PresentError(e error) *PresentedError {

	for _, h := range handlers {
		if msg, ok := h(e); ok {
			return &PresentedError{
				UserMessage: msg,
			}
		}
	}

	return &PresentedError{
		UserMessage: "Внутрення ошибка сервиса. Пожалуйста, повторите запрос позже.",
	}
}
