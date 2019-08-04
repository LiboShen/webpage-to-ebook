package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	epub "github.com/bmaupin/go-epub"
	readability "github.com/go-shiori/go-readability"
)

func fetchArticle(url string) (readability.Article, error) {
	return readability.FromURL(url, 30*time.Second)

}

func createEpub(article readability.Article) ([]byte, error) {
	e := epub.NewEpub(article.Title)
	e.AddSection("<h1>"+article.Title+"</h1>"+article.Content, article.Title, "", "")
	tmpDir, err := ioutil.TempDir("", "webpage-to-epub")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	e.Write(path.Join(tmpDir, "out.epub"))
	result, err := ioutil.ReadFile(path.Join(tmpDir, "out.epub"))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getFileName(title string) string {
	lowerTitle := strings.ToLower(title)
	r, _ := regexp.Compile(`[<>:"/\|?*-.]`)
	safeTitle := r.ReplaceAllString(lowerTitle, "_")
	return strings.Trim(safeTitle, "_")
}

func newEpubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Invalid methed", 400)
		return
	}
	url, ok := r.URL.Query()["target_url"]
	if !ok || len(url[0]) < 1 {
		http.Error(w, "Url parameter 'target_url' is missing", 400)
		return
	}
	article, err := fetchArticle(url[0])
	if err != nil {
		http.Error(w, "Failed to process web page", 500)
		return
	}
	result, err := createEpub(article)
	if err != nil {
		http.Error(w, "Failed to generate epub", 500)
		return
	}
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s.epub\"", getFileName(article.Title)))
	w.Header().Set("Content-Type", "application/epub+zip")
	w.Write(result)
}

func main() {
	http.HandleFunc("/epub/new", newEpubHandler)

	fmt.Println("server is running")
	http.ListenAndServe(":8080", nil)
}
