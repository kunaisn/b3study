import os
import gensim
from gensim.models import LsiModel
from gensim import corpora
from gensim import similarities
from collections import defaultdict
from nltk.tokenize import RegexpTokenizer
from nltk.stem.porter import PorterStemmer
import pickle
import json
from flask import Flask, request, jsonify


en_stop_list = [
    "i", "me", "my", "myself", "we", "our", "ours", "ourselves", "you", "your", "yours",
    "yourself", "yourselves", "he", "him", "his", "himself", "she", "her", "hers", "herself",
    "it", "its", "itself", "they", "them", "their", "theirs", "themselves", "what", "which",
    "who", "whom", "this", "that", "these", "those", "am", "is", "are", "was", "were", "be",
    "been", "being", "have", "has", "had", "having", "do", "does", "did", "doing", "a", "an",
    "the", "and", "but", "if", "or", "because", "as", "until", "while", "of", "at", "by", "for",
    "with", "about", "against", "between", "into", "through", "during", "before", "after",
    "above", "below", "to", "from", "up", "down", "in", "out", "on", "off", "over", "under",
    "again", "further", "then", "once", "here", "there", "when", "where", "why", "how", "all",
    "any", "both", "each", "few", "more", "most", "other", "some", "such", "no", "nor", "not",
    "only", "own", "same", "so", "than", "too", "very", "s", "t", "can", "will", "just", "don",
    "should", "now"
]


class EventData:
    id = -1
    text = ""
    date = ""
    entities = []
    tf_idf = {}

    def __init__(self, event):
        self.id = event["id"]
        self.date = event["date"]
        self.text = event["text"]
        self.entities = event["entities"]

    def set_tf_idf(self, tf_idf):
        self.tf_idf = tf_idf


def text_cleaning(events):
    # 外部ライブラリの初期化
    ps = PorterStemmer()
    all_tokens = []
    for event in events:
        # 余分な空白文字を削除、小文字化
        line = event.text.strip().lower()
        tokenizer = RegexpTokenizer(r'\w+')
        tokens = tokenizer.tokenize(line)
        stopped_tokens = [i for i in tokens if not i in en_stop_list]
        stemmed_tokens = [ps.stem(i) for i in stopped_tokens]
        all_tokens.append(stemmed_tokens)
    return all_tokens


def open_event_data_json(path):
    # JSONファイルを変換
    with open(path, "r") as f:
        d = json.load(f)
    events = []
    for event in d["events"]:
        events.append(EventData(event))
    return events


def write_event_data_json(d, path):
    with open(path, 'wt') as f:
        json.dump(d, f)


def create_corpus(events):
    # 全てのトークンを格納
    all_tokens = text_cleaning(events)
    # tokens内の単語それぞれの出現頻度を調べる
    frequency = defaultdict(int)
    for tokens in all_tokens:
        for t in tokens:
            frequency[t] += 1
    # borderの数以下の出現頻度の言葉を削除する
    border = 5
    all_tokens = [[t for t in tokens if frequency[t] > border] for tokens in all_tokens]
    # トークンを保存
    with open("texts.pkl", 'wb') as f:
        pickle.dump(all_tokens, f)
    # 辞書を作成
    dictionary = corpora.Dictionary(all_tokens)
    dictionary.filter_extremes(no_below=2, no_above=0.8)
    # 辞書を保存
    dictionary.save_as_text("dictionary.txt")
    # BoWのコーパスを作成
    corpus = [dictionary.doc2bow(t) for t in all_tokens]
    with open("corpus.pkl", 'wb') as f:
        pickle.dump(corpus, f)
    # BoWをtf-idfに変換
    tfidf = gensim.models.TfidfModel(corpus)
    tfidf.save('model.tfidf')
    corpus_tfidf = tfidf[corpus]
    idx = 0
    for tfidf in corpus_tfidf:
        tfidfdict = {}
        for t in tfidf:
            tfidfdict[str(t[0])] = t[1]
        events[idx].set_tf_idf(tfidfdict)
        idx += 1
    # 保存
    with open("corpus_tfidf.pkl", 'wb') as f:
        pickle.dump(corpus_tfidf, f)


def culc_cosine_sim(events):
    with open("corpus_tfidf.pkl", 'rb') as f:
        corpus = pickle.load(f)
    # 類似度行列を作成
    matrix = similarities.MatrixSimilarity(corpus)
    sims = matrix[corpus]
    for i in range(len(events)):
        events[i].set_con_sim(sims[i].tolist())
    return


def get_dict(events):
    data = {"events": []}
    for e in events:
        data["events"].append({
            "date": e.date,
            "id": e.id,
            "text": e.text,
            "entities": e.entities,
            "tf_idf": e.tf_idf})
    return data


app = Flask(__name__)


@app.route('/', methods=['GET'])
def culc_cos_sim_from_events():
    app.logger.info("start")
    try:
        all_events = open_event_data_json("/go/data/go/data/EventsDataJSON.json")
        create_corpus(all_events)
        app.logger.info("fin create_corpus")
        # culc_cosine_sim(all_events)
        # app.logger.info("fin culc_cosine_sim")
        write_event_data_json(get_dict(all_events), "/go/data/go/data/CosSim.json")
        app.logger.info("fin get_dict")
    except:
        app.logger.info("error")
        return "Bad Request", 400
    return "OK", 200


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8050)))
