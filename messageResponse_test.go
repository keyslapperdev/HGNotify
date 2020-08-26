package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
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
		Groups := make(GroupMap)
		wantedGroupName := genRandName(10)
		msgObj := newMsgObj
		actions := []string{"create", "add", "remove", "disband", "restrict", "list", "syncgroup", "syncallgroups", "usage", "help"}

		for _, wantedAction := range actions {
			msgObj.Message.Text = BotName + " " + wantedAction + " " + wantedGroupName
			args, msg, okay := msgObj.ParseArgs(Groups)

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

	t.Run("Schedule onetime parsed and returned", func(t *testing.T) {
		Groups := make(GroupMap)
		msgObj := newMsgObj
		wantedAction := "schedule"
		wantedSubAction := "onetime"
		wantedLabel := RandString(10)
		wantedDatetime := time.Now().Add(time.Hour).Format(time.RFC3339)
		wantedGroupName := genRandName(10)

		Groups[strings.ToLower(wantedGroupName)] = new(Group)

		msgObj.Message.Text = fmt.Sprintf("%s %s %s %s %v %s some message",
			BotName, wantedAction, wantedSubAction,
			wantedLabel, wantedDatetime, wantedGroupName,
		)
		args, msg, okay := msgObj.ParseArgs(Groups)

		if !okay {
			t.Fatalf("Error parsing schedule onetime request: %s", msg)
		}

		if args["action"] != wantedAction {
			t.Fatalf("Action %q not returned, got: %q",
				wantedAction,
				args["action"],
			)
		}

		if args["subAction"] != wantedSubAction {
			t.Fatalf("Sub action %q not returned, got: %q",
				wantedSubAction,
				args["subAction"],
			)
		}

		if args["label"] != wantedLabel {
			t.Fatalf("Label %q not returned, got: %q",
				wantedLabel,
				args["label"],
			)
		}

		if args["dateTime"] != wantedDatetime {
			t.Fatalf("DateTime %q not returned, got: %q",
				wantedDatetime,
				args["dateTime"],
			)
		}

		if args["groupName"] != wantedGroupName {
			t.Fatalf("Group Name %q not returned, got: %q",
				wantedGroupName,
				args["groupName"],
			)
		}
	})

	t.Run("Properly dispatches list action", func(t *testing.T) {
		Groups := make(GroupMap)
		msgObj := newMsgObj
		wantedAction := "schedule"
		wantedSubAction := "list"

		Groups[strings.ToLower(genRandName(10))] = new(Group)

		msgObj.Message.Text = fmt.Sprintf("%s %s %s ",
			BotName, wantedAction, wantedSubAction,
		)
		args, msg, okay := msgObj.ParseArgs(Groups)

		if !okay {
			t.Fatalf("Error parsing schedule onetime request: %s", msg)
		}

		if args["action"] != wantedAction {
			t.Fatalf("Action %q not returned, got: %q",
				wantedAction,
				args["action"],
			)
		}

		if args["subAction"] != wantedSubAction {
			t.Fatalf("Sub action %q not returned, got: %q",
				wantedSubAction,
				args["subAction"],
			)
		}
	})

	t.Run("Properly identifies as from admin", func(t *testing.T) {
		Groups := make(GroupMap)
		msgObj := newMsgObj
		MasterID = genUserGID(0)

		msgObj.Room.Type = "DM"                 // Master only recognized via DM
		msgObj.Message.Text = BotName + " list" // An action must be passed
		msgObj.Message.Sender.GID = MasterID
		msgObj.FromMaster = false

		_, msg, okay := msgObj.ParseArgs(Groups)

		if !okay {
			t.Fatalf("Something went wrong: %q", msg)
		}

		if !msgObj.FromMaster {
			t.Fatal("Message object is not noted as from the admin")
		}
	})

	t.Run("Properly finds group name for notify", func(t *testing.T) {
		Groups := make(GroupMap)
		wantedGroupName := strings.ToLower(genRandName(10))
		msgObj := newMsgObj

		t.Run("Notify within message", func(t *testing.T) {
			msgObj.Message.Text = "Some before text " + BotName + " " + wantedGroupName + " Some test text"

			args, msg, okay := msgObj.ParseArgs(Groups)

			if !okay {
				t.Fatalf("Something went wrong: %q", msg)
			}

			if args["action"] != "notify" || args["groupName"] != wantedGroupName {
				t.Fatalf("Notify action not properly parsed\nObject Result: %+v", args)
			}
		})

		t.Run("Notify in front of message", func(t *testing.T) {
			Groups := make(GroupMap)
			Groups[wantedGroupName] = new(Group) //Group must exist for this form
			msgObj.Message.Text = BotName + " " + wantedGroupName + " Some test text"

			args, msg, okay := msgObj.ParseArgs(Groups)

			if !okay {
				t.Fatalf("Something went wrong: %q", msg)
			}

			if args["action"] != "notify" || args["groupName"] != wantedGroupName {
				t.Fatalf("Notify action not properly parsed\nObject Result: %+v", args)
			}
		})
	})

	t.Run("Properly notes self when provided", func(t *testing.T) {
		Groups := make(GroupMap)
		msgObj := newMsgObj

		msgObj.Message.Text = BotName + " add groupName Self"

		args, msg, okay := msgObj.ParseArgs(Groups)

		if !okay {
			t.Fatalf("Something went wrong: %q", msg)
		}

		if args["self"] == "" {
			t.Fatal("Self not successfully returned")
		}
	})
}

