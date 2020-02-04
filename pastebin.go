package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Link to pastebin url
const PasteBinURL = "http://hgfix.net/paste/api/create"

// How long the paste will live. 1 day should
//be enough
const TTL = "1440"

func getUsageWithLink(usage string) string {
	usage = strings.Replace(usage, "```", "\n", -1)

	opts := make(url.Values)

	opts["text"] = append(opts["text"], usage)
	opts["private"] = append(opts["private"], "0")
	opts["expire"] = append(opts["expire"], TTL)
	opts["lang"] = append(opts["lang"], "text")
	opts["name"] = append(opts["name"], BotName)

	resp, err := http.PostForm(PasteBinURL, opts)
	if err != nil {
		return fmt.Sprintf("Error retrieving pastebin usage link. Please run `%v help` for full usage text.\n\nError received: %v", BotName, err)
	}

	usageLink, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error getting link from response body. Please run `%v help` for full usage text.\n\nError: %v`", BotName, err)
	}

	return fmt.Sprintf("Here is a link with the full usage statement %v\nUse `%v help` to see the full statement in the chat.", strings.Replace(string(usageLink), "view", "view/raw", 1), BotName)
}
