package main

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestCreateGroup(t *testing.T) {
	_ = spew.Dump

	Logger.Active(false)
	Groups := make(GroupList)

	selfName := randString(5)
	msgObj := messageResponse{
		Message: message{
			Sender: User{
				Name: selfName,
				GID:  genUserGID(0),
				Type: "HUMAN",
			},

			Mentions: nil,
		},
		Room: space{
			GID:  genRoomGID(0),
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
					GID:  genUserGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		},
		{
			Called: userMention{
				User: User{
					Name: "User 2 Name",
					GID:  genUserGID(0),
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
				GID:  genUserGID(0),
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

func TestDisbandGroup(t *testing.T) {
	_ = spew.Dump

	Logger.Active(false)
	Groups := make(GroupList)

	saveName := strings.ToLower(randString(10))

	msgObj := messageResponse{}
	msgObj.IsMaster = false

	t.Run("Disband group successsfully", func(t *testing.T) {
		group := new(Group)

		Groups[saveName] = group
		Groups.Disband(saveName, msgObj)

		if _, exist := Groups[saveName]; exist {
			t.Fatal("Group wasn't removed")
		}
	})

	t.Run("Disband didn't touch private group", func(t *testing.T) {
		group := new(Group)
		group.IsPrivate = true
		group.PrivacyRoomID = genRoomGID(0)

		Groups[saveName] = group

		Groups.Disband(saveName, msgObj)

		if _, exist := Groups[saveName]; !exist {
			t.Fatal("Private group was removed")
		}
	})

	t.Run("Doesn't die if group doesn't exist", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Panicked: %q", r)
			}
		}()

		Groups := new(GroupList)
		Groups.Disband(saveName, msgObj)
	})

}

func TestAddMembers(t *testing.T) {
	Logger.Active(false)

	Groups := make(GroupList)

	saveName := strings.ToLower(randString(10))

	Groups[saveName] = new(Group)
	group := Groups[saveName]

	selfName := randString(10)
	msgObj := messageResponse{
		Message: message{
			Sender: User{
				Name: selfName,
				GID:  genUserGID(0),
				Type: "HUMAN",
			},

			Mentions: nil,
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
		Time: "",
	}

	t.Run("Adds single member to group", func(t *testing.T) {
		wantedName := randString(10)
		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: wantedName,
					GID:  genUserGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		Groups.AddMembers(saveName, "", msgObj)

		var foundMember bool
		for _, member := range group.Members {
			if member.Name == wantedName {
				foundMember = true
			}
		}

		if !foundMember {
			t.Fatal("Member not added to the group")
		}
	})

	group.Members = nil

	t.Run("Adds multiple members to group", func(t *testing.T) {
		wantedMentions := []annotation{
			{
				Called: userMention{
					User: User{
						Name: "User 1 Name",
						GID:  genUserGID(0),
						Type: "HUMAN",
					},
				},
				Type: "USER_MENTION",
			},
			{
				Called: userMention{
					User: User{
						Name: "User 2 Name",
						GID:  genUserGID(0),
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
					GID:  genUserGID(0),
					Type: "BOT",
				},
			},
			Type: "USER_MENTION",
		})

		Groups.AddMembers(saveName, "", msgObj)

		if len(group.Members) != len(wantedMentions) {
			t.Fatalf("Correct embers not added to group\nWanted: %d\nGot: %d",
				len(group.Members),
				len(wantedMentions),
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

	group.Members = nil

	t.Run("Does not add same user twice", func(t *testing.T) {
		wantedName := randString(10)
		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: wantedName,
					GID:  genUserGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		Groups.AddMembers(saveName, "", msgObj)
		Groups.AddMembers(saveName, "", msgObj)

		if group.Members[0].Name != wantedName {
			t.Fatal("Correct memeber not added")
		}

		if len(group.Members) != 1 {
			t.Fatal("Incorrect number of members")
		}
	})

	group.Members = nil

	t.Run("Does not add member to private group", func(t *testing.T) {
		wantedName := randString(10)
		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: wantedName,
					GID:  genUserGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		group.IsPrivate = true
		group.PrivacyRoomID = genRoomGID(0)

		Groups.AddMembers(saveName, "", msgObj)

		if group.Members != nil {
			t.Fatal("Incorrectly added member to private group")
		}
	})
}

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

func genUserGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "users/" + StringWithCharset(length, "0123456789")
}

func genRoomGID(length int) string {
	if length == 0 {
		length = 10
	} //Defaulting 10

	return "spaces/" + StringWithCharset(length, "0123456789")
}
