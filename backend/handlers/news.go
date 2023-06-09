package handlers

import (
	"fmt"
	"net/http"
	"newsfeed/config"
	"newsfeed/models"
	"newsfeed/parsers"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type NewsHandlers struct {
	app  *config.Application
	data *models.Models
}

func NewNewsHandlers(app *config.Application, data *models.Models) *NewsHandlers {
	return &NewsHandlers{app, data}
}

func (n *NewsHandlers) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var news models.News
	err := readJson(w, r, &news)
	if err != nil {
		return
	}
	if news.FeedID == 0 {
		http.Error(w, "feed id not provided", http.StatusInternalServerError)
		return
	}
	dbNews, err := n.data.News.Create(&news)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, dbNews)
}

func (n *NewsHandlers) CreateByFeedUrl(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	url := getQueryParamString("url", r)
	if url == "" {
		http.Error(w, "feed url not provided", http.StatusInternalServerError)
		return
	}
	feed := n.data.Feeds.ByUrl(url)
	if feed == nil {
		http.Error(w, "feed needs to be created first", http.StatusInternalServerError)
		return
	}
	content, err := fetchFeed(url, w)
	if err != nil {
		return
	}
	_, news := parsers.ParseFeed(content)
	dbFeed := n.data.News.CreateBulk(news, feed.ID)
	writeJson(w, dbFeed)
}

func (n *NewsHandlers) Read(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	limit := getQueryParamInt("limit", 100, r)
	offset := getQueryParamInt("offset", 0, r)
	dbNews := n.data.News.Read(limit, offset)
	writeJson(w, dbNews)
}

func (n *NewsHandlers) ReadByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, "news id not provided", http.StatusInternalServerError)
		return
	}
	news := n.data.News.ByID(id)
	writeJson(w, news)
}

func (n *NewsHandlers) ReadByFeedID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, "news id not provided", http.StatusInternalServerError)
		return
	}
	news := n.data.News.ByFeedID(id)
	writeJson(w, news)
}

func (n *NewsHandlers) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var news models.News
	err := readJson(w, r, &news)
	if err != nil {
		return
	}
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, "news id not provided", http.StatusInternalServerError)
		return
	}
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		http.Error(w, "news id not a number", http.StatusInternalServerError)
		return
	}
	news.ID = uint(idUint)
	dbNews := n.data.News.Update(&news)
	writeJson(w, dbNews)
}

func (n *NewsHandlers) Remove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		http.Error(w, "news id not provided", http.StatusInternalServerError)
		return
	}
	n.data.News.Remove(id)
	w.Write([]byte(fmt.Sprintf("news %s deleted", id)))
}

func (n *NewsHandlers) Drop(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	n.data.News.Drop()
	w.Write([]byte("dropped news table"))
}
