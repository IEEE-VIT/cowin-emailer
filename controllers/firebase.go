package controllers

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	auth "firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitAuth() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	creds := []byte(os.Getenv("JSON_CREDS"))
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON(creds))
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error connecting to auth client: %v\n", err)
	}

	AuthClient = client
}

func CheckUID(uid string) bool {
	_, err := AuthClient.GetUser(context.Background(), uid)

	if err != nil {
		return false
	}
	return true
}

func DeleteFirebaseUser(uid string) error {
	err := AuthClient.DeleteUser(context.Background(), uid)
	return err
}
