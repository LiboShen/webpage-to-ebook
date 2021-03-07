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
	"encoding/base64"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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
}

func handler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if r.HTTPMethod != "GET" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid methed",
		}, nil
	}
	url := r.QueryStringParameters["target_url"]
	if url == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Url parameter 'target_url' is missing",
		}, nil
	}
	article, err := fetchArticle(url)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to process web page",
		}, nil
	}
	result, err := createEpub(article)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to generate epub",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Disposition": fmt.Sprintf("attachment; filename=\"%s.epub\"", getFileName(article.Title)),
			"Content-Type":        "application/epub+zip",
		},
		Body:            base64.StdEncoding.EncodeToString(result),
		IsBase64Encoded: true,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}
