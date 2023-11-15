package main

import (
	"encoding/json"
	"fmt"
	"main/apis/util"
	"math"
	"net/http"
	"os"
	"strings"
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

const URL = "http://172.27.0.3:8050" // 立て直すたびに変わるので要確認

func main() {
	// pythonでTF-IDFを計算
	gotData, err := sendPython()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	// エントロピーを計算
	for i, v := range gotData.Events {
		gotData.Events[i].Entropy = culcEntropy(v.Text, v.Entities)
	}
	err = writeJson(gotData)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func sendPython() (EventsDataJSON, error) {
	res, err := http.Get(URL)
	if err != nil {
		return EventsDataJSON{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return EventsDataJSON{}, fmt.Errorf("bad request: %d", res.StatusCode)
	}
	cosSim, err := os.ReadFile("/go/src/go/data/CosSim.json")
	if err != nil {
		return EventsDataJSON{}, err
	}
	var writtenData EventsDataJSON
	err = json.Unmarshal(cosSim, &writtenData)
	if err != nil {
		return EventsDataJSON{}, err
	}
	return writtenData, nil
}

func culcEntropy(texts string, entities []string) float64 {
	texts = strings.ToLower(util.CutRemoveWords(texts))
	newEntities := make([]string, len(entities))
	for i, s := range entities {
		newEntities[i] = strings.ToLower(s)
	}
	entitiesCount := make(map[string]int)
	for _, entity := range newEntities {
		count := 0
		for i := 0; i < len(texts)-len(entity)+1; i++ {
			if texts[i:i+len(entity)] == entity {
				count++
			}
		}
		entitiesCount[entity] = count * len(strings.Fields(entity))
	}
	textsWordsLen := len(strings.Fields(texts))
	sum := 0.0
	for _, v := range entitiesCount {
		p := float64(v) / float64(textsWordsLen)
		sum += p * math.Log2(p)
		if math.IsNaN(sum) {
			fmt.Println(texts)
			fmt.Println(entitiesCount)
			fmt.Println("")
		}
	}
	sum *= -1
	return sum
}

func writeJson(d EventsDataJSON) error {
	output, err := json.Marshal(&d)
	if err != nil {
		return err
	}
	f, err := os.Create("entropy.json")
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(output)
	return nil
}
