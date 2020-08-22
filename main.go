package main

import (
	"fmt"
	"net/http"
)

//Initializing global variables
var (
	Config = initConfig()

	Logger = startDBLogger(initDBConfig())
)

//Setting up general configurations for usage of the bot
var (
	CertFile    = Config.CertFile
	CertKeyFile = Config.CertKeyFile

	BotName  = Config.BotName
	MasterID = Config.MasterID

	baseRoute = "/"
	port      = ":8000"
)

//Arguments is generic type for passing arguments as a map
//between functions. I feel this is kinda an artifact of
//learing perl as my first language, but it sure does make
//things a bit more clear.
type Arguments map[string]string

func main() {
	Groups := make(GroupMap)
	Schedules := make(ScheduleMap)

	Logger.SetupTables()
	Logger.GetGroupsFromDB(Groups)
	Logger.GetSchedulesFromDB(Schedules)

	fmt.Println("Running!! on port " + port)

	http.HandleFunc(baseRoute, getRequestHandler(Groups, Schedules))
	http.HandleFunc(baseRoute+"readiness/", ReadinessCheck())

	var err error

	if Config.UseSSL == "true" {
		err = http.ListenAndServeTLS(port, CertFile, CertKeyFile, nil)
	} else {
		err = http.ListenAndServe(port, nil)
	}

	checkError(err)
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
