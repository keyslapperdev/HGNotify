package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

const (
	PORT string = ":8443"
	CERT string = "/home/z/ssl/certs/gapi_thezspot_net_d2713_9b85d_1577145599_6429c0f6539a8947b31a35ed9a430a7e.crt"
	KEY  string = "/home/z/ssl/keys/d2713_9b85d_3927f691549410111f93434afd1f37a7.key"

	BOTNAME  string = "@HGNotify"
	LOGBREAK string = "--------------------------------\n"
)

type (
	GenericJSON map[string]interface{}
	Arguments   map[string]string
)

var Groups = make(GroupList)

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

func (mr messageResponse) ParseArgs() (args Arguments, msg string, ok bool) {
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
				args["groupName"], msg, ok = Groups.Validate(v)
				break
			}
		}
	}

	return
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
			"text": "Thank you for inviting me! Here's how I work" + usage(),
		}
		jsonResp, e = json.Marshal(resp)

	case "MESSAGE":
		e = json.Unmarshal(jsonReq, &msgObj)
		checkError(e)

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
		resp := map[string]string{
			"text": "Oh, ummm! I'm not exactly sure what happened, or what type of request this is. But here's what I was made to do, if it helps." + usage(),
		}
		jsonResp, e = json.Marshal(resp)
	}

	//describe("Request: %v\nResponse: %v\n%s", string(jsonReq), string(jsonResp), LOGBREAK)
	fmt.Fprintf(w, "%s", string(jsonResp))
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func describe(msg string, v ...interface{}) {
	spew.Printf(msg, v...)
}

func inspectMessage(msgObj messageResponse) (retMsg, errMsg string, ok bool) {
	ok = true

	args, msg, okay := msgObj.ParseArgs()
	if !okay {
		errMsg = msg
		ok = false
		return
	}

	switch args["action"] {
	case "create":
		retMsg = Groups.Create(args["groupName"], msgObj)
	case "delete":
		retMsg = "Received Call to " + args["action"]
	case "add":
		retMsg = "Received Call to " + args["action"]
	case "remove":
		retMsg = "Received Call to " + args["action"]
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

func usage() string {
	msg := `
Usage: @HGNotify [options] [GroupName] [mentions...]
  Summary
    I was created to @ groups of people by using user created groups, since gchat doesn't seem to already have this functionality.

  Examples
    Mentions: "HEY! @HGNotify HG6, great job on that new product!" would turn into "HEY! @Alexander Wilcots @Robert Rabel @Robert Stone @James Frotten @Cai Black @Taylor Mitchell @Srimathy Thyagarajan, great job on that new product"

    Creating a group: "@HGNotify create HG1 @Brandon Husbands"

    Adding a group member: "@HGNotify add HG1 @Taylor Mitchell"

    Removing a group member "@HGNotify remove HG6 @Robert Stone"

    Delete a group: "@HGNotify delete Umbrella"

  Options
    create groupName mentions
      Create a group containing mentioned members. While I'm not sure why you would,
      you can initialize an empty group.

    delete groupName
      Delete a group. CAUTION: This can be done to a group containing members. I'd
      recommend only using delete when necessary.

    add groupName mentions
      Add mentioned members to the specified GroupName. This can only be used for
      groups that already exist. If you intend to create a new group use create.

    remove groupName mentions
      Remove mentioned members from the specified GroupName.

    usage
      Reprint's this message

  Notes
    - Group Names are case insensative.
    - Group Names can contain letters, numbers, underscores, and dashes
    - When managing groups, "@HGNotify" must be the first thing in the messages
    - Any problems please contact me at ----
`
	return fmt.Sprintf(" ```%s```", msg)
}

func main() {
	fmt.Println("Running!! on port " + PORT)

	http.HandleFunc("/", theHandler)

	e := http.ListenAndServeTLS(PORT, CERT, KEY, nil)
	checkError(e)
}
