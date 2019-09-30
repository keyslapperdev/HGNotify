package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func main() {
	fmt.Println("Running!! on port " + PORT)

	http.HandleFunc("/", theHandler)

	e := http.ListenAndServeTLS(PORT, CERT, KEY, nil)
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
			"text": "Thank you for inviting me! Here's what I'm about:" + usage(),
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

	describe("Request: %v\nResponse: %v\n%s", string(jsonReq), string(jsonResp), LOGBREAK)
	fmt.Fprintf(w, "%s", string(jsonResp))
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
    - When notifying a group the text "@HGNotify GroupName" will be replaced with the members of the group. Just a heads up, so be sure to place that where you'd like it to appear.

    - Any problems please contact me at ----
`
	return fmt.Sprintf(" ```%s```", msg)
}
