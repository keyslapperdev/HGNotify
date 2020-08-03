package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	// I will note. There are sleeps thrown about in the following tests
	// It's all my fault and I'm sorry. The database logging that
	// the application does happens async for speeed. Because of this, the
	// check for the async action could come before the action is done.
	// There is very likely a better way to do it, but I'm not sure yet.
	if os.Getenv("RUN_INTEGRATION") != "true" {
		t.Skip("Must set RUN_INTEGRATION=true to run integration tests")
	}
	Logger.Active(true)
	db := Logger.DB

	TestGroupMap := GroupMap{}
	TestSchedules := ScheduleMap{}
	groupName := genRandName(0)
	saveName := strings.ToLower(groupName)

	server := httptest.NewServer(getRequestHandler(TestGroupMap, TestSchedules))
	contentType := "application/json"

	//Testing the lifecycle of a group from handler to db
	t.Run("Create group", func(t *testing.T) {
		createReqString := BotName + " create " + groupName

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": createReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting creation request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "Created group \\\""+groupName+"\\\" with no users.") {
			t.Fatalf("Creation request unsuccessful: %q", string(respMsg))
		}

		if _, exist := TestGroupMap[saveName]; !exist {
			t.Fatal("Group not added in memory")
		}

		var gotGroup Group
		db.Model(&Group{}).Where(&Group{Name: groupName}).Take(&gotGroup)

		if reflect.DeepEqual(Group{}, gotGroup) {
			t.Fatal("Group not found in DB")
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != createReqString {
			t.Fatal("Create request not logged")
		}
	})

	tGroup := TestGroupMap[saveName]

	t.Run("Add members to group", func(t *testing.T) {
		wantedMembers := []Member{
			{Name: genRandName(0), GID: genUserGID(0)},
			{Name: genRandName(0), GID: genUserGID(0)},
			{Name: genRandName(0), GID: genUserGID(0)},
		}

		addReqString := fmt.Sprintf("%s add %s @%s @%s @%s",
			BotName,
			groupName,
			wantedMembers[0].Name,
			wantedMembers[1].Name,
			wantedMembers[2].Name,
		)

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": addReqString,
			"annotations": []map[string]interface{}{
				{
					"type": "USER_MENTION",
					"userMention": map[string]interface{}{
						"user": map[string]interface{}{
							"name":        wantedMembers[0].GID,
							"displayName": wantedMembers[0].Name,
							"type":        "HUMAN",
						},
						"type": "MENTION",
					},
				},
				{
					"type": "USER_MENTION",
					"userMention": map[string]interface{}{
						"user": map[string]interface{}{
							"name":        wantedMembers[1].GID,
							"displayName": wantedMembers[1].Name,
							"type":        "HUMAN",
						},
						"type": "MENTION",
					},
				},
				{
					"type": "USER_MENTION",
					"userMention": map[string]interface{}{
						"user": map[string]interface{}{
							"name":        wantedMembers[2].GID,
							"displayName": wantedMembers[2].Name,
							"type":        "HUMAN",
						},
						"type": "MENTION",
					},
				},
			},
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting memberAdd request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "I've added the user") {
			t.Fatalf("MemberAdd request unsuccessful: %q", string(respMsg))
		}

		if len(tGroup.Members) != len(wantedMembers) {
			t.Fatalf("Incorrect number of group members in memory:\nWanted: %d\nGot: %d",
				len(tGroup.Members),
				len(wantedMembers),
			)
		}

		for i, wantedMember := range wantedMembers {
			if wantedMember.Name != tGroup.Members[i].Name {
				t.Fatal("Incorrect members saved in memory")
			}
		}

		gotMembers := make([]Member, 0)
		db.Raw("SELECT * from members WHERE group_id = ?", tGroup.ID).Scan(&gotMembers)

		if len(gotMembers) != 3 {
			t.Fatalf("Incorrect number of members found in DB:\nExpected: 3\nGot: %d", len(gotMembers))
		}

		for i, wantedMember := range wantedMembers {
			if wantedMember.Name != gotMembers[i].Name {
				t.Fatal("Incorrect member saved in memory")
			}
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != addReqString {
			t.Fatal("Add request not logged")
		}
	})

	t.Run("List groups", func(t *testing.T) {
		listReqString := BotName + " list"

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": listReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting list request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "Here are all of the usable group names:") {
			t.Fatalf("Creation request unsuccessful: %q", string(respMsg))
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != listReqString {
			t.Fatal("List request not logged")
		}
	})

	t.Run("List specific group", func(t *testing.T) {
		listGroupReqString := BotName + " list " + groupName

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": listGroupReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting list request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "Here are details for ") {
			t.Fatalf("List group request unsuccessful: %q", string(respMsg))
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != listGroupReqString {
			t.Fatal("List group request not logged")
		}
	})

	t.Run("Set group to private", func(t *testing.T) {
		restrictReqString := BotName + " restrict " + groupName

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": restrictReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting restrict request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "I've set \\\""+groupName+"\\\" to be private,") {
			t.Fatalf("Restrict request unsuccessful: %q", string(respMsg))
		}

		senderRoomGID := jsonGDS["message"].(map[string]interface{})["space"].(map[string]interface{})["name"]
		if !tGroup.IsPrivate || tGroup.PrivacyRoomID != senderRoomGID {
			t.Fatal("Group not marked as restricted in memory")
		}

		var gotGroup Group
		db.Raw("SELECT * from groups WHERE id = ?", tGroup.ID).Scan(&gotGroup)

		if !gotGroup.IsPrivate || gotGroup.PrivacyRoomID != senderRoomGID {
			t.Fatal("Group not marked as restricted in db")
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != restrictReqString {
			t.Fatal("Restrict request not logged")
		}
	})

	t.Run("Notify group", func(t *testing.T) {
		msgTmpl := "This is a message for <replace> please do the needful"

		notifyReqString := strings.Replace(msgTmpl, "<replace>", BotName+" "+groupName, 1)

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": notifyReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting notify request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		expectedMessage := fmt.Sprintf("%s said:\\n\\n%s",
			jsonGDS["message"].(map[string]interface{})["sender"].(map[string]interface{})["displayName"],
			strings.Replace(
				msgTmpl,
				"<replace>",
				"\\u003c"+tGroup.Members[0].GID+
					"\\u003e \\u003c"+tGroup.Members[1].GID+
					"\\u003e \\u003c"+tGroup.Members[2].GID+"\\u003e ",
				1,
			),
		)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), expectedMessage) {
			t.Fatalf("Notify request unsuccessful: %q", string(respMsg))
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != notifyReqString {
			t.Fatal("Notify request not logged")
		}
	})

	t.Run("Remove member from group", func(t *testing.T) {
		memberToRemove := tGroup.Members[0]

		removeReqString := BotName + " remove " + groupName + " @" + memberToRemove.Name

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": removeReqString,
			"annotations": []map[string]interface{}{
				{
					"type": "USER_MENTION",
					"userMention": map[string]interface{}{
						"user": map[string]interface{}{
							"name":        memberToRemove.GID,
							"displayName": memberToRemove.Name,
							"type":        "HUMAN",
						},
						"type": "MENTION",
					},
				},
			},
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting memberRemove request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "I've removed the user") {
			t.Fatalf("MemberRemove request unsuccessful: %q", string(respMsg))
		}

		if len(tGroup.Members) != 2 {
			t.Fatalf("Incorrect number of group members in memory:\nWanted: 2\nGot: %d",
				len(tGroup.Members),
			)
		}

		for _, member := range tGroup.Members {
			if memberToRemove.Name == member.Name {
				t.Fatal("Removed member found in memory")
			}
		}

		var gotMember Member
		db.Raw("SELECT * FROM members WHERE deleted_at IS NULL AND name = ?", memberToRemove.Name).Scan(&gotMember)

		if gotMember.Name == memberToRemove.Name {
			t.Fatal("Did not remove member from DB")
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != removeReqString {
			t.Fatal("Remove request not logged")
		}
	})

	t.Run("Disband group", func(t *testing.T) {
		disbandReqString := BotName + " disband " + groupName

		jsonGDS := getPostTemplateWith(map[string]interface{}{
			"text": disbandReqString,
		})

		jsonData, err := json.Marshal(&jsonGDS)
		if err != nil {
			t.Fatal(err)
		}

		postData := bytes.NewBuffer(jsonData)

		resp, err := http.Post(server.URL, contentType, postData)
		if err != nil {
			t.Fatalf("Error posting disband request: %q", err.Error())
		}

		time.Sleep(time.Second / 5)

		respMsg, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(respMsg), "Group \\\""+groupName+"\\\" has been deleted,") {
			t.Fatalf("Disband request unsuccessful: %q", string(respMsg))
		}

		if _, exist := TestGroupMap[saveName]; exist {
			t.Fatal("Deleted Group still exists in memory")
		}

		var gotGroup Group
		db.Model(&Group{}).Where(&Group{Name: groupName}).Take(&gotGroup)

		if !reflect.DeepEqual(Group{}, gotGroup) {
			t.Fatal("Deleted group found in DB")
		}

		var gotLogEntry NotifyLog
		db.Model(&NotifyLog{}).Last(&gotLogEntry)

		if gotLogEntry.Message != disbandReqString {
			t.Fatal("Create request not logged")
		}
	})
}

