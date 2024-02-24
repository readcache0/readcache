package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gorilla/mux"
)

func unEscape(htmlStr string) string {
	htmlStr = strings.ReplaceAll(htmlStr, "&lt;", "<")
	htmlStr = strings.ReplaceAll(htmlStr, "&gt;", ">")
	htmlStr = strings.ReplaceAll(htmlStr, "&quot;", "\"")
	htmlStr = strings.ReplaceAll(htmlStr, "&#39;", "'")
	htmlStr = strings.ReplaceAll(htmlStr, "&amp;", "&")
	return htmlStr
}

func main() {
	r := mux.NewRouter()

	tmpl, err := template.ParseFiles("./public/index.html")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	r.HandleFunc("/p", getPage).Methods("GET")

	port := 3000
	fmt.Printf("Server listening on port %d\n", port)
	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func getPage(w http.ResponseWriter, r *http.Request) {

	originalURL := r.URL.Query().Get("url")

	var url string

	index := strings.Index(originalURL, "?source")
	if index == -1 {
		url = originalURL
	} else {
		url = originalURL[:index]
	}

	c := colly.NewCollector()
	var Data string

	c.OnHTML("pre", func(e *colly.HTMLElement) {
		Data = e.Text
		Data = unEscape(Data)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(Data))
		if err == nil {
			doc.Find("script").Remove()
			Data, _ = doc.Html()
		}

		Data = strings.TrimSpace(Data)
	})

	err := c.Visit(fmt.Sprintf("http://webcache.googleusercontent.com/search?q=cache:%s&strip=0&vwsrc=1&&hl=en&lr=lang_en", url))
	if err != nil || Data == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "<h1>Error fetching data from Google Web Cache</h1>")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, Data)
}
