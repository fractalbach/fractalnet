package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type JustEventType struct {
	EventType string
}
type Location struct {
	X int
	Y int
}
type MessageBomba struct {
	Value    uint8
	Location Location
}

var (
	exampleEmpty = []byte(``)
	exampleBomba = []byte(`{"EventType": "LaBomba","Value": 1, "Location":{ "X": 0, "Y": 47}}`)
)

func neater(jsonblob []byte) {

	t := new(JustEventType)
	err := json.Unmarshal(jsonblob, t)
	if err != nil {
		log.Println(err)
		return
	}

	switch t.EventType {
	case "":
		log.Println("Invalid Message: Missing EventType Field.")

	case "LaBomba":
		o := new(MessageBomba)
		json.Unmarshal(jsonblob, &o)
		fmt.Println("Wooooo!!", o)

	default:
		log.Println("Invalid Message: Unknown EventType.")
	}
	return
}

func main() {
	neater(exampleBomba)
}
