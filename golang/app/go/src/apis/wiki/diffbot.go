package wiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"main/apis/util"
	"net/http"
	"net/url"
)

// 受け取ったデータを保持する構造体
type DiffbotData struct {
	Objects []struct {
		Date            string `json:"date"`
		SiteName        string `json:"siteName"`
		Title           string `json:"title"`
		PublisherRegion string `json:"publisherRegion"`
		HumanLanguage   string `json:"humanLanguage"`
		Categories      []struct {
			Name string `json:"name"`
		} `json:"categories"`
		Text string `json:"text"`
	} `json:"objects"`
}

func Diffbot(path string) (DiffbotData, error) {
	encodingPath := url.QueryEscape(path)
	diffbot := "https://api.diffbot.com/v3/article?url=" + encodingPath
	diffbot += "&token=" + util.DiffbotAPIKey
	req, err := http.NewRequest("GET", diffbot, nil)
	if err != nil {
		return DiffbotData{}, err
	}
	req.Header.Add("accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return DiffbotData{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var d DiffbotData
	err = json.Unmarshal(body, &d)
	if err != nil {
		return DiffbotData{}, err
	}
	if len(d.Objects) < 1 {
		str := fmt.Sprintf("could not get a diffbot's responses, %s: ", string(body))
		return DiffbotData{}, errors.New(str)
	}
	return d, nil
}
