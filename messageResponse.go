package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
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
	Message message `json:"message"`
	Room    space   `json:"space"`
	Time    string  `json:"eventTime"`
	Type    string  `json:"type"`

	FromMaster bool
}

type message struct {
	Sender   User         `json:"sender"`
	Mentions []annotation `json:"annotations"`
	Thread   thread       `json:"thread"`
	Text     string       `json:"text"`
}

type annotation struct {
	Called userMention `json:"userMention"`
	Type   string      `json:"type"`
}

type thread struct {
	Name string `json:"name"`
}

type userMention struct {
	User `json:"user"`
}

type space struct {
	GID  string `json:"name"`
	Type string `json:"type"`
}

//parseArgs is a method used to take the string passed through the api and make
//sense of it for the bot.
func (mr *messageResponse) ParseArgs(Groups GroupMgr) (args Arguments, msg string, ok bool) {
	mr.FromMaster = false
	//The admin of the bot can be changed via configs, but they are defined by
	//the id google gives them, incase their name changes, and how they reach out.
	//i.e. The bot will only recognize the admin if messaged via DM. an admin
	//shouldn't be doing admin things in front of the common folk.
	if mr.Room.Type == "DM" && mr.Message.Sender.GID == MasterID {
		mr.FromMaster = true
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
			option != "schedule" &&
			option != "usage" &&
			option != "help" {
			if Groups.IsGroup(tempArgs[1]) {
				args["action"] = "notify"
				args["groupName"] = tempArgs[1]
			} else {
				msg = fmt.Sprintf("Invalid option received. I'm not sure what to do about %q.", tempArgs[1])
				ok = false
			}
		} else {
			args["action"] = option
		}
	} else {
		args["action"] = "notify"
	}

	if args["action"] == "schedule" {
		err := parseScheduleArgs(Groups, tempArgs, &args)
		if err != nil {
			ok = false
			msg = err.Error()
			return
		}
	} else {
		if nArgs < 3 {
			args["groupName"] = ""
		} else {
			args["groupName"] = tempArgs[2]
		}
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
func inspectMessage(Groups GroupMgr, Scheduler ScheduleMgr, msgObj messageResponse, args Arguments) (msg string) {
	switch args["action"] {
	case "create":
		msg = Groups.Create(args["groupName"], args["self"], msgObj)

	case "disband":
		msg = Groups.Disband(args["groupName"], msgObj)

	case "add":
		msg = Groups.AddMembers(args["groupName"], args["self"], msgObj)

	case "remove":
		msg = Groups.RemoveMembers(args["groupName"], args["self"], msgObj)

	case "restrict":
		msg = Groups.Restrict(args["groupName"], msgObj)

	case "notify":
		msg = Groups.Notify(args["groupName"], msgObj)

	case "list":
		msg = Groups.List(args["groupName"], msgObj)

	case "syncgroup":
		msg = Groups.SyncGroupMembers(args["groupName"], msgObj)

	case "syncallgroups":
		msg = Groups.SyncAllGroups(msgObj)

	case "usageShort":
		msg = usage("usageShort")

	case "usage":
		msg = getUsageWithLink(usage(""))

	case "help":
		msg = usage("")

	case "schedule":
		switch args["subAction"] {
		case "onetime":
			msg = Scheduler.CreateOnetime(args, Groups, msgObj)
		case "list":
			msg = Scheduler.List(msgObj)
		}

	default:
		//All of the argument things should be taken care of by the time we get here,
		//BUT It's better to handle the exceptions than let them bite you.
		msg = "Unknown action? Shouldn't have gotten here tho... reach out for someone to check my innards. You should seriously never see this message."
	}

	return
}

func parseScheduleArgs(Groups GroupMgr, elems []string, args *Arguments) error {
	if len(elems) < 3 {
		return errors.New("Not enough arguments for schedule action")
	}

	(*args)["subAction"] = elems[2]

	switch (*args)["subAction"] {
	case "onetime":
		if len(elems) < 7 {
			return fmt.Errorf("Not enough arguments for schedule onetime action\n ```%s``` ", usage("schedule:onetime"))
		}

		ptrn := regexp.MustCompile(`^\w{3,20}$`)
		if !ptrn.Match([]byte(elems[3])) {
			return fmt.Errorf("Label be Alphanumeric between 3 and 20 characters\n ```%s```", usage("onetime"))
		}
		(*args)["label"] = elems[3]

		datetime, err := time.Parse(time.RFC3339, elems[4])
		if err != nil {
			return fmt.Errorf("Error parsing your time %q. Must be formatted in RFC3339 format, please try again", elems[4])
		}

		//the scheduled message has to be at least 1 hour out.
		if !datetime.After(time.Now().Add(time.Minute * 59)) {
			return fmt.Errorf("Scheduled message must be at least 1 hour away from now")
		}
		(*args)["dateTime"] = elems[4]

		if !Groups.IsGroup(elems[5]) {
			return fmt.Errorf("Specificed group %q not found", elems[5])
		}
		(*args)["groupName"] = elems[5]

		(*args)["message"] = strings.Join(elems[6:], " ")

	case "list":
		(*args)["subAction"] = "list"

	//case "recurring":
	//case "remove":
	default:
		return fmt.Errorf("Unknown schedule subaction %q called", elems[2])
	}

	return nil
}
