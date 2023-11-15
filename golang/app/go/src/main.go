package main

import (
	"fmt"
	"log"
	"main/apis/sqldb"
	"main/apis/util"
	"main/apis/wiki"
	"os"
	"sync"
	"time"
)

// 実行コマンド：go run main.go

func main() {
	Get()
}

func Get() {
	// 開始日
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// 終了日（半開区間）[start, end)
	end := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	for next := start; next.Before(end); {
		err := GetDocuments(next)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		y, m, d := next.Date()
		fmt.Printf("done %s\n", fmt.Sprintf("%v_%v_%v", y, m, d))
		next = next.AddDate(0, 0, 1)
	}
}

func GetDocuments(t time.Time) error {
	events, err := wiki.GetEventData(t)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	if len(events) == 0 {
		fmt.Printf("%s, events are nothing or already registered\n", t.Format("2006-01-02"))
		return err
	}
	wiki, news := getWikiAndNewsData(events)
	err = queryDB(events, wiki, news)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// getWikiAndNewsDataはwiki内記事とnews記事の取得をする。
func getWikiAndNewsData(events []sqldb.Event) ([][]sqldb.WikiArt, [][]sqldb.NewsArt) {
	var wg sync.WaitGroup
	wg.Add(2)
	ch1 := make(chan [][]sqldb.WikiArt)
	ch2 := make(chan [][]sqldb.NewsArt)
	go wiki.GetAllWikiArticle(events, &wg, ch1)
	go wiki.GetAllNewsArticle(events, &wg, ch2)
	wiki := <-ch1
	news := <-ch2
	wg.Wait()
	return wiki, news
}

func queryDB(events []sqldb.Event, wiki [][]sqldb.WikiArt, news [][]sqldb.NewsArt) error {
	db, err := sqldb.ConnectDB()
	if err != nil {
		return err
	}
	log.Println("successfully connected db")
	defer db.Close()
	wikiId, err := sqldb.SlelctAllIdAndUrlWikiArticle(db)
	if err != nil {
		return err
	}
	newsId, err := sqldb.SelectAllIdAndUrlNewsArticle(db)
	if err != nil {
		return err
	}
	for i := 0; i < len(events); i++ {
		var wikiUrl []string
		for _, v := range wiki[i] {
			wikiUrl = append(wikiUrl, v.WikiSourceUrl)
		}
		events[i].EntitiesId = util.SearchId(wikiId, wikiUrl)
		var newsUrl []string
		for _, v := range news[i] {
			newsUrl = append(newsUrl, v.NewsSourceUrl)
		}
		events[i].NewsSourceUrlId = util.SearchId(newsId, newsUrl)
		if err != nil {
			return err
		}
		err = sqldb.InsertWikiEvent(db, events[i])
		if err != nil {
			return err
		}
	}
	log.Println("finished to add,", len(events), "events")
	if len(events) > 0 {
		err = sqldb.InsertDate(db, events[0].Date)
		if err != nil {
			return err
		}
	}
	return nil
}
