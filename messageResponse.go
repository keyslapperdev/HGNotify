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

	Room struct {
		GID  string `json:"name"`
		Name string `json:"displayName"`
	} `json:"space"`
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
			option != "disband" &&
			option != "restrict" &&
			option != "list" &&
			option != "usage" &&
			option != "help" {
			msg = fmt.Sprintf("Invalid option received. I'm not sure what to do about %q.", tempArgs[1])
			ok = false
		} else {
			args["action"] = option

			if nArgs < 3 {
				args["groupName"] = ""
			} else {
				args["groupName"] = tempArgs[2]
			}
		}
	} else {
		args["action"] = "notify"
		gi := 100000

		for i, v := range tempArgs {
			if v == BOTNAME {
				gi = i
			}

			if i == gi+1 {
				args["groupName"] = v
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
		if args["groupName"] == "" {
			retMsg = fmt.Sprintf("My apologies, you need to pass a group name to be able to create the group. ```%s```", usage("create"))
		} else {
			retMsg = Groups.Create(args["groupName"], msgObj)
		}
	case "disband":
		if args["groupName"] == "" {
			retMsg = fmt.Sprintf("You'd need to pass a group name for me to delete it. ```%s```", usage("disband"))
		} else {
			retMsg = Groups.Disband(args["groupName"], msgObj)
		}
	case "add":
		retMsg = Groups.AddMembers(args["groupName"], msgObj)
	case "remove":
		retMsg = Groups.RemoveMembers(args["groupName"], msgObj)
	case "restrict":
		if args["groupName"] == "" {
			retMsg = fmt.Sprintf("You'd need to pass a group name to toggle it's privacy settings. ```%s```", usage("restrict"))
		} else {
			retMsg = Groups.Restrict(args["groupName"], msgObj)
		}
	case "notify":
		retMsg = Groups.Notify(args["groupName"], msgObj)
	case "list":
		retMsg = Groups.List(args["groupName"], msgObj)
	case "usage":
		retMsg = usage("")
	case "help":
		retMsg = usage("")
	default:
		retMsg = "Unknown action? Shouldn't have gotten here tho... reach out for someone to check my innards. You should seriously never see this message."
	}

	return
}
