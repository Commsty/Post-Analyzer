package user

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type TelegramUserClient struct {
	appID       int
	appHash     string
	sessionPath string
}

func NewTelegramUserClient(appID int, appHash string, sessionPath string) (*TelegramUserClient, error) {

	if _, err := os.Stat(sessionPath); err != nil {
		directory := filepath.Dir(sessionPath)
		if err = os.MkdirAll(directory, 0700); err != nil {
			return nil, err
		}
	}

	client := &TelegramUserClient{
		appID:       appID,
		appHash:     appHash,
		sessionPath: sessionPath,
	}

	authCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.authenticate(authCtx); err != nil {
		return nil, fmt.Errorf("Authentication error: %w", err)
	}

	return client, nil
}

func (t *TelegramUserClient) authenticate(ctx context.Context) error {

	if ctx.Err() != nil {
		return fmt.Errorf("Authentication cancelled before start: %w", ctx.Err())
	}

	authClient := t.createRawClient(t.appID, t.appHash)

	return authClient.Run(ctx, func(ctx context.Context) error {

		if ctx.Err() != nil {
			return fmt.Errorf("Authentication cancelled before callback return: %w", ctx.Err())
		}

		return authClient.Auth().IfNecessary(ctx, t.createAuthFlow())
	})
}

func (t *TelegramUserClient) createRawClient(appID int, appHash string) *telegram.Client {
	return telegram.NewClient(
		appID,
		appHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{
				Path: t.sessionPath,
			},
		})
}

func (t *TelegramUserClient) createAuthFlow() auth.Flow {
	return auth.NewFlow(
		auth.Env("", auth.CodeAuthenticatorFunc(t.requestAuthCode)),
		auth.SendCodeOptions{},
	)
}

func (t *TelegramUserClient) requestAuthCode(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {

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
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case code := <-codeChan:
		return code, nil
	}

}
