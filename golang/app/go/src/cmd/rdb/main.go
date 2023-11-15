package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"main/apis/sqldb"
	"main/apis/util"
	"main/apis/wiki"
	"os"
	"strings"
	"time"
)

func main() {
	newsAry, err := sqldb.GetEmptyDataOfNewsArts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("get newsAry")
	for _, v := range newsAry {
		fmt.Println("req diffbot: ", v.NewsSourceUrl[7:])
		dbData, err := wiki.Diffbot(v.NewsSourceUrl)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		data := parseNewsArt(dbData, v.NewsSourceUrl)
		data.Id = v.Id
		db, err := sqldb.ConnectDB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		err = sqldb.UpdateDiffbotData(db, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		db.Close()
	}
}

func reGetDiffbotCutTail() {
	newsAry, err := sqldb.GetEmptyDataOfNewsArts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("get newsAry")
	for _, v := range newsAry {
		fmt.Println(v.NewsSourceUrl)
		path, err := trucAmp(v.NewsSourceUrl)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Println("req diffbot")
		dbData, err := wiki.Diffbot(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		data := parseNewsArt(dbData, v.NewsSourceUrl)
		data.Id = v.Id
		db, err := sqldb.ConnectDB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer db.Close()
		err = sqldb.UpdateDiffbotData(db, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}

func manualInput() {
	newsAry, err := sqldb.GetEmptyDataOfNewsArts()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	writeURL(newsAry)
	for _, v := range newsAry {
		fmt.Println(v.NewsSourceUrl)
		news, err := inputNewsArt()
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		var data sqldb.NewsArt
		if len(news.Objects) == 0 {
			data = sqldb.NewsArt{Timestamp: "2008-01-02", NewsSourceUrl: v.NewsSourceUrl}
		} else {
			data = parseNewsArt(news, v.NewsSourceUrl)
		}
		db, err := sqldb.ConnectDB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer db.Close()
		err = sqldb.UpdateDiffbotData(db, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "id: ", v.Id)
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}

func trucAmp(path string) (string, error) {
	if strings.Contains(path[len(path)-5:], "amp") {
		idx := strings.Index(path[len(path)-5:], "amp")
		return path[:len(path)-6+idx], nil
	}
	return "", errors.New("not found \"amp\"")
}

func inputNewsArt() (wiki.DiffbotData, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var body []byte
	for scanner.Scan() {
		get := string(scanner.Bytes())
		if get == "fin:" {
			break
		}
		if get == "no:" {
			return wiki.DiffbotData{}, nil
		}
		body = append(body, scanner.Bytes()...)
	}
	var data wiki.DiffbotData
	err := json.Unmarshal(body, &data)
	if err != nil {
		return wiki.DiffbotData{}, err
	}
	return data, nil
}

func parseNewsArt(news wiki.DiffbotData, path string) sqldb.NewsArt {
	date := news.Objects[0].Date
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", date)
	if err != nil {
		t = time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC)
	}
	var cat []string
	for _, v := range news.Objects[0].Categories {
		cat = append(cat, v.Name)
	}
	data := sqldb.NewsArt{
		Timestamp:       t.Format("2006-01-02"),
		SiteName:        news.Objects[0].SiteName,
		PublisherRegion: news.Objects[0].PublisherRegion,
		Category:        util.JoinStringByTab(cat),
		Title:           news.Objects[0].Title,
		HumanLanguage:   news.Objects[0].HumanLanguage,
		Text:            news.Objects[0].Text,
		NewsSourceUrl:   path,
	}
	return data
}

func writeURL(news []sqldb.NewsArt) {
	os.Remove("tmp.txt")
	f, err := os.Create("tmp.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, v := range news {
		str := []byte(v.NewsSourceUrl)
		_, err := f.Write(str)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		_, err = f.Write([]byte("\n"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}
