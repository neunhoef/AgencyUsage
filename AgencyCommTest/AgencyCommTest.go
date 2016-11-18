package main

import (
	"fmt"
	"encoding/json"
	"time"
	"../AgencyComm"
)

type obj map[string]interface{}

func main() {
	a := AgencyComm.Agency{Endpoints: []string{
		"localhost:5000",
		"localhost:5001",
		"localhost:5002"}}
	w := AgencyComm.WriteTransaction{
		      Update: obj{ "/a": obj{"op":"set", "new": 12}}}
	ww := AgencyComm.WriteTransaction{
		      Update: obj{ "/b": 12 },
	        Precondition: obj{ "/b": obj{ "oldEmpty": true } }}
	r := AgencyComm.ReadTransaction{Paths: []string{"/a", "/b/c"}}
	fmt.Println("Hello", a, w, r, ww)
	s, e := json.Marshal(w)
	fmt.Println("w:", e, string(s))
	s, e = json.Marshal(r)
	fmt.Println("t:", e, string(s))
	s, e = json.Marshal(ww)
	fmt.Println("ww:", e, string(s))
	var res []int64
	res, e = a.SendWrite(w)
	if e != nil {
		fmt.Println("Error:", e)
	} else {
		fmt.Println("Result:", res)
	}
	time.Sleep(1000000000)
	res, e = a.SendWrite(w, ww)
	if e != nil {
		fmt.Println("Error:", e)
	} else {
		fmt.Println("Result:", res)
	}
	time.Sleep(1000000000)
	list := []AgencyComm.WriteTransaction{w, ww}
	res, e = a.SendWrite(list...)
	if e != nil {
		fmt.Println("Error:", e)
	} else {
		fmt.Println("Result:", res)
	}
	time.Sleep(1000000000)
	var res2 []map[string]interface{}
	res2, e = a.SendRead(r)
	if e != nil {
		fmt.Println("Error:", e)
	} else {
		fmt.Println("Result:", res2)
	}
}


