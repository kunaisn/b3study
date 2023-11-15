package main

import (
	"encoding/json"
	"fmt"
	"math"
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
	cosSim, err := os.ReadFile("../toPy/entropy.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var writtenData EventsDataJSON
	err = json.Unmarshal(cosSim, &writtenData)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	sort.Slice(writtenData.Events, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", writtenData.Events[i].Date)
		dateJ, _ := time.Parse("2006-01-02", writtenData.Events[j].Date)
		return dateI.After(dateJ)
	})
	var gotTopics Topics
	gotTopics.Result = Classification(writtenData, 0.35)
	err = writeJson(gotTopics)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	count := make(map[int]int)
	for _, v := range gotTopics.Result {
		n := len(v.DocIds)
		if v, found := count[n]; found {
			count[n] = v + 1
		} else {
			count[n] = 1
		}
	}
	fmt.Println(count)
}

// 類似度によるトピックの分類
// a（0〜1で指定）をコサイン類似度の閾値とする
func Classification(d EventsDataJSON, a float64) []Topic {
	var topics []Topic
	for _, event := range d.Events {
		date, err := time.Parse("2006-01-02", event.Date)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return []Topic{}
		}
		inData := Document{Date: date, Entropy: event.Entropy}
		highestSim := struct {
			idx    int
			cosSim float64
		}{idx: -1}
		for j, topic := range topics {
			cosSim := CulcCosSim(event.TfIdf, topic.CenterGravity)
			if a < cosSim && highestSim.cosSim < cosSim && increaseEntropy(topic, inData) {
				highestSim = struct {
					idx    int
					cosSim float64
				}{idx: j, cosSim: cosSim}
			}
		}
		if highestSim.idx == -1 {
			topics = append(topics, Topic{DocIds: make(map[int]Document), CenterGravity: event.TfIdf})
			topics[len(topics)-1].DocIds[event.Id] = inData
		} else {
			topics[highestSim.idx].DocIds[event.Id] = inData
			topics[highestSim.idx].CulcCenterOfGravity(event.TfIdf)
		}
	}
	return topics
}

// トピックにdを追加することによって、時間と共にエントロピーが増大することに違反しないか確認する
func increaseEntropy(topic Topic, inData Document) bool {
	dateEntropyAvg := make(map[time.Time]float64)
	for _, doc := range topic.DocIds {
		if v, found := dateEntropyAvg[doc.Date]; found {
			dateEntropyAvg[doc.Date] = (v + doc.Entropy) / 2
		} else {
			dateEntropyAvg[doc.Date] = doc.Entropy
		}
	}
	for date, entropy := range dateEntropyAvg {
		if date.Before(inData.Date) && entropy > inData.Entropy {
			return false
		}
		if inData.Date.Before(date) && entropy < inData.Entropy {
			return false
		}
	}
	return true
}

// 二つのベクトルのコサイン類似度を計算
func CulcCosSim(n, m map[string]float64) float64 {
	length := func(n map[string]float64) float64 {
		var l float64
		for _, vN := range n {
			l += vN * vN
		}
		l = math.Sqrt(l)
		return l
	}
	var d float64
	for kN, vN := range n {
		if vM, found := m[kN]; found {
			d += vN * vM
		}
	}
	l := length(n) * length(m)
	return d / l
}

// 二つのベクトルの重心を計算
func (t *Topic) CulcCenterOfGravity(n map[string]float64) {
	for kN, vN := range n {
		if vT, found := t.CenterGravity[kN]; found {
			t.CenterGravity[kN] = vT + vN
		} else {
			t.CenterGravity[kN] = vN
		}
	}
	for k, v := range t.CenterGravity {
		t.CenterGravity[k] = v / 2
	}
}

func writeJson(d Topics) error {
	output, err := json.Marshal(&d)
	if err != nil {
		return err
	}
	f, err := os.Create("topics.json")
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(output)
	return nil
}
