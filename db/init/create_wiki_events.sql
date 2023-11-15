
CREATE TABLE wiki_event (
	event_id INT AUTO_INCREMENT PRIMARY KEY,
	date DATE,
	category LONGTEXT,
	tags LONGTEXT,
	text LONGTEXT,
	entitie LONGTEXT,
	news_source_url LONGTEXT
);

CREATE TABLE wiki_article (
	wiki_art_id INT AUTO_INCREMENT PRIMARY KEY,
	wiki_source_url LONGTEXT,
	text LONGTEXT,
	wiki_category LONGTEXT
);

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

CREATE TABLE searched_date (
	date DATE PRIMARY KEY
);
