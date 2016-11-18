package AgencyComm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Agency struct {
	Endpoints []string
	Current   int
}

type WriteTransaction struct {
	Update       map[string]interface{} // the update part
	Precondition map[string]interface{} // the precondition part
}

func (w WriteTransaction) MarshalJSON() (res []byte, err error) {
	buf := bytes.Buffer{}
	buf.Write([]byte("["))
  b, _ := json.Marshal(w.Update)
	buf.Write(b)
	if len(w.Precondition) != 0 {
		buf.Write([]byte(","))
	  b, _ = json.Marshal(w.Precondition)
		buf.Write(b)
  }
	buf.Write([]byte("]"))
	res = buf.Bytes()
	err = nil
	return
}

type ReadTransaction struct {
	Paths []string
}

func (w ReadTransaction) MarshalJSON() (res []byte, err error) {
  return json.Marshal(w.Paths)
}

func (a *Agency) Add(endpoint string) int {
	for i := 0; i < len(a.Endpoints); i++ {
    if a.Endpoints[i] == endpoint {
			return i;
		}
	}
	a.Endpoints = append(a.Endpoints, endpoint)
	return len(a.Endpoints) - 1
}

func (a *Agency) sendWithFailover(path string, trx ...interface{}) (resp *http.Response, err error) {
	for count := 0; count < 100; count++ {   // some retries
		buf := bytes.Buffer{}
		b, _ := json.Marshal(trx)
		buf.Write(b)

		resp, err = http.Post("http://" + a.Endpoints[a.Current] + path,
												  "application/json", &buf)
		if err != nil {
			a.Current = (a.Current + 1) % len(a.Endpoints)
			fmt.Println("Error in request, trying next (", a.Current, ")")
			continue    // try another server immediately
		}

		// Otherwise we have a StatusCode
		if resp.StatusCode == 307 {
			loc := resp.Header["Location"][0]
			fmt.Println("Follow redirect to", loc);
			var theUrl *url.URL
			theUrl, err = url.Parse(loc)
			if err != nil {
				return
			}
		  host := theUrl.Host
			// add endpoint if it is not yet there:
			a.Current = a.Add(host)
			continue    // try again immediately
		}

		if resp.StatusCode >= 500 {
			fmt.Println("HTTP status code >= 500, retrying another agent after 0.5s")
			a.Current = (a.Current + 1) % len(a.Endpoints)
		  time.Sleep(500000000)   // wait for half a second
			continue
		}

		return    // Simply tell the caller what has happened
	}
	err = errors.New("sendWithFailover lost patience")
	return
}

func (a *Agency) SendWrite(trx ...WriteTransaction) (indexes []int64, err error) {
	forgotten := make([]interface{}, len(trx))
	for i := 0; i < len(trx); i++ {
		forgotten[i] = trx[i]
	}
	var resp *http.Response
	resp, err = a.sendWithFailover("/_api/agency/write", forgotten...)
	if err != nil {
		fmt.Println("Error:", err)
		indexes = make([]int64, len(trx))
		return
	}
	defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK &&
	   resp.StatusCode != http.StatusPreconditionFailed {
		fmt.Println("StatusCode bad:", resp.StatusCode)
		indexes = make([]int64, len(trx))
    err = errors.New("bad HTTP response code")
		return;
	}
	body, parseErr := ioutil.ReadAll(resp.Body)
	if parseErr == nil {
		asJSON := make(map[string][]int64, 1)
		parseErr = json.Unmarshal(body, &asJSON)
		indexes = asJSON["results"]
	  err = nil
	} else {
		indexes = make([]int64, len(trx))
		err = errors.New("could not parse response body")
	}
	return
}

func (a *Agency) SendRead(trx ...ReadTransaction) (result []map[string]interface{}, err error) {
	forgotten := make([]interface{}, len(trx))
	for i := 0; i < len(trx); i++ {
		forgotten[i] = trx[i]
	}
	var resp *http.Response
	resp, err = a.sendWithFailover("/_api/agency/read", forgotten...)
	if err != nil {
		fmt.Println("Error:", err)
		result = make([]map[string]interface{}, len(trx))
		return
	}
  if resp.StatusCode != http.StatusOK {
		fmt.Println("StatusCode bad:", resp.StatusCode)
		result = make([]map[string]interface{}, len(trx))
    err = errors.New("bad HTTP response code")
		return;
	}
	body, parseErr := ioutil.ReadAll(resp.Body)
	if parseErr == nil {
		asJSON := make([]map[string]interface{}, len(trx))
		parseErr = json.Unmarshal(body, &asJSON)
		result = asJSON
	  err = nil
	} else {
		fmt.Println("Error: Could not parse response body:", string(body))
		result = make([]map[string]interface{}, len(trx))
		err = errors.New("could not parse response body")
	}
	return
}

