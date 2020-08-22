package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSetupTables(t *testing.T) {
	db := Logger.DB

	t.Run("Setup correct tables", func(t *testing.T) {
		Logger.SetupTables()

		gotTables := make([]struct{ TableName string }, 0)
		db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'hgnotify_beta';").Scan(&gotTables)

		wantedTables := []string{"notify_logs", "members", "groups", "schedules"}

		for _, wantedTable := range wantedTables {
			var found bool

			for _, gotTable := range gotTables {
				if wantedTable == gotTable.TableName {
					found = true
				}
			}

			if !found {
				t.Fatalf("Table: %q not created.", wantedTable)
			}
		}
	})
}

func TestSaveCreatedGroup(t *testing.T) {
	db := Logger.DB

	t.Run("Correctly adds empty group", func(t *testing.T) {
		wantedGroup := &Group{Name: RandString(10)}

		Logger.SaveCreatedGroup(wantedGroup)

		var gotGroup struct{ Name string }
		db.Raw("SELECT name FROM groups WHERE name = ?;", wantedGroup.Name).Scan(&gotGroup)

		if gotGroup.Name != wantedGroup.Name {
			t.Fatalf("Incorrect group name in database.\nWanted: %q\nGot: %v",
				wantedGroup.Name,
				gotGroup.Name,
			)
		}
	})

	t.Run("Correctly adds group with memebers", func(t *testing.T) {
		wantedGroup := &Group{
			Name: RandString(10),
			Members: []Member{
				{Name: RandString(15), GID: "users/" + StringWithCharset(10, "0123456789")},
			},
		}

		Logger.SaveCreatedGroup(wantedGroup)

		var gotMember Member
		db.Raw("SELECT * FROM members WHERE group_id = (SELECT id FROM groups WHERE name = ? );", wantedGroup.Name).Scan(&gotMember)

		if wantedGroup.Members[0].Name != gotMember.Name ||
			wantedGroup.Members[0].GroupID != gotMember.GroupID ||
			wantedGroup.Members[0].GID != gotMember.GID {
			t.Fatalf("Incorrect group in database.\nWanted: %+v\nGot: %+v",
				wantedGroup.Members,
				gotMember,
			)
		}
	})
}

func TestDBDisbandGroup(t *testing.T) {
	db := Logger.DB

	t.Run("Successfully Removes Group", func(t *testing.T) {
		unwantedGroup := &Group{Name: RandString(10)}
		db.Create(&unwantedGroup)

		Logger.DisbandGroup(unwantedGroup)

		var gotGroup Group
		var emptyGroup Group

		db.Where(&unwantedGroup).First(&gotGroup)

		if !reflect.DeepEqual(gotGroup, emptyGroup) {
			t.Fatalf("Got group that should be deleted:\nGot: %+v", gotGroup)
		}
	})
}

func TestUpdatePrivacyDB(t *testing.T) {
	db := Logger.DB

	wantedName := genRandName(0)
	wantedRoomID := genRoomGID(0)

	initialGroup := &Group{Name: wantedName}

	db.Model(&Group{}).Create(initialGroup)

	t.Run("Correctly sets privacy", func(t *testing.T) {
		initialGroup.IsPrivate = true
		initialGroup.PrivacyRoomID = wantedRoomID

		Logger.UpdatePrivacyDB(initialGroup)

		var gotGroup Group
		db.Raw("SELECT name, privacy_room_id, is_private FROM groups WHERE name = ?", wantedName).
			Scan(&gotGroup)

		if gotGroup.PrivacyRoomID != wantedRoomID || !gotGroup.IsPrivate {
			t.Fatalf("Did not set room as private properly:\nGot: %+v", gotGroup)
		}
	})

	t.Run("Correctly unsets privacy", func(t *testing.T) {
		initialGroup.IsPrivate = false
		initialGroup.PrivacyRoomID = ""

		Logger.UpdatePrivacyDB(initialGroup)

		var gotGroup Group
		db.Raw("SELECT name, privacy_room_id, is_private FROM groups WHERE name = ?", wantedName).
			Scan(&gotGroup)

		if gotGroup.PrivacyRoomID != "" || gotGroup.IsPrivate {
			t.Fatalf("Did not set room as public properly:\nGot: %+v", gotGroup)
		}
	})
}

