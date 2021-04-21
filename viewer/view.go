package main

import (
  "net/http"
  "os"
  "github.com/gorilla/mux"
  "path/filepath"
  "github.com/bankole7782/mof"
  "github.com/bankole7782/paelito/paelito_shared"
  "compress/gzip"
  "github.com/pkg/errors"
  "io"
  // "fmt"
  "strings"
  "html/template"
  "encoding/json"

)

func viewBook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  filename := vars["filename"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  bookName := strings.ReplaceAll(filename, ".pae1", "")
  obFolder := filepath.Join(rootPath, ".ob", bookName)
  os.MkdirAll(obFolder, 0777)

  inPath := filepath.Join(rootPath, "out", filename)
  inputFile, err := os.Open(inPath)
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  defer inputFile.Close()

  zr, err := gzip.NewReader(inputFile)
  if err != nil {
    panic(err)
  }

  mofBytes, err := io.ReadAll(zr)
  if err != nil {
    panic(err)
  }

  err = os.WriteFile(filepath.Join(obFolder, "out.mof"), mofBytes, 0777)
  if err != nil {
    panic(err)
  }

  err = mof.UndoMOF(filepath.Join(obFolder, "out.mof"), obFolder)
  if err != nil {
    panic(err)
  }

  bookPath := filepath.Join(obFolder, "out")

  rawTOC, err := os.ReadFile(filepath.Join(bookPath, "rtoc.json"))
  if err != nil {
    panic(err)
  }
  rawTOCObjs := make([]map[string]string, 0)
  err = json.Unmarshal(rawTOC, &rawTOCObjs)
  if err != nil {
    panic(err)
  }
  type Context struct {
    BookName string
    TOC []map[string]string
  }
  tmpl := template.Must(template.ParseFS(content, "templates/view_book.html"))
  tmpl.Execute(w, Context{bookName, rawTOCObjs})

}


func getBookAsset(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  filename := vars["filename"]
  assetName := vars["filename2"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  obFolder := filepath.Join(rootPath, ".ob", strings.ReplaceAll(filename, ".pae1", ""), "out")

  http.ServeFile(w, r, filepath.Join(obFolder, assetName))
}


func viewBookChapter(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  bookName := vars["book_name"]
  chapterFilename := vars["ch_filename"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  obFolder := filepath.Join(rootPath, ".ob", bookName, "out")

  rawTOC, err := os.ReadFile(filepath.Join(obFolder, "rtoc.json"))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  rawTOCObjs := make([]map[string]string, 0)
  err = json.Unmarshal(rawTOC, &rawTOCObjs)
  if err != nil {
    errorPage(w, errors.Wrap(err, "json error"))
    return
  }

  rawChapterHTML, err := os.ReadFile(filepath.Join(obFolder, chapterFilename))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  type Context struct {
    BookName string
    TOC []map[string]string
    PageContents template.HTML
  }
  tmpl := template.Must(template.ParseFS(content, "templates/view_book_chapter.html"))
  tmpl.Execute(w, Context{bookName, rawTOCObjs, template.HTML(string(rawChapterHTML))})
}
