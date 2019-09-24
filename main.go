package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	logBreak string = "--------------------------------\n"
)

func bigHandle(w http.ResponseWriter, r *http.Request) {
	var (
		payload  map[string]interface{}
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
			"text": "Thank you for including me!",
		}
		jsonResp, e = json.Marshal(resp)

	case "MESSAGE":
		//stub
	case "test":
		resp := map[string]string{
			"text": "I've received the message",
		}
		jsonResp, e = json.Marshal(resp)

	default:
		//stub
	}

	fmt.Printf("Request: %v\nRespose: %s\n%s", string(jsonReq), string(jsonResp), logBreak)
	fmt.Fprintf(w, "%s", string(jsonResp))
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	const (
		port = ":8443"
		cert = "/home/z/ssl/certs/gapi_thezspot_net_d2713_9b85d_1577145599_6429c0f6539a8947b31a35ed9a430a7e.crt"
		key  = "/home/z/ssl/keys/d2713_9b85d_3927f691549410111f93434afd1f37a7.key"
	)

	fmt.Println("Running!! on port " + port)

	http.HandleFunc("/", bigHandle)

	e := http.ListenAndServeTLS(port, cert, key, nil)
	checkError(e)
}
