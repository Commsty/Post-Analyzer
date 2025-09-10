package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func (t telegramUserClient) authenticate(ctx context.Context) error {

	if ctx.Err() != nil {
		return ErrTimeLimit
	}

	authClient := t.createRawClient()

	return authClient.Run(ctx, func(ctx context.Context) error {

		if ctx.Err() != nil {
			return ErrTimeLimit
		}

		return authClient.Auth().IfNecessary(ctx, t.createAuthFlow())
	})
}

func (t telegramUserClient) createRawClient() *telegram.Client {
	return telegram.NewClient(
		t.appID,
		t.appHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{
				Path: t.sessionPath,
			},
		})
}

func (t telegramUserClient) createAuthFlow() auth.Flow {
	return auth.NewFlow(
		auth.Env("", auth.CodeAuthenticatorFunc(t.requestAuthCode)),
		auth.SendCodeOptions{},
	)
}

func (t telegramUserClient) requestAuthCode(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {

		fmt.Print("Enter code from TGSupport chat: ")
		var code string

		if _, err := fmt.Scanln(&code); err != nil {
			errChan <- err
			return
		}

		codeChan <- code
	}()

	select {
	case <-ctx.Done():
		return "", ErrTimeLimit
	case err := <-errChan:
		return "", err
	case code := <-codeChan:
		return code, nil
	}

}
