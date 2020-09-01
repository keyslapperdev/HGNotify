package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	chat "google.golang.org/api/chat/v1"
)

// Grabbing config information for contact with Hangouts Chat's api.
var serviceKeyPath = os.Getenv("CHAT_SERVICE_KEY_PATH")

// The first time the token is receieved and validated, it'll be set to this
// variable. So as to not make a request every time someone makes uses the
// bot. If I find out that more than one token is used ( which I don't think
// there should be ), I'll make this into an arrally
var cachedAuthToken string

func getChatService(client *http.Client) *chat.Service {
	service, err := chat.New(client)
	if err != nil {
		log.Fatal("Error creating chat service " + err.Error())
	}

	return service
}

func getChatClient() *http.Client {
	ctx := context.Background()

	data, err := ioutil.ReadFile(serviceKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	creds, err := google.CredentialsFromJSON(
		ctx,
		data,
		"https://www.googleapis.com/auth/chat.bot",
	)
	if err != nil {
		log.Fatal(err)
	}

	return oauth2.NewClient(ctx, creds.TokenSource)
}

func isValidRequest(authToken string) bool {
	if authToken == cachedAuthToken {
		return true
	}

	verified, err := validateJWT(authToken)
	if err != nil {
		log.Println("Error validating auth token: ", err.Error())
		return false
	}

	if verified {
		cachedAuthToken = authToken
		return true
	}

	return false
}