func getPostTemplateWith(addition map[string]interface{}) map[string]interface{} {
	pt := getPostTemplate()

	for k, v := range addition {
		pt["message"].(map[string]interface{})[k] = v
	}

	return pt
}

func getPostTemplate() map[string]interface{} {
	return map[string]interface{}{
		"type":      "MESSAGE",
		"eventTime": "2020-06-11T00:41:07.457887Z",
		"message": map[string]interface{}{
			"name": "spaces/spaceName/messages/messageID",
			"sender": map[string]interface{}{
				"name":        "users/1234567890",
				"displayName": "Sender Name",
				"avatarUrl":   "",
				"email":       "e@mail.com",
				"type":        "HUMAN",
				"domainId":    "domainId",
			},
			"createTime": "2020-06-11T00:41:07.457887Z",
			"text":       "text",
			"space": map[string]interface{}{
				"name":        "spaces/123435567798",
				"type":        "ROOM",
				"displayName": "Test Room",
				"threaded":    true,
			},
		},
		"space": map[string]interface{}{
			"name":        "spaces/123435567798",
			"type":        "ROOM",
			"displayName": "Test Room",
			"threaded":    true,
		},
	}
}

//In message the mentions are referred to as `annotations`
/*

	"annotations": []map[string]interface{}{{
		"type":       "USER_MENTION",
		"startIndex": 0,
		"length":     20,
		"userMention": map[string]interface{}{
			"user": map[string]interface{}{
				"name":        "users/108301888196217761156",
				"displayName": "DevelopmentHGNotify",
				"avatarUrl":   "",
				"type":        "BOT",
			},
			"type": "MENTION",
		},
	}, {
		"type":       "USER_MENTION",
		"startIndex": 30,
		"length":     13,
		"userMention": map[string]interface{}{
			"user": map[string]interface{}{
				"name":        "users/105372275486099500707",
				"displayName": "Robert Rabel",
				"avatarUrl":   "",
				"email":       "robert.rabel@endurance.com",
				"type":        "HUMAN",
				"domainId":    "49oqyxm",
			},
			"type": "MENTION",
		},
	}},
*/
