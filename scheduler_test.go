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

		if gotSchedule.GroupID != Groups[strings.ToLower(wantedGroupName)].ID {
			t.Errorf("Incorrect group ID\nGot: %d\nWanted: %d\n",
				gotSchedule.GroupID,
				Groups[strings.ToLower(wantedGroupName)].ID,
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
