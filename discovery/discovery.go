package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var agents = []string{
	"localhost:5000",
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received a request from:", r.RemoteAddr)
	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		asJSON := make(map[string]interface{})
		err = json.Unmarshal(body, &asJSON)
		fmt.Println("Body:", asJSON)
	}
	// Now send a result:
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

func register(myAddress string) (myUUID uuid.UUID) {
	for {
		myUUID = uuid.NewV4()
		body := bytes.Buffer{}
		fmt.Fprintf(&body, `[[{"/servers/%s": { "op": "set", "new": "%s" }},
												  {"/servers/%s": { "oldEmpty": true }}]]`,
			myUUID, myAddress, myUUID)
		resp, err := http.Post("http://"+agents[0]+"/_api/agency/write",
			"application/json", &body)
		if err == nil {
			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Request to agency was good, statusCode:",
				resp.StatusCode, " body:", string(respBody))
			return
		}
		fmt.Println("Error in http request: ", err)
		time.Sleep(1000000000) // 1 second
	}
}

func unregister(myUUID uuid.UUID) {
	body := bytes.Buffer{}
	fmt.Fprintf(&body, `[[{"/servers/%s": { "op": "delete" }}]]`, myUUID)
	http.Post("http://"+agents[0]+"/_api/agency/write", "application/json", &body)
}

func getColleagues(myUUID uuid.UUID) (result []map[string]map[string]string) {
	body := bytes.Buffer{}
	body.WriteString(`[["/servers"]]`)
	resp, err := http.Post("http://"+agents[0]+"/_api/agency/read",
		"application/json", &body)
	if resp == nil || resp.Body == nil || err != nil {
		fmt.Println("Read request to agency was bad: ", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respBody, &result)
	return
}

func pingColleagues(myUUID uuid.UUID,
	colleagues *[]map[string]map[string]string) {
	servers := (*colleagues)[0]["servers"]
	for k, v := range servers {
		if k != myUUID.String() {
			body := bytes.Buffer{}
			fmt.Fprintf(&body, `{"myId": "%s"}`, myUUID.String())
			resp, err := http.Post("http://"+v+"/hello", "application/json", &body)
			if err != nil {
				fmt.Println("Post to colleague went wrong: ", v, " ", err)
			} else {
				defer resp.Body.Close()
				fmt.Println("Reached colleague: ", v)
			}
		}
	}
}

func main() {
	// Create server:
	http.HandleFunc("/hello", hello)
	myAddress := os.Args[1]
	go http.ListenAndServe(myAddress, nil)

	// Make a uuid and register with that uuid:
	myUUID := register(myAddress)
	defer unregister(myUUID)
	fmt.Println("My ID is ", myUUID)

	for i := 0; i < 10; i++ {
		time.Sleep(1000000000)
		// Get all the others:
		colleagues := getColleagues(myUUID)
		pingColleagues(myUUID, &colleagues)
	}
}
