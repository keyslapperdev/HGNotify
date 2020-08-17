package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/chat/v1"
)

//Grabbing config information for contact with Hangouts Chat's api.
var serviceKeyPath = os.Getenv("CHAT_SERVICE_KEY_PATH")

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
