package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//Initializing global variables
var (
	Config   = loadConfig("secret/config.yml")
	dbConfig = loadDBConfig("secret/dbconfig.yml")

	Groups = make(GroupList)
	Logger = startDBLogger(dbConfig)
)

//Setting up general configurations for usage of the bot
var (
	CertFile    = Config.CertFile
	CertKeyFile = Config.CertKeyFile

	port = Config.Port

	BotName  = Config.BotName
	MasterID = Config.MasterID
)

type (
	//GenericJSON holds type for general JSON objects
	GenericJSON map[string]interface{}

	//Arguments is generic type for passing arguments as a map
	//between functions. I feel this is kinda an artifact of
	//learing perl as my first language, but it sure does make
	//things a bit more clear.
	Arguments map[string]string
)

func main() {
	Logger.SetupTables()
	Logger.GetGroupsFromDB(Groups)

	fmt.Println("Running!! on port " + port)

	http.HandleFunc("/", theHandler)
	e := http.ListenAndServeTLS(port, CertFile, CertKeyFile, nil)
	checkError(e)
}

func theHandler(w http.ResponseWriter, r *http.Request) {
	var (
		payload  GenericJSON
		msgObj   messageResponse
		jsonReq  []byte
		jsonResp []byte
		e        error
	)

	jsonReq, e = ioutil.ReadAll(r.Body)
	checkError(e)

	e = json.Unmarshal(jsonReq, &payload)
	checkError(e)

	switch payload["type"] {
	case "ADDED_TO_SPACE":
		resp := map[string]string{
			"text": "Thank you for inviting me! Here's what I'm about:" + usage(""),
		}
		jsonResp, e = json.Marshal(resp)

	case "MESSAGE":
		e = json.Unmarshal(jsonReq, &msgObj)
		checkError(e)

		//Log every usage of hgnotify to the db.
		go Logger.CreateLogEntry(msgObj)

		var msg string
		resMsg, errMsg, ok := inspectMessage(msgObj)

		if !ok {
			msg = errMsg
		} else {
			msg = resMsg
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

func usage(option string) string {
	options := make(map[string]string)

	options["create"] = `
create groupName [mentions]
    Create a group containing mentioned members. While I'm not sure why you would, you can initialize an empty group. Add "self" to the list of mentions to add yourself.`

	options["disband"] = `
disband groupName
    Delete a group. CAUTION: This can be done to a group containing members. I'd recommend only using delete when necessary.`

	options["add"] = `
add groupName mentions
    Add mentioned members to the specified GroupName. This can only be used for groups that already exist. If you intend to create a new group use create. Add "self" to the list of mentions to add yourself.`

	options["remove"] = `
remove groupName mentions
    Remove mentioned members from the specified GroupName. Add "self" to the list of mentions to remove yourself.`

	options["restrict"] = `
restrict groupName
    Toggles group privacy, this disallows any interaction with the group outside the room it was restricted in. (Default: Public)`

	options["usage"] = `
help|usage
    Reprint's this message`

	options["list"] = `
list [groupName]
  If used with no groupName, you will receive a list of all groups you can currently use. This will not show any private group that you do not have access to. If used with a groupName, you will see more information about the group specified.`

	options["notify"] = `
groupName
  Replaces groupName with mentions for the group members along with the following/surrounding/leading message.`

	usageShort := "`@HGNotify [options] [GroupName] [mentions...]`"

	if option == "usageShort" {
		return usageShort
	} else if option != "" {
		return options[option]
	}

	summary := "I was created to @ groups of people by using user created groups, since gchat doesn't seem to already have this functionality."

	limitation := "Due to a chat bot in Google Chat not being able to add users to a room, if you use a group in a room where the group members are not already in, they will not be mentioned, you'll just see their user id. It's gravely disapointing, but the result of a limitation within googles chat system."

	examples := `
Mentions: "HEY! @HGNotify HG6, great job on that new product!" would turn into "HEY! @Alexander Wilcots @Robert Rabel @Robert Stone @James Frotten @Cai Black @Taylor Mitchell @Srimathy Thyagarajan, great job on that new product"

Creating a group: "@HGNotify create HG1 @Brandon Husbands"

Making a group private: "@HGNotify restrict HG1"

Adding a group member: "@HGNotify add HG1 @Taylor Mitchell"

Removing a group member: "@HGNotify remove HG6 @Robert Stone"

Delete a group: "@HGNotify disband Umbrella"`

	notes := `
- Group Names are case insensative.
- Group Names can contain letters, numbers, underscores, and dashes maximum length is 40 characters
- When managing groups, "@HGNotify" must be the first thing in the messages
- When notifying a group the text "@HGNotify GroupName" will be replaced with the members of the group. Just a heads up, so be sure to place that where you'd like it to appear.

- The bot is manged by mentioning people. If someone is unable to be mentioned, to get them removed from a group, you can reach out to the maintainer.
- Any problems, comments, or suggestions please send me a message in gchat or email me at alexander.wilcots@endurance.com`

	return fmt.Sprintf(" Usage ``%s`` Summary ```%s``` *LIMITATION* Please read ```%s``` Examples ```%s``` Options ```%s``` Notes ```%s```",
		usageShort,
		summary,
		limitation,
		examples,
		fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
			options["create"],
			options["add"],
			options["remove"],
			options["disband"],
			options["restrict"],
			options["list"],
			options["notify"],
			options["usage"],
		),
		notes,
	)
}
