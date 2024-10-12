package adapters

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/Gregmus2/sync-service/internal/common"
)

func NewFirebaseClient(firebaseApp *firebase.App) (*auth.Client, error) {
	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, err
	}

	return authClient, nil
}

func NewFirebaseApp(config *common.Config) (*firebase.App, error) {
	fbConfig := &firebase.Config{
		ProjectID: config.FirebaseProjectID,
	}

	return firebase.NewApp(context.Background(), fbConfig)
}
