package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

type EventsDataJSON struct {
	Events []sendDataElem `json:"events"`
}

type sendDataElem struct {
	Id       int                `json:"id"`
	Date     string             `json:"date"`
	Text     string             `json:"text"`
	Entities []string           `json:"entities"`
	TfIdf    map[string]float64 `json:"tf_idf"`
	Entropy  float64            `json:"entropy"`
}

type Topics struct {
	Result []Topic `json:"result"`
}

type Topic struct {
	DocIds        map[int]Document   `json:"doc_ids"`
	CenterGravity map[string]float64 `json:"center_gravity"`
}

type Document struct {
	Date    time.Time `json:"date"`
	Entropy float64   `json:"entropy"`
}

func main() {
	var event EventsDataJSON
	var topics Topics
	eventByte, err := os.ReadFile("../toPy/entropy.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = json.Unmarshal(eventByte, &event)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	topicsByte, err := os.ReadFile("../topics/topics.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = json.Unmarshal(topicsByte, &topics)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	eventText := make(map[int]string)
	for _, v := range event.Events {
		eventText[v.Id] = v.Date + " - " + v.Text
	}
	for _, topic := range topics.Result {
		if len(topic.DocIds) < 5 {
			continue
		}
		events := make([]string, 0)
		for id := range topic.DocIds {
			events = append(events, eventText[id])
		}
		sort.Slice(events, func(i, j int) bool {
			dateI, _ := time.Parse("2006-01-02", events[i][:10])
			dateJ, _ := time.Parse("2006-01-02", events[j][:10])
			return dateI.Before(dateJ)
		})
		fmt.Println("・Topic")
		for _, e := range events {
			fmt.Println(e)
		}
		fmt.Println("")
	}
}
