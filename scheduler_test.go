package main

import (
	"strings"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type schedCreateFunc func(Arguments, GroupMgr, messageResponse) string

func TestCreateOnetime(t *testing.T) {
	scheduler := make(ScheduleMap)
	testCreateSchedule(t, scheduler, scheduler.CreateOnetime, false)
}

func TestCreateRecurring(t *testing.T) {
	scheduler := make(ScheduleMap)
	testCreateSchedule(t, scheduler, scheduler.CreateRecurring, true)
}

func testCreateSchedule(t *testing.T, scheduler ScheduleMap, createFunc schedCreateFunc, isRecurring bool) {
	Logger.Active(false)

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

	t.Run("Successfully creates scheduled event", func(t *testing.T) {
		createFunc(args, Groups, msgObj)

		gotSchedule := scheduler[msgObj.Room.GID+":"+args["label"]]

		if gotSchedule.SessKey != msgObj.Room.GID+":"+msgObj.Message.Sender.GID {
			t.Errorf("Bad session key\nGot: %s\nWanted: %s\n",
				gotSchedule.SessKey,
				msgObj.Room.GID+":"+msgObj.Message.Sender.GID,
			)
		}

		if gotSchedule.IsRecurring != isRecurring {
			t.Error("Schedule marked incorrectly")
		}

		if gotSchedule.ExecuteOn.Format(time.RFC3339) != args["dateTime"] {
			t.Errorf("Incorrect time set to execute on\nGot: %s\nWanted: %s\n",
				gotSchedule.ExecuteOn.Format(time.RFC3339),
				args["dateTime"],
			)
		}

		if gotSchedule.GroupID != Groups[strings.ToLower(wantedGroupName)].ID {
			t.Errorf("Incorrect group ID\nGot: %+v\nWanted: %+v\n",
				gotSchedule.GroupID,
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

		// Timer is started in async, so I'm adding a little sleep
		// give it the time it needs
		t.Run("Starts timer", func(t *testing.T) {
			time.Sleep(time.Millisecond)

			if gotSchedule.timer == nil {
				t.Errorf("Timer not started.")
			}
		})
	})

	t.Run("Successfully updates scheduled event", func(t *testing.T) {
		updatedGroupName := genRandName(10)

		Groups[strings.ToLower(updatedGroupName)] = new(Group)
		Groups[strings.ToLower(updatedGroupName)].ID = 2

		args["groupName"] = updatedGroupName

		createFunc(args, Groups, msgObj)

		gotSchedule := scheduler[msgObj.Room.GID+":"+args["label"]]

		if gotSchedule.UpdatedOn.IsZero() {
			t.Error("Updating message didn't update updateOn time")
		}

		if gotSchedule.GroupID != Groups[strings.ToLower(updatedGroupName)].ID {
			t.Errorf("Field 'GroupID' not updated \nGot: %+v\nWanted: %+v\n",
				gotSchedule.GroupID,
				Groups[strings.ToLower(updatedGroupName)].ID,
			)
		}
	})
}

func TestListSchedules(t *testing.T) {
	roomGID := genRoomGID(20)

	msgObj := messageResponse{}
	msgObj.Room.GID = roomGID

	wantedSchedule1 := &Schedule{
		Creator:      "Meee",
		IsRecurring:  false,
		CreatedOn:    time.Now(),
		ExecuteOn:    time.Now().Add(time.Hour * 2),
		GroupID:      1,
		MessageLabel: "MessageLabel",
		MessageText:  "Text",
	}

	scheduleYaml1, err := yaml.Marshal(wantedSchedule1)
	if err != nil {
		t.Error("Problem marshalling schedule to yaml: " + err.Error())
	}

	wantedSchedule2 := &Schedule{
		Creator:      "Meee",
		IsRecurring:  false,
		CreatedOn:    time.Now(),
		ExecuteOn:    time.Now().Add(time.Hour * 2),
		GroupID:      1,
		MessageLabel: "OtherMessageLabel",
		MessageText:  "Text",
	}

	scheduleYaml2, err := yaml.Marshal(wantedSchedule2)
	if err != nil {
		t.Error("Problem marshalling schedule to yaml: " + err.Error())
	}

	sm := make(ScheduleMap)
	sm[roomGID+":"+wantedSchedule1.MessageLabel] = wantedSchedule1
	sm[roomGID+":"+wantedSchedule2.MessageLabel] = wantedSchedule2

	t.Run("Correctly returns list of schedules", func(t *testing.T) {
		gotText := sm.List(msgObj)

		if !strings.Contains(gotText, string(scheduleYaml1)) {
			t.Errorf("Expected to find first schedule:\n %q\nwithin\n %q\n but did not",
				scheduleYaml1,
				gotText,
			)
		}

		if !strings.Contains(gotText, string(scheduleYaml2)) {
			t.Errorf("Expected to find second schedule:\n %q\nwithin\n %q\n but did not",
				scheduleYaml2,
				gotText,
			)
		}
	})
}

func TestRemoveSchedule(t *testing.T) {
	Logger.Active(false)

	roomGID := genRoomGID(10)
	labelToRemove := RandString(10)

	args := make(Arguments)
	args["label"] = labelToRemove

	msgObj := messageResponse{}
	msgObj.Room.GID = roomGID

	schedKeyToRemove := roomGID + ":" + labelToRemove
	sm := make(ScheduleMap)
	sm[schedKeyToRemove] = &Schedule{
		Creator:      "Meee",
		IsRecurring:  false,
		CreatedOn:    time.Now(),
		ExecuteOn:    time.Now().Add(time.Hour * 2),
		GroupID:      1,
		MessageLabel: "MessageLabel",
		MessageText:  "Text",
		timer:        time.NewTimer(time.Hour), // if the schedule is being removed the timer should be running
	}

	sm[roomGID+":"+RandString(10)] = &Schedule{}

	t.Run("Removes single schedule", func(t *testing.T) {
		sm.Remove(args, msgObj)

		if len(sm) != 1 {
			t.Errorf("Incorrect amount of schedules found.\nGot: %d\nExpected 1\n", len(sm))
		}

		if sm.hasSchedule(schedKeyToRemove) {
			t.Error("Requested schedule not removed")
		}
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
