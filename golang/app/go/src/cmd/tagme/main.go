package main

import (
	"encoding/json"
	"fmt"
	"io"
	"main/apis/sqldb"
	"main/apis/util"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type TagMeData struct {
	Test        string `json:"test"`
	Annotations []struct {
		Spot            string  `json:"spot"`
		Start           int     `json:"start"`
		LinkProbability float64 `json:"link_probability"`
		Rho             float64 `json:"rho"`
		End             int     `json:"end"`
		ID              int     `json:"id"`
		Title           string  `json:"title"`
	} `json:"annotations"`
	Time      int    `json:"time"`
	API       string `json:"api"`
	Lang      string `json:"lang"`
	Timestamp string `json:"timestamp"`
}

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

func main() {
	// イベントデータを抽出
	d, err := getEventDataAllText("2022-01-01", "2022-12-31")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	// TagMeでエンティティを抽出
	err = d.SetEntitiesFromTagMe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	// pythonに送るデータを書き込む
	err = writeSendPythonData(d)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func writeSendPythonData(d EventsDataJSON) error {
	jsonData, err := json.Marshal(&d)
	if err != nil {
		return err
	}
	jsonDataUrl := "/go/src/go/data/EventsDataJSON.json"
	f, err := os.Create(jsonDataUrl)
	if err != nil {
		return err
	}
	f.Write(jsonData)
	f.Close()
	return nil
}

func getEventDataAllText(start, end string) (EventsDataJSON, error) {
	db, err := sqldb.ConnectDB()
	if err != nil {
		return EventsDataJSON{}, err
	}
	defer db.Close()
	eventData, err := sqldb.SelectEvents(db, start, end)
	if err != nil {
		return EventsDataJSON{}, err
	}
	var d EventsDataJSON
	for _, v := range eventData {
		var e sendDataElem
		e.Id = v.Id
		e.Date = v.Date
		e.Text = util.TruncTailBracketsText(v.Text)
		d.Events = append(d.Events, e)
	}
	return d, nil
}

func (d *EventsDataJSON) SetEntitiesFromTagMe() error {
	nowProg := 0
	fmt.Println("started to get Entieties")
	for i, e := range d.Events {
		ents, err := getEntities(util.CutRemoveWords(e.Text))
		if err != nil {
			return err
		}
		for j, s := range ents {
			ents[j] = util.CutRemoveWords(s)
		}
		d.Events[i].Entities = ents
		tmpProg := int((float64(i) / float64(len(d.Events))) * 100)
		if nowProg != tmpProg {
			nowProg = tmpProg
			fmt.Printf("\rfinished: %2d%%", nowProg)
		}
	}
	fmt.Println("\rfinished: 100%")

	return nil
}

// getEntitiesは与えられた文字列のエンティティを抽出します。
// 関連度が0.1を下回ったものは除外されます。
func getEntities(text string) ([]string, error) {
	text = util.TruncTailBracketsText(text)
	data, err := queryTagMe(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	entities := make([]string, 0)
	for _, v := range data.Annotations {
		if v.LinkProbability < 0.1 {
			continue
		}
		entities = append(entities, v.Spot)
	}
	return entities, nil
}

// queryTagMeはTagMeのAPIを使用した結果を得ます。
func queryTagMe(text string) (TagMeData, error) {
	values := url.Values{}
	values.Set("text", text)
	values.Set("gcube-token", util.TagMeAPIKey)
	req, err := http.NewRequest(
		"POST",
		"https://tagme.d4science.org/tagme/tag",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return TagMeData{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return TagMeData{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, res.StatusCode)
		body, _ := io.ReadAll(res.Body)
		fmt.Fprintln(os.Stderr, string(body))
		return TagMeData{}, err
	}
	body, _ := io.ReadAll(res.Body)
	var d TagMeData
	err = json.Unmarshal(body, &d)
	if err != nil {
		return TagMeData{}, err
	}
	return d, nil
}
