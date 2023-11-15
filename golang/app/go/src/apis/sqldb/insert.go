package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"main/apis/util"
	"os"
)

// wiki_eventを登録する
func InsertWikiEvent(db *sql.DB, d Event) error {
	stmt, err := db.Prepare("INSERT INTO wiki_event(date, category, tags, text, entitie, news_source_url) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		d.Date,
		d.Category,
		util.JoinStringByTab(d.Tags),
		d.Text,
		util.JoinIntByTab(d.EntitiesId),
		util.JoinIntByTab(d.NewsSourceUrlId))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// wiki記事を登録する
func InsertWikiArticle(db *sql.DB, d WikiArt) error {
	stmt, err := db.Prepare("INSERT INTO wiki_article(wiki_source_url, text, wiki_category) VALUES(?,?,?)")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[insertWikiArticle()]", err)
		return errors.New(str)
	}
	defer stmt.Close()
	_, err = stmt.Exec(d.WikiSourceUrl, d.Text, d.WikiCategory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// news記事を登録する
func InsertNewsArticle(db *sql.DB, d NewsArt) error {
	stmt, err := db.Prepare("INSERT INTO news_diffbot(timestamp, site_name, publisher_region, category, title, text, human_language, news_source_url) VALUES(?,?,?,?,?,?,?,?)")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[insertNewsArticle()]", err)
		return errors.New(str)
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		d.Timestamp,
		d.SiteName,
		d.PublisherRegion,
		d.Category,
		d.Title,
		d.Text,
		d.HumanLanguage,
		d.NewsSourceUrl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

// スクレイピングを行った日付を登録する
func InsertDate(db *sql.DB, date string) error {
	stmt, err := db.Prepare("INSERT INTO searched_date(date) VALUES(?)")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[insertNewsArticle()]", err)
		return errors.New(str)
	}
	defer stmt.Close()
	_, err = stmt.Exec(date)
	if err != nil {
		return err
	}
	return nil
}
