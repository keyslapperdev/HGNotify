package main

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestGroupCreate(t *testing.T) {
	_ = spew.Dump

	Logger.Active(false)
	Groups := make(GroupList)

	selfName := randString(5)
	msgObj := messageResponse{
		Message: message{
			Sender: User{
				Name: selfName,
				GID:  genGID(0),
				Type: "HUMAN",
			},

			Mentions: nil,
		},
		Room: space{
			GID:  "spaces/a",
			Type: "ROOM",
		},
		Time: "",
	}

	t.Run("Created new empty group", func(t *testing.T) {
		newGroupName := randString(10)

		Groups.Create(newGroupName, "", msgObj)
		saveName := strings.ToLower(newGroupName)

		group, exists := Groups[saveName]
		if !exists {
			t.Fatal("New group wasn't created")
		}

		if group.Members != nil {
			t.Fatal("New group isn't empty")
		}
	})

	wantedMentions := []annotation{
		{
			Called: userMention{
				User: User{
					Name: "User 1 Name",
					GID:  genGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		},
		{
			Called: userMention{
				User: User{
					Name: "User 2 Name",
					GID:  genGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		},
	}

	msgObj.Message.Mentions = wantedMentions

	unwantedBotName := randString(10)
	msgObj.Message.Mentions = append(msgObj.Message.Mentions, annotation{
		Called: userMention{
			User: User{
				Name: unwantedBotName,
				GID:  genGID(0),
				Type: "BOT",
			},
		},
		Type: "USER_MENTION",
	})

	t.Run("Creating Group with multiple members", func(t *testing.T) {
		newGroupName := randString(10)

		Groups.Create(newGroupName, "", msgObj)
		saveName := strings.ToLower(newGroupName)

		group := Groups[saveName]

		if len(group.Members) != len(wantedMentions) {
			t.Fatalf("Incorrect amount of members added\nWanted: %d\nGot: %d\n",
				len(wantedMentions),
				len(group.Members),
			)
		}

		t.Run("Ignore bot in mention", func(t *testing.T) {
			for _, member := range group.Members {
				if member.Name == unwantedBotName {
					t.Fatal("Bot was found in group")
				}
			}
		})
	})

	t.Run("Creating group with self included", func(t *testing.T) {
		newGroupName := randString(10)

		Groups.Create(newGroupName, "self", msgObj)
		saveName := strings.ToLower(newGroupName)

		group := Groups[saveName]

		var foundSelf bool
		for _, member := range group.Members {
			if member.Name == selfName {
				foundSelf = true
			}
		}

		if !foundSelf {
			t.Fatal("Sender was not added in new group")
		}
	})
}

/*
type messageResponse struct {
    Message  message `json:"message"`
    Room     space `json:"space"`
    Time     string `json:"eventTime"`
    IsMaster bool
}

type message struct {
    Sender User `json:"sender"`

    Mentions []annotation `json:"annotations"`

    Text string `json:"text"`
}

type annotation struct {
    Called userMention `json:"userMention"`
    Type string `json:"type"`
}

type userMention struct {
    User `json:"user"`
}

type space struct {
    GID  string `json:"name"`
    Type string `json:"type"`
}
*/

//Test helper data
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randString(length int) string {
	return StringWithCharset(length, charset)
}

func genGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "users/" + StringWithCharset(length, "0123456789")
}
