package main

import (
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	newMsgObj := messageResponse{
		Message: message{
			Sender: User{
				Name: genRandName(10),
				GID:  genUserGID(0),
				Type: "HUMAN",
			},
		},
		Room: space{
			GID:  genRoomGID(0),
			Type: "ROOM",
		},
	}

	t.Run("Action and Group args are parsed and returned", func(t *testing.T) {
		wantedGroupName := genRandName(10)
		msgObj := newMsgObj
		actions := []string{"create", "add", "remove", "disband", "restrict", "list", "syncgroup", "syncallgroups", "usage", "help"}

		for _, wantedAction := range actions {
			msgObj.Message.Text = BotName + " " + wantedAction + " " + wantedGroupName
			args, msg, okay := msgObj.parseArgs()

			if !okay {
				t.Fatalf("Error parsing argument %q: %s", wantedAction, msg)
			}

			if args["action"] != wantedAction {
				t.Fatalf("Action %q not returned", wantedAction)
			}

			if args["groupName"] != wantedGroupName {
				t.Fatal("Group name not returned")
			}
		}
	})

	t.Run("Properly identifies as from admin", func(t *testing.T) {
		msgObj := newMsgObj
		MasterID = genUserGID(0)

		msgObj.Room.Type = "DM"                 // Master only recognized via DM
		msgObj.Message.Text = BotName + " list" // An action must be passed
		msgObj.Message.Sender.GID = MasterID
		msgObj.FromMaster = false

		_, msg, okay := msgObj.parseArgs()

		if !okay {
			t.Fatalf("Something went wrong: %q", msg)
		}

		if !msgObj.FromMaster {
			t.Fatal("Message object is not noted as from the admin")
		}
	})

	t.Run("Properly finds group name for notify", func(t *testing.T) {
		wantedGroupName := strings.ToLower(genRandName(10))
		msgObj := newMsgObj

		t.Run("Notify within message", func(t *testing.T) {
			msgObj.Message.Text = "Some before text " + BotName + " " + wantedGroupName + " Some test text"

			args, msg, okay := msgObj.parseArgs()

			if !okay {
				t.Fatalf("Something went wrong: %q", msg)
			}

			if args["action"] != "notify" || args["groupName"] != wantedGroupName {
				t.Fatalf("Notify action not properly parsed\nObject Result: %+v", args)
			}
		})

		t.Run("Notify in front of message", func(t *testing.T) {
			Groups = make(GroupMap)
			Groups[wantedGroupName] = new(Group) //Group must exist for this form
			msgObj.Message.Text = BotName + " " + wantedGroupName + " Some test text"

			args, msg, okay := msgObj.parseArgs()

			if !okay {
				t.Fatalf("Something went wrong: %q", msg)
			}

			if args["action"] != "notify" || args["groupName"] != wantedGroupName {
				t.Fatalf("Notify action not properly parsed\nObject Result: %+v", args)
			}
		})
	})

	t.Run("Properly notes self when provided", func(t *testing.T) {
		msgObj := newMsgObj

		msgObj.Message.Text = BotName + " add groupName Self"

		args, msg, okay := msgObj.parseArgs()

		if !okay {
			t.Fatalf("Something went wrong: %q", msg)
		}

		if args["self"] == "" {
			t.Fatal("Self not successfully returned")
		}
	})
}
