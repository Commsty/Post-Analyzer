package openrouter

type openRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Reasoning   Reasoning `json:"reasoning"`
	Verbosity   string    `json:"verbosity"`
	Temperature float32   `json:"temperature"`
	ToPP        float32   `json:"top_p"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Reasoning struct {
	Enabled bool `json:"enabled"`
}

type openRouterResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}