func TestInspectMessage(t *testing.T) {
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

	actions := []string{"notify", "create", "add", "remove", "disband", "restrict", "list", "syncgroup", "syncallgroups"}

	t.Run("Correctly calls method for given action", func(t *testing.T) {
		for _, action := range actions {
			MockGroups := MockGroupMap{}
			MockSchedule := MockScheduler{}
			msgObj := newMsgObj
			args := make(Arguments)

			args["action"] = action

			inspectMessage(MockGroups, MockSchedule, msgObj, args)

			_, Called := MockGroups[action]
			if !Called {
				t.Fatalf("Groups not called for action %q\nGot: %+v", action, MockGroups)
			}
		}
	})

	scheduleSubActions := []string{"onetime", "list"}

	t.Run("Correctly calls method for given schedule sub action", func(t *testing.T) {
		for _, action := range scheduleSubActions {
			MockGroups := MockGroupMap{}
			MockSchedule := MockScheduler{}
			msgObj := newMsgObj
			args := make(Arguments)

			args["action"] = "schedule"
			args["subAction"] = action

			inspectMessage(MockGroups, MockSchedule, msgObj, args)

			_, Called := MockSchedule[action]
			if !Called {
				t.Fatalf("Scheduler not called for sub action %q\n", action)
			}
		}
	})
}

// Test helper funcs
type MockGroupMap map[string]bool

func (mgm MockGroupMap) Create(string, string, messageResponse) string {
	mgm["create"] = true
	return ""
}
func (mgm MockGroupMap) Disband(string, messageResponse) string {
	mgm["disband"] = true
	return ""
}
func (mgm MockGroupMap) AddMembers(string, string, messageResponse) string {
	mgm["add"] = true
	return ""
}
func (mgm MockGroupMap) RemoveMembers(string, string, messageResponse) string {
	mgm["remove"] = true
	return ""
}
func (mgm MockGroupMap) Restrict(string, messageResponse) string {
	mgm["restrict"] = true
	return ""
}
func (mgm MockGroupMap) Notify(string, messageResponse) string {
	mgm["notify"] = true
	return ""
}
func (mgm MockGroupMap) List(string, messageResponse) string {
	mgm["list"] = true
	return ""
}
func (mgm MockGroupMap) SyncGroupMembers(string, messageResponse) string {
	mgm["syncgroup"] = true
	return ""
}
func (mgm MockGroupMap) SyncAllGroups(messageResponse) string {
	mgm["syncallgroups"] = true
	return ""
}

//Unused, just needs to exist for the interface
func (mgm MockGroupMap) GetGroup(string) *Group { return new(Group) }
func (mgm MockGroupMap) IsGroup(string) bool    { return true }

type MockScheduler map[string]bool

func (ms MockScheduler) CreateOnetime(args Arguments, Groups GroupMgr, msgObj messageResponse) string {
	ms["onetime"] = true
	return ""
}

func (ms MockScheduler) List(msgObj messageResponse) string {
	ms["list"] = true
	return ""
}
