package wiki

import (
	"fmt"
	"log"
	"main/apis/sqldb"
	"main/apis/util"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GetEventDataはイベントデータを取得する。
// WikipediaのCurrent_eventsからスクレイピングするが、
// すでにDBに登録されている場合は取得を中止する。
//
// （todo: すでにDBに登録されている場合にDB内のデータを戻す）
func GetEventData(t time.Time) ([]sqldb.Event, error) {
	// DBに接続
	db, err := sqldb.ConnectDB()
	if err != nil {
		return make([]sqldb.Event, 0), err
	}
	defer db.Close()
	// すでにスクレイピングをしていたか確認する
	found, err := sqldb.SelectDate(db, t.Format("2006-01-02"))
	if err != nil {
		// DB関係のエラー
		return []sqldb.Event{}, err
	} else if found {
		// すでにスクレイピング済みのため中断
		return []sqldb.Event{}, nil
	}
	// 日付からURLを生成
	y, m, d := t.Date()
	urlTail := fmt.Sprintf("%v_%v", m, y)
	url := "https://en.wikipedia.org/wiki/Portal:Current_events/" + urlTail
	// http接続して全HTML文を取得
	doc, err := getHTML(url)
	if err != nil {
		// HTML文の取得失敗
		return nil, fmt.Errorf("failed get html: %v", err)
	}
	// 該当の日付のブロックのみを切り出す
	section := doc.Find("div#" + fmt.Sprintf("%v_%v_%v", y, m, d))
	// データの抽出処理をする
	return exCurrentEvent(section, t.Format("2006-01-02")), nil
}

// getHTMLはhttp接続を行って、goqueryで扱えるようにしたHTML文を受け取る。
func getHTML(url string) (*goquery.Document, error) {
	// urlにアクセスして、レスポンスを受け取る
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get: url[%v]\n %v", url, err)
	}
	defer res.Body.Close()
	// ステータスコードを確認
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// HTMLを読み込む
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// exCurrentEventは与えられた日付のイベントを抽出する。
func exCurrentEvent(slct *goquery.Selection, date string) []sqldb.Event {
	// イベントが格納されている二つ目のブロックに移動する
	content := slct.Children().Next().Children()
	// 最終的に得られるイベントデータを格納
	events := make([]sqldb.Event, 0)
	// 再帰するたびに、タグと紐づいたwiki内記事が溜まる
	stackTags := make([]string, 0)
	stackTagsEntities := make([]string, 0)
	// 探索中のカテゴリを保存
	nowCategory := ""
	// <---深さ優先探索でイベントデータを取得する--->
	// 再帰するための関数の宣言、実行はもっと下
	var dfs func(i int, slct *goquery.Selection)
	dfs = func(i int, slct *goquery.Selection) {
		// li要素が渡されるので、ul要素を抽出する
		ulSlct := slct.ChildrenFiltered("ul")
		if ulSlct.Size() != 0 {
			tagSlct := slct.Children().Not("ul")
			stackTags = append(stackTags, tagSlct.Text())
			addEntitiesCount := 0
			tagSlct.Each(func(i int, slct *goquery.Selection) {
				val, found := slct.Attr("href")
				if found {
					stackTagsEntities = append(stackTagsEntities, val)
					addEntitiesCount++
				}
			})
			liSlct := ulSlct.ChildrenFiltered("li")
			liSlct.Each(dfs)
			stackTags = stackTags[:len(stackTags)-1]
			stackTagsEntities = stackTagsEntities[:len(stackTagsEntities)-addEntitiesCount]
		} else {
			// ul要素がなければ最下層まで潜り切っているため、渡されたliがイベントの本文である
			// ニュース記事のURLを格納
			stackNewsSourceUrl := make([]string, 0)
			// 追加したwiki内記事の数を格納
			addEntitiesCount := 0
			aTagSlct := slct.ChildrenFiltered("a")
			aTagSlct.Each(func(i int, slct *goquery.Selection) {
				url, foundHref := slct.Attr("href")
				_, foundRel := slct.Attr("rel")
				if foundHref && foundRel {
					// rel属性が付いている場合は、ニュース記事である
					stackNewsSourceUrl = append(stackNewsSourceUrl, url)
				} else if foundHref {
					// 上記以外は全てwiki内記事である
					stackTagsEntities = append(stackTagsEntities, url)
					addEntitiesCount++
				}
			})
			// イベントデータを保存
			events = append(events, sqldb.Event{
				Date:          date,
				Category:      nowCategory,
				Tags:          util.CopyStrAry(stackTags),
				Text:          slct.Text(),
				Entities:      util.CopyStrAry(stackTagsEntities),
				NewsSourceUrl: util.CopyStrAry(stackNewsSourceUrl),
			})
			// タグは他のイベントでも参照するため、このブロックで追加したwiki内記事を削除する
			stackTagsEntities = stackTagsEntities[:len(stackTagsEntities)-addEntitiesCount]
		}
	}
	// カテゴリごとに探索を行う
	content.Each(func(i int, slct *goquery.Selection) {
		// pタグの場合、カテゴリを変更
		if slct.Is("p") {
			nowCategory = slct.Text()
			// 末尾に改行コードが付いている場合は取り除く
			if nowCategory[len(nowCategory)-1] == '\n' {
				nowCategory = nowCategory[:len(nowCategory)-1]
			}
		}
		// ulタグの場合、イベントを探索
		if slct.Is("ul") {
			// ul要素内に複数のli要素がタグごとにあるため、個別に処理する。
			liSlct := slct.ChildrenFiltered("li")
			// 深さ優先探索を開始
			liSlct.Each(dfs)
		}
	})
	return events
}

// GetAllWikiArticleはeventsに含まれる全てのwiki内記事を調べ、戻す
func GetAllWikiArticle(events []sqldb.Event, wg *sync.WaitGroup, ch chan [][]sqldb.WikiArt) {
	defer wg.Done()
	defer close(ch)
	var wikiArtAry [][]sqldb.WikiArt
	for i, event := range events {
		wikiArtAry = append(wikiArtAry, []sqldb.WikiArt{})
		for _, entitie := range event.Entities {
			art, err := GetWikiArticle(entitie)
			if err != nil {
				wikiArtAry[i] = append(wikiArtAry[i], sqldb.WikiArt{})
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			wikiArtAry[i] = append(wikiArtAry[i], art)
		}
	}
	ch <- wikiArtAry
}

// wiki内の記事を抽出して戻す
func GetWikiArticle(path string) (sqldb.WikiArt, error) {
	// wiki内記事がすでにDBに登録されているか確認する
	db, err := sqldb.ConnectDB()
	if err != nil {
		return sqldb.WikiArt{}, err
	}
	defer db.Close()
	get, err := sqldb.SelectWikiArticle(db, path)
	// DB関連のエラー
	if err != nil {
		return sqldb.WikiArt{}, fmt.Errorf("error %s [getWikiArticle()]: %s ", path, err)
	}
	// DBにすでに登録されている
	if get.Id != -1 {
		return get, nil
	}
	log.Println("started to get a wiki art, " + path)
	url := "https://en.wikipedia.org" + path
	doc, err := getHTML(url)
	if err != nil {
		return sqldb.WikiArt{}, err
	}
	/*// 深さ優先探索でstyle要素を除去する
	var dfs func(i int, slct *goquery.Selection)
	dfs = func(i int, slct *goquery.Selection) {
		slct.Find("style").Each(func(i int, styleSlct *goquery.Selection) {
			styleSlct.Remove()
		})
		next := slct.Children()
		if next.Size() != 0 {
			next.Each(dfs)
		}
	}
	textSec := doc.Find("div.mw-parser-output")
	textSec.Each(dfs)*/
	// body要素全て
	allText, err := doc.Find("html").Html()
	if err != nil {
		return sqldb.WikiArt{}, err
	}
	// ノーマルカテゴリを抽出する
	catSec := doc.Find("div#mw-normal-catlinks > ul").Children()
	category := make([]string, 0)
	catSec.Each(func(i int, slct *goquery.Selection) {
		category = append(category, slct.Text())
	})
	art := sqldb.WikiArt{
		WikiSourceUrl: path,
		Text:          allText,
		WikiCategory:  util.JoinStringByTab(category),
	}
	// データベースに登録
	err = sqldb.InsertWikiArticle(db, art)
	if err != nil {
		return sqldb.WikiArt{}, err
	}
	// 抽出の終了
	log.Println("finished to get a article, " + path)
	// Wikipediaに負荷をかけすぎないように調整（Diffbotの処理時間より長くならない程度）
	time.Sleep(2 * time.Second)
	return art, nil
}

// eventsに含まれる全てのnews記事を調べ、戻す
func GetAllNewsArticle(events []sqldb.Event, wg *sync.WaitGroup, ch chan [][]sqldb.NewsArt) {
	defer wg.Done()
	defer close(ch)
	var newsArtAry [][]sqldb.NewsArt
	for i, event := range events {
		newsArtAry = append(newsArtAry, []sqldb.NewsArt{})
		for _, url := range event.NewsSourceUrl {
			art, err := getNewsArticle(url)
			if err != nil {
				newsArtAry[i] = append(newsArtAry[i], art)
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			newsArtAry[i] = append(newsArtAry[i], art)
		}
	}
	ch <- newsArtAry
}

// ニュース記事を分析して結果を戻す、Diffbotのリクエスト待ちが数秒かかる
func getNewsArticle(url string) (sqldb.NewsArt, error) {
	emptyVal := sqldb.NewsArt{
		Timestamp:     "2006-01-02",
		NewsSourceUrl: url,
	}
	db, err := sqldb.ConnectDB()
	if err != nil {
		return emptyVal, err
	}
	defer db.Close()
	get, err := sqldb.SelectNewsArticle(db, url)
	// DB関連のエラー
	if err != nil {
		return emptyVal, fmt.Errorf("error %s [getNewsArticle()]: %s ", url[7:], err)
	}
	// DBにすでに登録されている
	if get.Id != -1 {
		return get, nil
	}
	art := emptyVal
	sqldb.InsertNewsArticle(db, art)
	return emptyVal, err
	/*
		// 最後にデータを登録する
		art := emptyVal
		log.Println("started to get a news art, " + url[7:])
		news, err := Diffbot(url)
		if err != nil {
			sqldb.InsertNewsArticle(db, art)
			return emptyVal, err
		}
		// Objectsの0番目は存在する。また、現状の仕様[2023/10/28]では1以上の添え字でデータが渡されることはない。
		date := news.Objects[0].Date
		t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", date)
		if err != nil {
			t = time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC)
		}
		var cat []string
		for _, v := range news.Objects[0].Categories {
			cat = append(cat, v.Name)
		}
		log.Println("finished to get a article, " + url[7:])
		art = sqldb.NewsArt{
			Timestamp:       t.Format("2006-01-02"),
			SiteName:        news.Objects[0].SiteName,
			PublisherRegion: news.Objects[0].PublisherRegion,
			Category:        util.JoinStringByTab(cat),
			Title:           news.Objects[0].Title,
			HumanLanguage:   news.Objects[0].HumanLanguage,
			Text:            news.Objects[0].Text,
			NewsSourceUrl:   url,
		}
		err = sqldb.InsertNewsArticle(db, art)
		if err != nil {
			return emptyVal, err
		}
		return art, nil
	*/
}
