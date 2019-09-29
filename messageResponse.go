package main

import (
	"fmt"
	"strings"
)

type User struct {
	Name string `json:"displayName"`
	GID  string `json:"name"`
	Type string `json:"type"`
}

type messageResponse struct {
	Message struct {
		Sender struct {
			GID  string `json:"name"`
			Name string `json:"displayName"`
		} `json:"sender"`

		Mentions []struct {
			Called struct {
				User `json:"user"`
			} `json:"userMention"`
			Type string `json:"type"`
		} `json:"annotations"`

		Text string `json:"text"`
	} `json:"message"`
}

func (mr messageResponse) parseArgs() (args Arguments, msg string, ok bool) {
	tempArgs := strings.Fields(mr.Message.Text)
	nArgs := len(tempArgs)
	if nArgs < 2 {
		msg = BOTNAME + " seems to have been called with no params. Just a heads up."
		ok = false

		return
	}

	args = make(Arguments)
	ok = true
	msg = ""

	if tempArgs[0] == BOTNAME {
		option := strings.ToLower(tempArgs[1])

		if option != "create" &&
			option != "add" &&
			option != "remove" &&
			option != "delete" &&
			option != "list" &&
			option != "usage" {
			msg = fmt.Sprintf("Invalid option received, %q. Full Message: %q", tempArgs[1], mr.Message.Text)
			ok = false
		} else {
			args["action"] = option

			if option != "usage" {
				if nArgs < 3 {
					args["groupName"] = ""
				} else {
					args["groupName"] = tempArgs[2]
				}
			}
		}
	} else {
		args["action"] = "notify"
		pu := 100000

		for i, v := range tempArgs {
			if v == BOTNAME {
				pu = i
			}

			if i == pu+1 {
				args["groupName"], ok = Groups.CheckGroup(v)
				if !ok {
					msg = fmt.Sprintf("Group %q doesn't seem to exist yet, try initializing it with \"@HGNotify create %s\".", v, v)
				}
				break
			}
		}
	}

	return
}

func inspectMessage(msgObj messageResponse) (retMsg, errMsg string, ok bool) {
	ok = true

	args, msg, okay := msgObj.parseArgs()
	if !okay {
		errMsg = msg
		ok = false
		return
	}

	switch args["action"] {
	case "create":
		retMsg = Groups.Create(args["groupName"], msgObj)
	case "delete":
		retMsg = Groups.Delete(args["groupName"])
	case "add":
		retMsg = Groups.AddMembers(args["groupName"], msgObj)
	case "remove":
		retMsg = Groups.RemoveMembers(args["groupName"], msgObj)
	case "notify":
		retMsg = "Received Call to " + args["action"]
	case "list":
		retMsg = Groups.List(args["groupName"])
	case "usage":
		retMsg = usage()
	default:
		retMsg = "Unknown action? Shouldn't have gotten here tho... reach out for someone to check my innards. You should seriously never see this message."
	}

	return
}
