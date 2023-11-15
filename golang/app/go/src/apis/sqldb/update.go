package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
)

func UpdateDiffbotData(db *sql.DB, d NewsArt) error {
	stmt, err := db.Prepare("UPDATE news_diffbot SET timestamp = ?, site_name = ?, publisher_region = ?, category = ?, title = ?, text = ?, human_language = ? WHERE news_art_id = ?")
	if err != nil {
		str := fmt.Sprintf("%s: %v\n", "failed to generate statement[UpdateDiffbotData()]", err)
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
		d.Id)
	if err != nil {
		return err
	}
	return nil
}
