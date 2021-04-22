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


func unpackBook(filename string) error {
  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    return err
  }

  bookName := strings.ReplaceAll(filename, ".pae1", "")
  obFolder := filepath.Join(rootPath, ".ob", bookName)
  os.MkdirAll(obFolder, 0777)

  inPath := filepath.Join(rootPath, "out", filename)
  inputFile, err := os.Open(inPath)
  if err != nil {
    return errors.Wrap(err, "os error")
  }
  defer inputFile.Close()

  zr, err := gzip.NewReader(inputFile)
  if err != nil {
    return errors.Wrap(err, "gzip error")
  }

  mofBytes, err := io.ReadAll(zr)
  if err != nil {
    return errors.Wrap(err, "io error")
  }

  err = os.WriteFile(filepath.Join(obFolder, "out.mof"), mofBytes, 0777)
  if err != nil {
    return errors.Wrap(err, "os error")
  }

  err = mof.UndoMOF(filepath.Join(obFolder, "out.mof"), obFolder)
  if err != nil {
    return errors.Wrap(err, "mof error")
  }


  return nil
}


type TableOfContent struct {
  Title string
  FileName string
  SubTOC []map[string]string
}


func viewBook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  filename := vars["filename"]

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  err = unpackBook(filename)
  if err != nil {
    errorPage(w, err)
    return
  }
  bookName := strings.ReplaceAll(filename, ".pae1", "")
  obFolder := filepath.Join(rootPath, ".ob", bookName)

  bookPath := filepath.Join(obFolder, "out")

  rawTOC, err := os.ReadFile(filepath.Join(bookPath, "rtoc.json"))
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

  tocs := make([]TableOfContent, 0)
  for _, rawTOCObj := range rawTOCObjs {
    subTOCPath := filepath.Join(bookPath, strings.ReplaceAll(rawTOCObj["html_filename"], ".html", "") + "_toc.json")
    subTOCRaw, err := os.ReadFile(subTOCPath)
    if err != nil {
      errorPage(w, errors.Wrap(err, "os error"))
      return
    }
    obj := make([]map[string]string, 0)
    err = json.Unmarshal(subTOCRaw, &obj)
    if err != nil {
      errorPage(w, errors.Wrap(err, "json error"))
      return
    }
    tocs = append(tocs, TableOfContent{rawTOCObj["name"], rawTOCObj["html_filename"], obj})
  }

  rawDetails, err := os.ReadFile(filepath.Join(bookPath, "details.json"))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  detailsObj := make(map[string]string)
  err = json.Unmarshal(rawDetails, &detailsObj)
  if err != nil {
    errorPage(w, errors.Wrap(err, ""))
  }
  authors := make([]string, 0)
  for k, v := range detailsObj {
    if strings.HasPrefix(k, "Author") {
      authors = append(authors, v)
    }
  }
  type Context struct {
    BookName string
    TOC []TableOfContent
    FirstFilename string
    Details map[string]string
    Authors []string
  }
  wv.SetTitle(bookName + " | Paelito: A book reader.")
  tmpl := template.Must(template.ParseFS(content, "templates/view_book.html"))
  tmpl.Execute(w, Context{bookName, tocs, rawTOCObjs[0]["html_filename"], detailsObj, authors})
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

  tocs := make([]TableOfContent, 0)
  for _, rawTOCObj := range rawTOCObjs {
    subTOCPath := filepath.Join(obFolder, strings.ReplaceAll(rawTOCObj["html_filename"], ".html", "") + "_toc.json")
    subTOCRaw, err := os.ReadFile(subTOCPath)
    if err != nil {
      errorPage(w, errors.Wrap(err, "os error"))
      return
    }
    obj := make([]map[string]string, 0)
    err = json.Unmarshal(subTOCRaw, &obj)
    if err != nil {
      errorPage(w, errors.Wrap(err, "json error"))
      return
    }
    tocs = append(tocs, TableOfContent{rawTOCObj["name"], rawTOCObj["html_filename"], obj})
  }


  var PreviousChapter string
  var NextChapter string
  for i, obj := range rawTOCObjs {
    if obj["html_filename"] == chapterFilename {
      if i != 0 {
        PreviousChapter = rawTOCObjs[i-1]["html_filename"]
      }
      if i + 1 != len(rawTOCObjs) {
        NextChapter = rawTOCObjs[i+1]["html_filename"]
      }
    }
  }
  rawChapterHTML, err := os.ReadFile(filepath.Join(obFolder, chapterFilename))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
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
    TOC []TableOfContent
    PageContents template.HTML
    PreviousChapter string
    NextChapter string
    HasBackground bool
    HasCSS bool
    CurrentChapter string
  }
  tmpl := template.Must(template.ParseFS(content, "templates/view_book_chapter.html"))
  tmpl.Execute(w, Context{bookName, tocs, template.HTML(string(rawChapterHTML)), PreviousChapter, NextChapter, hasBG,
    hasCSS, chapterFilename})
}
