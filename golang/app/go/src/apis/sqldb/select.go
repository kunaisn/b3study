package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"main/apis/util"
	"strings"
)

// SelectWikiArticleはDBからwiki記事を検索し、抽出する。
// 見つからなかった場合、Idが-1になる。
func SelectWikiArticle(db *sql.DB, wikiSourceUrl string) (WikiArt, error) {
	stmt, err := db.Prepare("SELECT wiki_art_id, wiki_source_url FROM wiki_article WHERE wiki_source_url = ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SelectWikiArticle()]", err)
		return WikiArt{}, errors.New(str)
	}
	defer stmt.Close()
	wiki := WikiArt{Id: -1, WikiSourceUrl: wikiSourceUrl}
	err = stmt.QueryRow(wikiSourceUrl).Scan(
		&wiki.Id,
		&wiki.WikiSourceUrl,
	)
	// 通常のエラーは通さないが、データがないだけのエラーはそのまま通したい
	if err != sql.ErrNoRows && err != nil {
		return WikiArt{}, err
	}
	// データがなかった場合、エラーを戻さずにIdが-1の要素を戻す
	return wiki, nil
}

// SlelctAllIdAndUrlWikiArticleはDBに存在するwiki記事の「IDとURL」の組みを全て抽出する
func SlelctAllIdAndUrlWikiArticle(db *sql.DB) (map[string]int, error) {
	stmt, err := db.Prepare(
		"SELECT wiki_art_id, wiki_source_url FROM wiki_article")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SlelctAllIdAndUrlWikiArticle()]", err)
		return nil, errors.New(str)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	idAndUrl := make(map[string]int)
	for rows.Next() {
		var path string
		var id int
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		idAndUrl[path] = id
	}
	return idAndUrl, nil
}

// SelectNewsArticleはDBからnews記事を検索し、抽出する。
// 見つからなかった場合、Idが-1になる。
func SelectNewsArticle(db *sql.DB, newsSourceUrl string) (NewsArt, error) {
	stmt, err := db.Prepare("SELECT news_art_id, news_source_url FROM news_diffbot WHERE news_source_url = ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SelectNewsArticle()]: ", err)
		return NewsArt{}, errors.New(str)
	}
	defer stmt.Close()
	news := NewsArt{Id: -1}
	err = stmt.QueryRow(newsSourceUrl).Scan(
		&news.Id,
		&news.NewsSourceUrl,
	)
	// 行がなかった場合は、エラーを戻さずにIdが-1の要素を戻す
	if err != sql.ErrNoRows && err != nil {
		return NewsArt{}, err
	}
	return news, nil
}

// SelectAllIdAndUrlNewsArticleはDBに存在するnews記事の「IDとURL」の組みを全て抽出する
func SelectAllIdAndUrlNewsArticle(db *sql.DB) (map[string]int, error) {
	stmt, err := db.Prepare("SELECT news_art_id, news_source_url FROM news_diffbot")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SelectAllIdAndUrlNewsArticle()]", err)
		return nil, errors.New(str)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	idAndUrl := make(map[string]int)
	for rows.Next() {
		var path string
		var id int
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		idAndUrl[path] = id
	}
	return idAndUrl, nil
}

// SelectEventsは与えられた日付に起こったイベントを抽出する。[start, end]
func SelectEvents(db *sql.DB, start, end string) ([]Event, error) {
	stmt, err := db.Prepare("SELECT event_id, date, category, tags, text, entitie, news_source_url FROM wiki_event WHERE DATE(date) BETWEEN ? AND ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[selectEvents()]: ", err)
		return []Event{}, errors.New(str)
	}
	defer stmt.Close()
	var events []Event
	rows, err := stmt.Query(start, end)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e Event
		var tag, entitiesIdLine, newsUrlIdLine string
		err := rows.Scan(&e.Id, &e.Date, &e.Category, &tag, &e.Text, &entitiesIdLine, &newsUrlIdLine)
		if err != nil {
			break
		}
		entitiesId, err := util.SplitIntByTab(entitiesIdLine)
		if err != nil {
			return nil, err
		}
		newsSourceUrlId, err := util.SplitIntByTab(newsUrlIdLine)
		if err != nil {
			return nil, err
		}
		e.Tags = strings.Split(tag, "\t")
		e.EntitiesId = entitiesId
		e.NewsSourceUrlId = newsSourceUrlId
		events = append(events, e)
	}
	return events, nil
}

// SelectDateは探索済み日付の中に、与えられた日付が含まれるかどうかを確認する。
func SelectDate(db *sql.DB, date string) (bool, error) {
	stmt, err := db.Prepare("SELECT date FROM searched_date WHERE date = ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SelectDate()]", err)
		return false, errors.New(str)
	}
	defer stmt.Close()
	var str string
	err = stmt.QueryRow(date).Scan(&str)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// SelectNewsArtsEmptyDataは、diffbotからデータ取得ができなかったものを取り出す。
func SelectNewsArtsEmptyData(db *sql.DB) ([]NewsArt, error) {
	stmt, err := db.Prepare("SELECT news_art_id, news_source_url FROM news_diffbot WHERE timestamp = ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[SelectNewsArtsEmptyData()]", err)
		return nil, errors.New(str)
	}
	defer stmt.Close()
	rows, err := stmt.Query("2006-01-02")
	if err != nil {
		return nil, err
	}
	var news []NewsArt
	for rows.Next() {
		var path string
		var id int
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		news = append(news, NewsArt{Id: id, NewsSourceUrl: path})
	}
	return news, nil
}
