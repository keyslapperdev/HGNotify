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

	t.Run("Correctly adds self to group", func(t *testing.T) {
		Groups.AddMembers(saveName, "self", msgObj)

		var foundSelf bool
		for _, member := range group.Members {
			if member.Name == selfName {
				foundSelf = true
			}
		}

		if !foundSelf {
			t.Fatal("Member not added to the group")
		}
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

func TestRemoveMembers(t *testing.T) {
	Logger.Active(false)

	Groups := make(GroupList)

	saveName := strings.ToLower(randString(10))

	Groups[saveName] = new(Group)
	group := Groups[saveName]

	selfGID := randString(10)
	msgObj := messageResponse{
		Message: message{
			Sender: User{
				GID: selfGID,
			},

			Mentions: nil,
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
		Time: "",
	}

	t.Run("Removes single member", func(t *testing.T) {
		memberGID := genUserGID(0)
		group.Members = []Member{{
			GID: memberGID,
		}}

		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: randString(10),
					GID:  memberGID,
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		Groups.RemoveMembers(saveName, "", msgObj)

		if len(group.Members) > 0 {
			t.Fatal("Did not remove member")
		}
	})

	t.Run("Removes single member from group with multiple members", func(t *testing.T) {
		wantedMembers := []Member{
			{GID: genUserGID(0)},
			{GID: genUserGID(0)},
			{GID: genUserGID(0)},
		}

		GIDToRemove := genUserGID(0)
		unWantedMember := Member{
			GID: GIDToRemove,
		}

		group.Members = wantedMembers
		group.Members = append(group.Members, unWantedMember)

		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: randString(10),
					GID:  GIDToRemove,
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		Groups.RemoveMembers(saveName, "", msgObj)

		if len(group.Members) != len(wantedMembers) {
			t.Fatalf("Incorrect member count\nWanted: %d\nGot: %d",
				len(group.Members),
				len(wantedMembers),
			)
		}

		var foundUnwanted bool
		for _, member := range group.Members {
			if member.GID == unWantedMember.GID {
				foundUnwanted = true
			}
		}

		if foundUnwanted {
			t.Fatal("Unwanted member found")
		}
	})

	t.Run("Removes multiple members", func(t *testing.T) {
		unWantedMembers := []Member{
			{GID: genUserGID(0)},
			{GID: genUserGID(0)},
		}

		group.Members = unWantedMembers

		msgObj.Message.Mentions = []annotation{
			{
				Called: userMention{
					User: User{
						Name: randString(10),
						GID:  unWantedMembers[0].GID,
						Type: "HUMAN",
					},
				},
				Type: "USER_MENTION",
			},
			{
				Called: userMention{
					User: User{
						Name: randString(10),
						GID:  unWantedMembers[1].GID,
						Type: "HUMAN",
					},
				},
				Type: "USER_MENTION",
			},
		}

		Groups.RemoveMembers(saveName, "", msgObj)

		if len(group.Members) != 0 {
			t.Fatalf("Incorrect member count should be 0, Got: %d", len(group.Members))
		}

		if _, exist := Groups[saveName]; !exist {
			t.Fatal("Somehow the group was nuked when emptied.... Big problem")
		}
	})

	t.Run("Removes self from group", func(t *testing.T) {
		group.Members = []Member{
			{GID: selfGID},
			{GID: genUserGID(0)},
			{GID: genUserGID(0)},
		}

		Groups.RemoveMembers(saveName, "self", msgObj)

		var foundUnwanted bool
		for _, member := range group.Members {
			if member.GID == msgObj.Message.Sender.Name {
				foundUnwanted = true
			}
		}

		if foundUnwanted {
			t.Fatal("Found self in still in group")
		}
	})

	group.Members = nil

	t.Run("Does not remove member from private group", func(t *testing.T) {
		unWantedGID := genUserGID(0)
		group.Members = []Member{{
			GID: unWantedGID,
		}}

		msgObj.Message.Mentions = []annotation{{
			Called: userMention{
				User: User{
					Name: randString(10),
					GID:  unWantedGID,
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		}}

		group.IsPrivate = true
		group.PrivacyRoomID = genRoomGID(0)

		Groups.RemoveMembers(saveName, "", msgObj)

		if len(group.Members) != 1 {
			t.Fatal("Incorrectly removed member")
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
