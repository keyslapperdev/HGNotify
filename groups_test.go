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
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
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
					Name: randString(10),
					GID:  genUserGID(0),
					Type: "HUMAN",
				},
			},
			Type: "USER_MENTION",
		},
		{
			Called: userMention{
				User: User{
					Name: randString(10),
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
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
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

func TestRestrictGroup(t *testing.T) {
	Logger.Active(false)

	saveName := strings.ToLower(randString(10))

	wantedRoomGID := genRoomGID(0)
	msgObj := messageResponse{
		Room: space{
			GID: wantedRoomGID,
		},
	}

	t.Run("Set room ID when toggling privacy", func(t *testing.T) {
		Groups := make(GroupList)
		Groups[saveName] = new(Group)
		group := Groups[saveName]

		Groups.Restrict(saveName, msgObj)

		if group.PrivacyRoomID != wantedRoomGID || !group.IsPrivate {
			t.Fatal("Privacy not set properly")
		}
	})

	t.Run("Toggle Privacy properly", func(t *testing.T) {
		Groups := make(GroupList)
		Groups[saveName] = new(Group)
		group := Groups[saveName]

		Groups.Restrict(saveName, msgObj)

		if !group.IsPrivate {
			t.Fatal("Didn't set privacy")
		}

		origState := group.IsPrivate

		Groups.Restrict(saveName, msgObj)

		if origState == group.IsPrivate {
			t.Fatal("Privacy did not toggle properly from on to off")
		}

		Groups.Restrict(saveName, msgObj)

		if origState != group.IsPrivate {
			t.Fatal("Privacy did not toggle properly from off to on")
		}
	})
}

func TestNotifyGroup(t *testing.T) {
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
			},
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
	}

	BotName = "@TestBot"
	testText := " test text."

	t.Run("Group with single user notified", func(t *testing.T) {
		wantedGID := genUserGID(0)
		group.Members = []Member{{
			GID: wantedGID,
		}}

		msgObj.Message.Text = BotName + " " + saveName + testText

		gotText := Groups.Notify(saveName, msgObj)

		if strings.Contains(gotText, BotName) {
			t.Fatalf("Botname should not be in generated text\nGot: %q",
				gotText,
			)
		}

		if strings.Contains(gotText, saveName) {
			t.Fatalf("Group name should not be in text\nGot: %q",
				gotText,
			)
		}

		if !strings.Contains(gotText, selfName) {
			t.Fatalf("Sender name should be included in message\nGot %q",
				gotText,
			)
		}

		if !strings.Contains(gotText, testText) {
			t.Fatalf("Result text should contain the message\nGot: %q",
				gotText,
			)
		}

		if !strings.Contains(gotText, wantedGID) {
			t.Fatalf("Result text should contain single user's GID\nGot: %q",
				gotText,
			)
		}
	})

	t.Run("Group with multiple users notified", func(t *testing.T) {
		wantedGID1 := genUserGID(0)
		wantedGID2 := genUserGID(0)

		group.Members = []Member{
			{GID: wantedGID1},
			{GID: wantedGID2},
		}

		msgObj.Message.Text = BotName + " " + saveName + testText

		gotText := Groups.Notify(saveName, msgObj)

		if strings.Contains(gotText, BotName) {
			t.Fatalf("Botname should not be in generated text\nGot: %q",
				gotText,
			)
		}

		if strings.Contains(gotText, saveName) {
			t.Fatalf("Group name should not be in text\nGot: %q",
				gotText,
			)
		}

		if !strings.Contains(gotText, selfName) {
			t.Fatalf("Sender name should be included in message\nGot %q",
				gotText,
			)
		}

		if !strings.Contains(gotText, testText) {
			t.Fatalf("Result text should contain the message\nGot: %q",
				gotText,
			)
		}

		wantedGrouping := "<" + wantedGID1 + "> <" + wantedGID2 + ">"
		if !strings.Contains(gotText, wantedGrouping) {
			t.Fatalf("Result text should contain all user mentions\nGot: %q",
				gotText,
			)
		}
	})
}

func TestListGroups(t *testing.T) {
	Logger.Active(false)

	selfName := randString(10)
	msgObj := messageResponse{
		Message: message{
			Sender: User{
				Name: selfName,
			},
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
	}

	t.Run("Lists multiple groups", func(t *testing.T) {
		groupName1 := "group1"
		groupName2 := "group2"
		groupName3 := "group3"

		Groups := make(GroupList)

		Groups[groupName1] = &Group{Name: groupName1}
		Groups[groupName2] = &Group{Name: groupName2}
		Groups[groupName3] = &Group{Name: groupName3}

		gotText := Groups.List("", msgObj)

		if !strings.Contains(gotText, groupName1) ||
			!strings.Contains(gotText, groupName2) ||
			!strings.Contains(gotText, groupName3) {
			t.Fatalf("Wanted groups not in list output\nGot: %q", gotText)
		}
	})

	t.Run("List excludes private groups", func(t *testing.T) {
		groupName1 := "group1"
		groupName2 := "group2"
		privateGroupName := "privategroup"

		Groups := make(GroupList)

		Groups[groupName1] = &Group{Name: groupName1}
		Groups[groupName2] = &Group{Name: groupName2}
		Groups[privateGroupName] = &Group{
			Name:          privateGroupName,
			IsPrivate:     true,
			PrivacyRoomID: genRoomGID(0),
		}

		gotText := Groups.List("", msgObj)

		if !strings.Contains(gotText, groupName1) || !strings.Contains(gotText, groupName2) {
			t.Fatalf("Wanted groups not in list output\nGot: %q", gotText)
		}

		if strings.Contains(gotText, privateGroupName) {
			t.Fatal("Should not contain private group in output")
		}
	})

	t.Run("List includes private group in privacy room", func(t *testing.T) {
		groupName1 := "group1"
		groupName2 := "group2"
		privateGroupName1 := "privategroup1"
		privateGroupName2 := "privategroup2"

		Groups := make(GroupList)

		wantedPrivacyRoomID := genRoomGID(0)

		Groups[groupName1] = &Group{Name: groupName1}
		Groups[groupName2] = &Group{Name: groupName2}
		Groups[privateGroupName1] = &Group{
			Name:          privateGroupName1,
			IsPrivate:     true,
			PrivacyRoomID: wantedPrivacyRoomID,
		}
		Groups[privateGroupName2] = &Group{
			Name:          privateGroupName2,
			IsPrivate:     true,
			PrivacyRoomID: genRoomGID(0),
		}

		msgObj.Room.GID = wantedPrivacyRoomID

		gotText := Groups.List("", msgObj)

		if !strings.Contains(gotText, groupName1) ||
			!strings.Contains(gotText, groupName2) ||
			!strings.Contains(gotText, privateGroupName1) {
			t.Fatalf("Wanted groups not in list output\nGot: %q", gotText)
		}

		if strings.Contains(gotText, privateGroupName2) {
			t.Fatal("Private groups from other rooms shouldn't show up in other rooms with privacy")
		}
	})

	t.Run("Listing single group displays member information", func(t *testing.T) {
		wantedGroupName := strings.ToLower(randString(10))
		wantedName := randString(10)
		wantedGID := genUserGID(0)

		Groups := make(GroupList)
		Groups[wantedGroupName] = &Group{
			Name: wantedGroupName,
			Members: []Member{{
				Name: wantedName,
				GID:  wantedGID,
			}},
		}

		gotText := Groups.List(wantedGroupName, msgObj)

		if !strings.Contains(gotText, wantedGroupName) ||
			!strings.Contains(gotText, wantedName) ||
			!strings.Contains(gotText, wantedGID) {
			t.Fatal("Group detailed list doesn't contain proper information")
		}
	})

	t.Run("Does not allow listing of a private group", func(t *testing.T) {
		unWantedGroupName := strings.ToLower(randString(10))

		Groups := make(GroupList)
		Groups[unWantedGroupName] = &Group{
			Name: unWantedGroupName,
			Members: []Member{{
				Name: genUserGID(0),
				GID:  genRoomGID(0),
			}},
			IsPrivate:     true,
			PrivacyRoomID: genRoomGID(0),
		}

		gotText := Groups.List(unWantedGroupName, msgObj)

		if !strings.Contains(gotText, "you may not view it") {
			t.Fatal("Should not be able to view private group")
		}

	})
}

//Test helper data
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(
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
