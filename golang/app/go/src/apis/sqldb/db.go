package sqldb

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
・イベント管理テーブル（MySQL）
CREATE TABLE wiki_event (
	event_id INT AUTO_INCREMENT PRIMARY KEY,
	date DATE,
	category LONGTEXT,
	tags LONGTEXT,
	text LONGTEXT,
	entitie LONGTEXT,
	news_source_url LONGTEXT
);

・wiki内の記事を管理（MySQL）
CREATE TABLE wiki_article (
	wiki_art_id INT AUTO_INCREMENT PRIMARY KEY,
	wiki_source_url LONGTEXT,
	text LONGTEXT,
	wiki_category LONGTEXT
);

・wiki外の記事を管理（MySQL）
CREATE TABLE news_diffbot (
	news_art_id INT AUTO_INCREMENT PRIMARY KEY,
	timestamp DATE,
	site_name LONGTEXT,
	publisher_region LONGTEXT,
	category LONGTEXT,
	title LONGTEXT,
	text LONGTEXT,
	human_language LONGTEXT,
	news_source_url LONGTEXT
);
*/

type Event struct {
	Id              int
	Date            string
	Category        string
	Tags            []string
	Text            string
	Entities        []string
	EntitiesId      []int
	NewsSourceUrl   []string
	NewsSourceUrlId []int
}

type WikiArt struct {
	Id            int
	WikiSourceUrl string
	Text          string
	WikiCategory  string
}

type NewsArt struct {
	Id              int
	Timestamp       string
	SiteName        string
	PublisherRegion string
	Category        string
	Title           string
	HumanLanguage   string
	Text            string
	NewsSourceUrl   string
}

// 接続先DBの設定
const dsn = "docker:docker@tcp(db:3306)/data?charset=utf8mb4"

// DBに接続して応答を確認する (DBをクローズしないので、呼び出し元で「db.Close()」する)
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", "failed to open DB", err)
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 1200)
	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", "failed to connect DB", err)
		return nil, err
	}
	return db, nil
}

// wiki記事番号を調べて戻す。登録されていなかった場合は登録する。// 廃止(10月28日)
func GetWikiArtNumAndInsert(db *sql.DB, wikiAry []WikiArt) ([]int, error) {
	idAry := make([]int, 0)
	for _, wiki := range wikiAry {
		if len(wiki.WikiSourceUrl) == 0 {
			continue
		}
		get, err := SelectWikiArticle(db, wiki.WikiSourceUrl)
		if err != nil {
			return []int{}, err
		}
		if get.Id == -1 {
			err = InsertWikiArticle(db, wiki)
			if err != nil {
				return []int{}, err
			}
			get, err = SelectWikiArticle(db, wiki.WikiSourceUrl)
			if err != nil {
				return []int{}, err
			}
		}
		idAry = append(idAry, get.Id)
	}
	sort.Slice(idAry, func(i, j int) bool { return idAry[i] < idAry[j] })
	return idAry, nil
}

// news記事番号を調べて戻す。登録されていなかった場合は登録する。// 廃止(10月28日)
func GetNewsArtNumAndInsert(db *sql.DB, newsAry []NewsArt) ([]int, error) {
	idAry := make([]int, 0)
	for _, news := range newsAry {
		if len(news.NewsSourceUrl) == 0 {
			continue
		}
		get, err := SelectNewsArticle(db, news.NewsSourceUrl)
		if err != nil {
			return []int{}, err
		}
		if get.Id == -1 {
			err := InsertNewsArticle(db, news)
			if err != nil {
				return []int{}, err
			}
			get, err = SelectNewsArticle(db, news.NewsSourceUrl)
			if err != nil {
				return []int{}, err
			}
		}
		idAry = append(idAry, get.Id)
	}
	sort.Slice(idAry, func(i, j int) bool { return idAry[i] < idAry[j] })
	return idAry, nil
}

// GetEmptyDataOfNewsArtsは、diffbotからデータ取得ができなかったデータを戻す
// 「diffbotからデータ取得ができなかったデータ」は、タイムスタンプとURL以外が空の値で登録されている
func GetEmptyDataOfNewsArts() ([]NewsArt, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	news, err := SelectNewsArtsEmptyData(db)
	if err != nil {
		return nil, err
	}
	return news, nil
}
