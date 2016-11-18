package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"../AgencyComm"
)

var agents []string
var theAgency AgencyComm.Agency

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

type obj map[string]interface{}

func register(myAddress string) (myUUID uuid.UUID) {
	for {
		myUUID = uuid.NewV4()
		w := AgencyComm.WriteTransaction{
			     Update:       obj{ "/servers/" + myUUID.String():
					                    obj{ "op": "set", "new": myAddress }},
					 Precondition: obj{ "/servers/" + myUUID.String():
					                    obj{ "oldEmpty": true } } }
    res, e := theAgency.SendWrite(w)
    if e == nil && res[0] != 0 {
			fmt.Println("Request to agency was good, registered.")
			return
		}
		fmt.Println("Could not register: ", e, res)
		time.Sleep(1000000000) // 1 second
	}
}

func unregister(myUUID uuid.UUID) {
	w := AgencyComm.WriteTransaction{
		Update: obj{ "/servers/" + myUUID.String(): obj{ "op": "delete" } } }
  theAgency.SendWrite(w)
}

func getColleagues(myUUID uuid.UUID) (result []map[string]interface{}) {
	r := AgencyComm.ReadTransaction{Paths: []string{"/servers"}}
	res, e := theAgency.SendRead(r)
	if e != nil {
		fmt.Println("Read request to agency was bad: ", e)
		return
	}
	return res
}

func pingColleagues(myUUID uuid.UUID,
	colleagues *[]map[string]interface{}) {
	servers := (*colleagues)[0]["servers"].(map[string]interface{})
	for k, v := range servers {
		if k != myUUID.String() {
			name := v.(string)
			body := bytes.Buffer{}
			fmt.Fprintf(&body, `{"myId": "%s"}`, myUUID.String())
			resp, err := http.Post("http://"+name+"/hello", "application/json", &body)
			if err != nil {
				fmt.Println("Post to colleague went wrong: ", name, " ", err)
			} else {
				defer resp.Body.Close()
				fmt.Println("Reached colleague: ", name)
			}
		}
	}
}

func main() {
	// Build agency object:
	agents = make([]string, len(os.Args) - 2)
	for j := 2; j < len(os.Args); j++ {
		agents[j-2] = os.Args[j]
	}
	theAgency = AgencyComm.Agency{Endpoints: agents}

	// Create server:
	http.HandleFunc("/hello", hello)
	myAddress := os.Args[1]
	pos := strings.LastIndex(myAddress, ":")
	myPort := myAddress[pos+1:]
	go http.ListenAndServe("0.0.0.0:"+myPort, nil)

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
