# webpage-to-ebook

A simple API server extracts web articles and turn them into ebook format (only `epub` currently).


## Install & Run

git clone.

Then

```
glide install
go run server.go
```

## Internal

It uses [go-readability](https://github.com/go-shiori/go-readability_) for atricle extracting and [epub-gen](https://github.com/bmaupin/go-epub) for epub generating.