func TestSaveMemberAddition(t *testing.T) {
	db := Logger.DB
	initGroup := &Group{Name: genRandName(0)}

	db.Model(&Group{}).Create(initGroup)

	t.Run("Correctly adds member to empty group", func(t *testing.T) {
		wantedMemberName := genRandName(0)

		initGroup.Members = append(initGroup.Members, Member{
			Name: wantedMemberName,
		})

		Logger.SaveMemberAddition(initGroup)

		gotMembers := make([]Member, 0)
		db.Raw("SELECT * from members WHERE group_id = ?", initGroup.ID).Scan(&gotMembers)

		if len(gotMembers) != 1 {
			t.Fatalf("Incorrect number of members returned:\nExpected: 1\nGot: %d", len(gotMembers))
		}

		if gotMembers[0].Name != wantedMemberName {
			t.Fatalf("Retrieved incorrect member:\nWanted: %q\nGot %q",
				wantedMemberName,
				gotMembers[0].Name,
			)
		}
	})

	t.Run("Correctly adds member to non empty group", func(t *testing.T) {
		wantedMemberName := genRandName(0)

		initGroup.Members = append(initGroup.Members, Member{
			Name: wantedMemberName,
		})

		Logger.SaveMemberAddition(initGroup)

		gotMembers := make([]Member, 0)
		db.Raw("SELECT * from members WHERE group_id = ?", initGroup.ID).Scan(&gotMembers)

		if len(gotMembers) != 2 {
			t.Fatalf("Incorrect number of members returned:\nExpected: 2\nGot: %d", len(gotMembers))
		}

		if gotMembers[1].Name != wantedMemberName {
			t.Fatalf("Retrieved incorrect member:\nWanted: %q\nGot %q",
				wantedMemberName,
				gotMembers[0].Name,
			)
		}
	})
}

func TestSaveMemberRemoval(t *testing.T) {
	db := Logger.DB

	initGroup := &Group{
		Name: genRandName(0),
		Members: []Member{
			{Name: genRandName(0)},
			{Name: genRandName(0)},
			{Name: genRandName(0)},
		},
	}

	db.Model(&Group{}).Create(&initGroup)

	member1 := initGroup.Members[0]
	member2 := initGroup.Members[1]
	member3 := initGroup.Members[2]

	t.Run("Successfully removes multiple members", func(t *testing.T) {
		//Simulate member having already been removed from memory
		initGroup.Members = initGroup.Members[len(initGroup.Members)-1:]

		Logger.SaveMemberRemoval(initGroup, []Member{member1, member2})

		gotMembers := make([]Member, 0)
		db.Raw("SELECT * from members WHERE deleted_at IS NULL AND group_id = ?", initGroup.ID).Scan(&gotMembers)

		if len(gotMembers) != 1 {
			t.Fatalf("Incorrect number of members returned:\nExpected: 1\nGot: %d", len(gotMembers))
		}

		if gotMembers[0].Name != member3.Name {
			t.Fatal("Incorrect member remaining")
		}
	})

	t.Run("Successfully removes single member", func(t *testing.T) {
		initGroup.Members = []Member{}

		Logger.SaveMemberRemoval(initGroup, []Member{member3})

		gotMembers := make([]Member, 0)
		db.Raw("SELECT * from members WHERE deleted_at IS NULL AND group_id = ?", initGroup.ID).Scan(&gotMembers)

		if len(gotMembers) != 0 {
			t.Fatalf("Incorrect number of members returned:\nExpected: 0\nGot: %d", len(gotMembers))
		}

	})
}

func TestGetGroupsFromDB(t *testing.T) {
	db := Logger.DB
	groups := make(GroupMap)

	groupNames := []string{genRandName(0), genRandName(0), genRandName(0)}

	memberNames := []string{genRandName(0), genRandName(0), genRandName(0)}

	db.Model(&Group{}).Create(&Group{
		Name: groupNames[0],
		Members: []Member{
			{Name: memberNames[0]},
			{Name: memberNames[1]},
		},
	})

	db.Model(&Group{}).Create(&Group{
		Name: groupNames[1],
		Members: []Member{
			{Name: memberNames[0]},
		},
	})

	db.Model(&Group{}).Create(&Group{
		Name: groupNames[2],
		Members: []Member{
			{Name: memberNames[0]},
			{Name: memberNames[1]},
			{Name: memberNames[2]},
		},
	})

	t.Run("Successfully retrieves groups from DB", func(t *testing.T) {
		Logger.GetGroupsFromDB(groups)

		for _, groupName := range groupNames {
			saveName := strings.ToLower(groupName)
			group, exist := groups[saveName]

			if !exist {
				t.Fatal("Wanted group wasn't retrieved")
			}

			for i, member := range group.Members {
				if member.Name != memberNames[i] {
					t.Fatal("Wanted member not associated with correct group")
				}
			}
		}
	})
}

