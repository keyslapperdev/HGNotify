package main

import (
	"fmt"
	"strings"
)

//User struct defines a user in the context of google's chat api
type User struct {
	Name string `json:"displayName"`
	GID  string `json:"name"`
	Type string `json:"type"`
}

//messageResponse contains pretty much everything important to this
//bot from gchat's payload
type messageResponse struct {
	Message struct {
		Sender struct {
			User
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
		Type string `json:"type"`
	} `json:"space"`

	Time     string `json:"eventTime"`
	IsMaster bool
}

//parseArgs is a method used to take the string passed through the api and make
//sense of it for the bot.
func (mr *messageResponse) parseArgs() (args Arguments, msg string, ok bool) {
	mr.IsMaster = false
	//The admin of the bot can be changed via configs, but they are defined by
	//the id google gives them, incase their name changes, and how they reach out.
	//i.e. The bot will only recognize the admin if messaged via DM. and admin
	//shouldn't be doing admin things in front of the common folk.
	if mr.Room.Type == "DM" && mr.Message.Sender.GID == MasterID {
		mr.IsMaster = true
		//This prepends the botname to the message so that the admin doesn't have to
		//@ the bot when DM-ing it. The conditional allows you to do either.
		if !strings.HasPrefix(mr.Message.Text, BotName) {
			mr.Message.Text = BotName + " " + mr.Message.Text
		}
	}

	tempArgs := strings.Fields(mr.Message.Text)
	nArgs := len(tempArgs)
	if nArgs < 2 {
		msg = BotName + " seems to have been called with no params. Just a heads up."
		ok = false

		return
	}

	args = make(Arguments)
	ok = true
	msg = ""

	if tempArgs[0] == BotName {
		option := strings.ToLower(tempArgs[1])

		//I'd like to find a cleaner way to triage the string, but this seems
		//to be the best as of now. I do worry, that this could get hard to maintain.
		if option != "create" &&
			option != "add" &&
			option != "remove" &&
			option != "disband" &&
			option != "restrict" &&
			option != "list" &&
			option != "syncgroup" &&
			option != "syncallgroups" &&
			option != "usage" &&
			option != "help" {
			if isGroup(tempArgs[1]) {
				args["action"] = "notify"
				args["groupName"] = tempArgs[1]
			} else {
				msg = fmt.Sprintf("Invalid option received. I'm not sure what to do about %q.", tempArgs[1])
				ok = false
			}
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
	}

	//Logic introduced for adding/removing yourself from a group
	if args["action"] == "add" || args["action"] == "remove" || args["action"] == "create" {
		for _, item := range tempArgs {
			if strings.ToLower(item) == "self" {
				args["self"] = "self"
				break
			}
		}
	}

	//This one is a bit jank, I'll admit (PR's Welcome :D). What's happening is gi
	//is a value set to be arbitrarily high, it stands for group index. The for loop
	//loops through the words given by the message text, separated by whitespace. When
	//When it identified the bot name, it sets gi. After gi is set, the next thing should
	//be the group name. So the loop just sets the next item to be the group name and
	//proceeds
	if args["action"] == "notify" {
		gi := 100000

		for i, v := range tempArgs {
			if v == BotName {
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

//inspectMessage method (maybe should be renamed) takes the parsed arguments
//then reacts accordingly.
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
			retMsg = Groups.Create(args["groupName"], args["self"], msgObj)
		}

	case "disband":
		if args["groupName"] == "" {
			retMsg = fmt.Sprintf("You'd need to pass a group name for me to delete it. ```%s```", usage("disband"))
		} else {
			retMsg = Groups.Disband(args["groupName"], msgObj)
		}

	case "add":
		retMsg = Groups.AddMembers(args["groupName"], args["self"], msgObj)

	case "remove":
		retMsg = Groups.RemoveMembers(args["groupName"], args["self"], msgObj)

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

	case "syncgroup":
		retMsg = Groups.SyncGroupMembers(args["groupName"], msgObj)

	case "syncallgroups":
		retMsg = Groups.SyncAllGroups(msgObj)

	case "usage":
		retMsg = usage("usageshort")

	case "help":
		retMsg = usage("")

	default:
		//All of the argument things should be taken care of by the time we get here,
		//BUT It's better to handle the exceptions than let them bite you.
		retMsg = "Unknown action? Shouldn't have gotten here tho... reach out for someone to check my innards. You should seriously never see this message."
	}

	return
}
