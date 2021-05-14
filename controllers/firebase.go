package controllers

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	auth "firebase.google.com/go/v4/auth"
)

var AuthClient *auth.Client

func InitAuth() {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Auth(context.Background())
	fmt.Println(client)
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
