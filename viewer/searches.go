package main

import (
  "net/http"
  "github.com/gorilla/mux"
  "github.com/bankole7782/paelito/paelito_shared"
  "os"
  "path/filepath"
  "github.com/pkg/errors"
  "encoding/json"
  "html/template"
  // "fmt"
  "strconv"
)


func searchBook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  bookName := vars["book_name"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  obFolder := filepath.Join(rootPath, ".ob", bookName, "out")

  rawDetails, err := os.ReadFile(filepath.Join(obFolder, "details.json"))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  detailsObj := make(map[string]string)
  err = json.Unmarshal(rawDetails, &detailsObj)
  if err != nil {
    errorPage(w, errors.Wrap(err, "json error"))
    return
  }

  if r.Method == http.MethodGet {
    type Context struct {
      BookName string
      BookTitle string
    }

    tmpl := template.Must(template.ParseFS(content, "templates/search_book.html"))
    tmpl.Execute(w, Context{bookName, detailsObj["FullTitle"]})

  } else {

    rawIndex, err := os.ReadFile(filepath.Join(obFolder, "index.json"))
    if err != nil {
      errorPage(w, errors.New("Please remake the book with the latest version of paelito. Index file missing"))
      return
    }

    mapOfWordPositions := make(map[string][]paelito_shared.WordPosition)
    err = json.Unmarshal(rawIndex, &mapOfWordPositions)
    if err != nil {
      errorPage(w, errors.Wrap(err, "json error"))
      return
    }

    hasBG := false
    if paelito_shared.DoesPathExists(filepath.Join(obFolder, "bg.png")) {
      hasBG = true
    }
    hasCSS := false
    if paelito_shared.DoesPathExists(filepath.Join(obFolder, "book.css")) {
      hasCSS = true
    }

    type Context struct {
      BookName string
      BookTitle string
      WordSearchedFor string
      WordPositions []paelito_shared.WordPosition
      WordPosition paelito_shared.WordPosition
      PageContents template.HTML
      BookId string
      HasBackground bool
      HasCSS bool
    }
    wordPositions, ok := mapOfWordPositions[r.FormValue("word_searched_for")]

    rawChapterHTML := make([]byte, 0)
    var wordPosition paelito_shared.WordPosition
    if ok {
      wordPosition = wordPositions[0]
      rawChapterHTML, err = os.ReadFile(filepath.Join(obFolder, wordPosition.HtmlFilename))
      if err != nil {
        errorPage(w, errors.Wrap(err, "os error"))
        return
      }

    }

    tmpl := template.Must(template.ParseFS(content, "templates/search_results.html"))
    tmpl.Execute(w, Context{bookName, detailsObj["FullTitle"], r.FormValue("word_searched_for"), wordPositions,
      wordPosition, template.HTML(string(rawChapterHTML)), detailsObj["BookId"], hasBG, hasCSS})

  }
}


func viewASearchResult(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  bookName := vars["book_name"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  obFolder := filepath.Join(rootPath, ".ob", bookName, "out")

  rawDetails, err := os.ReadFile(filepath.Join(obFolder, "details.json"))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  detailsObj := make(map[string]string)
  err = json.Unmarshal(rawDetails, &detailsObj)
  if err != nil {
    errorPage(w, errors.Wrap(err, "json error"))
    return
  }

  rawIndex, err := os.ReadFile(filepath.Join(obFolder, "index.json"))
  if err != nil {
    errorPage(w, errors.New("Please remake the book with the latest version of paelito. Index file missing"))
    return
  }

  mapOfWordPositions := make(map[string][]paelito_shared.WordPosition)
  err = json.Unmarshal(rawIndex, &mapOfWordPositions)
  if err != nil {
    errorPage(w, errors.Wrap(err, "json error"))
    return
  }

  hasBG := false
  if paelito_shared.DoesPathExists(filepath.Join(obFolder, "bg.png")) {
    hasBG = true
  }
  hasCSS := false
  if paelito_shared.DoesPathExists(filepath.Join(obFolder, "book.css")) {
    hasCSS = true
  }

  type Context struct {
    BookName string
    BookTitle string
    WordSearchedFor string
    WordPositions []paelito_shared.WordPosition
    WordPosition paelito_shared.WordPosition
    PageContents template.HTML
    BookId string
    HasBackground bool
    HasCSS bool
  }
  wordPositions, ok := mapOfWordPositions[vars["word"]]

  rawChapterHTML := make([]byte, 0)
  var wordPosition paelito_shared.WordPosition
  if ok {
    searchIndex, err := strconv.Atoi(vars["search_index"])
    if err != nil {
      errorPage(w, errors.Wrap(err, "strconv"))
      return
    }
    wordPosition = wordPositions[searchIndex]
    rawChapterHTML, err = os.ReadFile(filepath.Join(obFolder, wordPosition.HtmlFilename))
    if err != nil {
      errorPage(w, errors.Wrap(err, "os error"))
      return
    }

  }

  tmpl := template.Must(template.ParseFS(content, "templates/search_results.html"))
  tmpl.Execute(w, Context{bookName, detailsObj["FullTitle"], vars["word"], wordPositions,
    wordPosition, template.HTML(string(rawChapterHTML)), detailsObj["BookId"], hasBG, hasCSS})

}
