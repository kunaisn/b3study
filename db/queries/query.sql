-- name: InsertWikiEvent :exec
INSERT INTO wiki_event(
    date, category, tags, text, entitie, news_source_url
) VALUES (?,?,?,?,?,?);

-- name: InsertWikiArticle :exec
INSERT INTO wiki_article(
    wiki_source_url, text, wiki_category
) VALUES (?,?,?);

-- name: InsertNewsArticle :exec
INSERT INTO news_diffbot(
    timestamp, site_name, publisher_region, category, title, text, human_language, news_source_url
) VALUES (?,?,?,?,?,?,?,?);

-- name: InsertDate :exec
INSERT INTO searched_date(
    date
) VALUES (?);

-- name: SelectWikiArticle :one
SELECT wiki_art_id, wiki_source_url, text, wiki_category
FROM wiki_article
WHERE wiki_source_url = ?;

-- name: SelectNewsArticle :one
SELECT news_art_id, timestamp, site_name, publisher_region, category, title, text, human_language, news_source_url
FROM news_diffbot
WHERE news_source_url = ?;

-- name: SelectEvents :many
SELECT category, tags, text, entitie, news_source_url 
FROM wiki_event
WHERE date = ?;

-- name: SelectDate :one
SELECT date
FROM searched_date
WHERE date = ?
LIMIT 1;

-- name: SelectNewsArtsEmptyData :many
SELECT news_art_id, news_source_url
FROM news_diffbot
WHERE timestamp = ?
LIMIT 1;

-- name: UpdateDiffbotData :exec
UPDATE news_diffbot 
SET timestamp = ?, site_name = ?, publisher_region = ?, category = ?, title = ?, text = ?, human_language = ? 
WHERE news_art_id = ?;