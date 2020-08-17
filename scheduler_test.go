package main

import (
	"strings"
	"testing"
	"time"
)

func TestCreateOnetime(t *testing.T) {
	wantedGroupName := genRandName(10)

	args := make(Arguments)
	args["dateTime"] = time.Now().Add(time.Hour).Format(time.RFC3339)
	args["label"] = RandString(10)
	args["message"] = RandString(20)
	args["groupName"] = wantedGroupName

	Groups := make(GroupMap)
	Groups[strings.ToLower(wantedGroupName)] = new(Group)
	Groups[strings.ToLower(wantedGroupName)].ID = 1

	msgObj := messageResponse{}
	msgObj.Room.GID = genRoomGID(0)
	msgObj.Message.Sender.GID = genUserGID(0)
	msgObj.Message.Thread.Name = RandString(10)

	t.Run("Successfully adds onetime event", func(t *testing.T) {
		scheduler := make(ScheduleMap)

		scheduler.CreateOnetime(args, Groups, msgObj)

		gotSchedule := scheduler[msgObj.Room.GID+":"+args["label"]]

		if gotSchedule.SessKey != msgObj.Room.GID+":"+msgObj.Message.Sender.GID {
			t.Errorf("Bad session key\nGot: %s\nWanted: %s\n",
				gotSchedule.SessKey,
				msgObj.Room.GID+":"+msgObj.Message.Sender.GID,
			)
		}

		if gotSchedule.IsRecurring {
			t.Error("Onetime schedule marked as recurring")
		}

		if gotSchedule.ExecuteOn.Format(time.RFC3339) != args["dateTime"] {
			t.Errorf("Incorrect time set to execute on\nGot: %s\nWanted: %s\n",
				gotSchedule.ExecuteOn.Format(time.RFC3339),
				args["dateTime"],
			)
		}

		if gotSchedule.Group.ID != Groups[strings.ToLower(wantedGroupName)].ID {
			t.Errorf("Incorrect group ID\nGot: %+v\nWanted: %+v\n",
				gotSchedule.Group.ID,
				Groups[strings.ToLower(wantedGroupName)].ID,
			)
		}

		if gotSchedule.ThreadKey != msgObj.Message.Thread.Name {
			t.Errorf("Incorrect threadkey\nGot: %s\nWanted: %s\n",
				gotSchedule.ThreadKey,
				msgObj.Message.Thread.Name,
			)
		}

		if gotSchedule.MessageLabel != args["label"] {
			t.Errorf("Incorrect message label\nGot: %s\nWanted: %s\n",
				gotSchedule.MessageLabel,
				args["label"],
			)
		}

		if gotSchedule.MessageText != args["message"] {
			t.Errorf("Incorrect message text\nGot: %s\nWanted: %s\n",
				gotSchedule.MessageText,
				args["Message"],
			)
		}

		t.Run("Starts timer", func(t *testing.T) {
			time.Sleep(time.Microsecond)

			if gotSchedule.timer == nil {
				t.Errorf("Timer not started.")
			}
		})
	})
}

func TestGetLables(t *testing.T) {
	sm := make(ScheduleMap)

	t.Run("Correctly returns nothing if empty", func(t *testing.T) {
		gotLabels := sm.GetLabels()

		if len(gotLabels) != 0 {
			t.Errorf("Returns labels when none should exist.\nFound %d",
				len(gotLabels),
			)
		}
	})

	wantedLabel := RandString(10)
	sm[wantedLabel] = new(Schedule)

	t.Run("Correctly returns one label", func(t *testing.T) {
		gotLabels := sm.GetLabels()

		if len(gotLabels) != 1 {
			t.Errorf("Incorrect amount of labels found.\nShould be 1 Found %d",
				len(gotLabels),
			)
		}

		if gotLabels[0] != wantedLabel {
			t.Errorf("Incorrect label returned.\nGot: %s\nWanted: %s",
				gotLabels[0],
				wantedLabel,
			)
		}
	})

	sm[RandString(10)] = new(Schedule)

	t.Run("Correctly returns more than one label", func(t *testing.T) {
		gotLabels := sm.GetLabels()

		if len(gotLabels) != 2 {
			t.Errorf("Incorrect amount of labels found.\nShould be 2 Found %d",
				len(gotLabels),
			)
		}
	})
}