func TestSaveScheduledEvent(t *testing.T) {
	db := Logger.DB

	group := &Group{Name: RandString(10)}
	db.Create(group)

	schedule := &Schedule{
		SessKey:      "sesskey",
		Creator:      "me",
		IsRecurring:  false,
		ExecuteOn:    time.Now().Add(time.Hour * 2),
		GroupID:      group.ID,
		ThreadKey:    "threadkey",
		MessageLabel: "messageLabel",
		MessageText:  "text",
	}

	t.Run("Correctly adds scheduled onetime event", func(t *testing.T) {

		Logger.SaveScheduledEvent(schedule)

		var gotSchedule Schedule
		db.Raw("SELECT * FROM schedules WHERE sess_key = ?", "sesskey").Scan(&gotSchedule)

		if reflect.DeepEqual(schedule, gotSchedule) {
			t.Errorf("Incorrect information returned:\nWanted: %+v\nGot: %+v",
				schedule,
				gotSchedule,
			)
		}
	})

	t.Run("Correctly updates existing scheduled onetime event", func(t *testing.T) {
		wantedText := "wanted text"
		schedule.MessageText = wantedText

		Logger.SaveScheduledEvent(schedule)

		var gotSchedule Schedule
		db.Raw("SELECT * FROM schedules WHERE sess_key = ?", "sesskey").Scan(&gotSchedule)

		if gotSchedule.MessageText != wantedText {
			t.Errorf("Schedule not updated:\nGot Text: %s\nWanted Text: %s",
				gotSchedule.MessageText,
				schedule.MessageText,
			)
		}
	})

}

func TestGetGroupByID(t *testing.T) {
	db := Logger.DB

	t.Run("Correctly retreives empty group", func(t *testing.T) {
		wantedGroup := &Group{Name: RandString(10)}
		db.Create(wantedGroup)

		gotGroup := Logger.GetGroupByID(wantedGroup.ID)

		if gotGroup.Name != wantedGroup.Name {
			t.Errorf("Incorrect group name in database.\nWanted: %q\nGot: %q",
				wantedGroup.Name,
				gotGroup.Name,
			)
		}
	})

	t.Run("Correctly retreives group with one member", func(t *testing.T) {
		wantedGroup := &Group{
			Name: RandString(10),
			Members: []Member{
				{Name: RandString(15), GID: "users/" + StringWithCharset(10, "0123456789")},
			},
		}
		db.Create(wantedGroup)

		gotGroup := Logger.GetGroupByID(wantedGroup.ID)

		if len(gotGroup.Members) == 0 {
			t.Fatal("No members returned")
		}

		if wantedGroup.Members[0].Name != gotGroup.Members[0].Name ||
			wantedGroup.Members[0].GroupID != gotGroup.Members[0].GroupID ||
			wantedGroup.Members[0].GID != gotGroup.Members[0].GID {
			t.Errorf("Incorrect member retrieved from database.\nWanted: %+v\nGot: %+v",
				wantedGroup.Members,
				gotGroup.Members,
			)
		}
	})

	t.Run("Correctly retreives group with multiple members", func(t *testing.T) {
		wantedGroup := &Group{
			Name: RandString(10),
			Members: []Member{
				{Name: RandString(15), GID: "users/" + StringWithCharset(10, "0123456789")},
				{Name: RandString(15), GID: "users/" + StringWithCharset(10, "0123456789")},
				{Name: RandString(15), GID: "users/" + StringWithCharset(10, "0123456789")},
			},
		}
		db.Create(wantedGroup)

		gotGroup := Logger.GetGroupByID(wantedGroup.ID)

		if len(gotGroup.Members) == 0 {
			t.Fatal("No members returned")
		}

		for i, gotMember := range wantedGroup.Members {
			if wantedGroup.Members[i].Name != gotMember.Name ||
				wantedGroup.Members[i].GroupID != gotMember.GroupID ||
				wantedGroup.Members[i].GID != gotMember.GID {
				t.Errorf("Incorrect member retrieved from database.\nWanted: %+v\nGot: %+v",
					wantedGroup.Members[i],
					gotGroup.Members,
				)
			}
		}
	})
}
