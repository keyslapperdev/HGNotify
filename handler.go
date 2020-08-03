package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getRequestHandler(Groups GroupMgr, Scheduler ScheduleMgr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			msgObj   messageResponse
			jsonReq  []byte
			jsonResp []byte
			e        error
		)

		jsonReq, e = ioutil.ReadAll(r.Body)
		checkError(e)

		e = json.Unmarshal(jsonReq, &msgObj)
		checkError(e)

		switch msgObj.Type {
		case "ADDED_TO_SPACE":
			resp := map[string]string{
				"text": "Thank you for inviting me! Here's what I'm about:" + usage(""),
			}
			jsonResp, e = json.Marshal(resp)

		case "MESSAGE":
			//Log every usage of hgnotify to the db.
			go Logger.CreateLogEntry(msgObj)

			var msg string

			args, errMsg, okay := msgObj.ParseArgs(Groups)
			if okay {
				msg = inspectMessage(Groups, Scheduler, msgObj, args)
			} else {
				msg = errMsg
			}

			resp := map[string]string{
				"text": msg,
			}
			jsonResp, e = json.Marshal(resp)

		default:
			//Not too sure of any message type that's not Added or Message, there is removed
			//But that one doesn't allow messages to be sent, sooooooooo?
			resp := map[string]string{
				"text": "Oh, ummm! I'm not exactly sure what happened, or what type of request this is. But here's what I was made to do, if it helps." + usage(""),
			}
			jsonResp, e = json.Marshal(resp)
		}

		fmt.Fprintf(w, "%s", string(jsonResp))
	}
}

func ReadinessCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}
