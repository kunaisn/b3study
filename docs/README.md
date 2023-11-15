# 学部研究用プログラム_データ収集・分析

ここに記すのは、2023年11月14日時点説明である。

## 概要

このプログラムは、2023年10月から始めた研究に使用するために作成したプログラム群である。

テーマは、Wikipediaデータを用いた出来事木の構築である。


### ファイルごとの説明

* [golang](/golang)：Go言語のプログラム、基本的にはこの中のファイルを実行する
* [python3](/python3)：Pythonプログラム、FlaskでAPIのように機能し、Goプログラムからのリクエストに答える
* [db](/db)：データベース、MySQLを使用している
* [docs](/docs)：README.mdが置いてある
* [etc](/etc)：研究で用いた雑多な資料やスクリーンショットなどが格納される

## 環境（ソフトウェアとバージョン）

* OS：MacOS Sonoma 14.0　（Intel Mac）
* Go言語:1.19.2-alpine3.15
* MySQL：8.0.34
* python：3.12.0
* Dockerを使用

※arm64Macで実行する場合、docker-compose.ymlファイルのmysql設定で「platform: linux/amd64」を追加する

## Golang

機能は随時追加され、統合や廃止が頻繁に起こる。

### golangのファイル構成

ファイルは[golang/app/go/src](/golang/app/go/src)に存在している。

* main.go：[データ収集](#データ収集)プログラムが実行される（mainパッケージ）
* apis
  * sqldb（sqldbパッケージ）
    * db.go：DB関連を扱う
    * select.go：SELECT文を発行する
    * insert.go：INSERT文を発行する
    * update.go：UPDATE文を発行する
  * wiki（wikiパッケージ）
    * scraping.go：goqueryを用いてスクレイピングを行う
    * diffbot.go：diffbotを扱う　// 現在（2023年11月14日）API Keyが停止されている
  * util（utilパッケージ）
    * util.go：汎用関数を置いておく
    * sec.go：APIキーなどを置いておく（Gitで追跡されない）
* cmd（mainパッケージ）
  * rdb
    * main.go：DBのデータをもとに、Diffbot's APIを再度叩く
  * tagme
    * main.go：TagMe APIを叩き、文書から固有名詞を抽出する（1）
  * toPy
    * main.go：[Python3] APIを用いてTF-IDFを計算し、TagMeデータから情報エントロピーを計算してまとめる（2）
  * topics
    * main.go：TF-IDFを用いてコサイン類似度を計算し、情報エントロピーを考慮してトピックを分類する（3）
  * test
    * maing.go：分類したトピックを表示する（そのうち統合か廃止を行うため、testとしている）（4）

※(1)から(4)は順番に実行する必要がある。

Pythonに渡したいデータは[golang/app/go/data](/golang/app/go/data)に保存している。

同様に、Pythonから受け取りたいデータも同フォルダを用いて受け取る。


### データ収集

Wikipediaの[Current_events](https://en.wikipedia.org/wiki/Portal:Current_events/January_2023)からイベントを取得する。取得するデータは下記の通りである。

* Current_events
  * 日付
  * カテゴリ
  * タグ
  * 本文
  * 関連wikiページ群（同一IDが二つある場合がある）
  * 関連newsソース

上記に加えて、「**関連wikiページ群**」と「**関連newsソース**」のリンク先のデータも取得する。

#### 関連wikiページ群

取得したイベントには、多数のwiki内の記事がリンクされている。この記事の本文とカテゴリを取得する。リクエストは、Wikipediaサーバへの負荷をかけすぎないようにするため、2秒間隔で行う。この2秒という間隔は、取得に時間がかかる「関連newsソース」より長くならない程度に設定した。

* wiki_article
  * リンク（「...en.wikipedia.org」が省略されたURL）
  * 本文（HTML文全て）
  * カテゴリ

#### 関連newsソース

イベントには根拠となるニュース記事が存在するためこれも収集する。ニュース記事の構造はサイトによって大きく異なるため、Diffbot's APIを使用して構造化を行っている。場合によって、記事を正しく取得できないことがある。  
タイムスタンプが取得できなかった場合、**2007-01-02**として登録される。
また、エラーにより記事を取得できなかった場合、タイムスタンプは**2006-01-02**として、URLとID以外は空の値で登録される。

* news_diffbot
  * タイムスタンプ
  * 地域
  * カテゴリ
  * タイトル
  * 本文
  * 言語
  * 記事リンク（URL）

## Python3

イベントテキストをもとにgensimを使用してTF-IDFを訓練し、その結果をファイルに保存する。

依存関係は[requirements.txt](/python3/requirements.txt)を参照されたい。
