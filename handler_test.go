package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestHandler(t *testing.T) {
	Logger.Active(false)
	contentType := "application/json"

	server := httptest.NewServer(getRequestHandler(GroupMap{}))

	t.Run("Correctly responds when added to room", func(t *testing.T) {
		data := bytes.NewBuffer([]byte(`{
  "type": "ADDED_TO_SPACE",
  "eventTime": "2020-06-11T00:41:07.457887Z",
  "message": {
    "name": "spaces/spaceName/messages/messageID",
    "sender": {
      "name": "users/1234567890",
      "displayName": "Person Name",
      "avatarUrl": "",
      "email": "e@mail.com",
      "type": "HUMAN",
      "domainId": "domainId"
    },
    "createTime": "2020-06-11T00:41:07.457887Z",
    "text": "@DevelopmentHGNotify",
    "space": {
      "name": "spaces/spaceName",
      "type": "ROOM",
      "displayName": "Test Room",
      "threaded": true
    }
  }
}`))

		resp, err := http.Post(server.URL, contentType, data)
		if err != nil {
			t.Fatalf("Error posting data: %q", err.Error())
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Incorrect status code\nWanted: 200\nGot: %d", resp.StatusCode)
		}

		respText, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if !strings.Contains(string(respText), "Here's what I'm about:") {
			t.Fatalf("Incorrect message returned\nGot %q", string(respText))
		}

	})

	t.Run("Correctly responds when messaged", func(t *testing.T) {
		data := bytes.NewBuffer([]byte(`{
  "type": "MESSAGE",
  "eventTime": "2020-06-11T00:41:07.457887Z",
  "message": {
    "name": "spaces/spaceName/messages/messageID",
    "sender": {
      "name": "users/1234567890",
      "displayName": "Person Name",
      "avatarUrl": "",
      "email": "e@mail.com",
      "type": "HUMAN",
      "domainId": "domainId"
    },
    "createTime": "2020-06-11T00:41:07.457887Z",
    "text": "@DevelopmentHGNotify list",
    "space": {
      "name": "spaces/spaceName",
      "type": "ROOM",
      "displayName": "Test Room",
      "threaded": true
    }
  }
}`))

		resp, err := http.Post(server.URL, contentType, data)
		if err != nil {
			t.Fatalf("Error posting data: %q", err.Error())
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Incorrect status code\nWanted: 200\nGot: %d", resp.StatusCode)
		}

		respText, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if !strings.Contains(string(respText), "no groups to show") {
			t.Fatalf("Incorrect message returned\nGot %q", string(respText))
		}

	})

}
