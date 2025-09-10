package presenter

type handler func(error) (string, bool)

type PresentedError struct {
	UserMessage string
}

func (p PresentedError) Error() string {
	return p.UserMessage
}
