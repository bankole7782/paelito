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
  "fmt"
  "strings"
  "html/template"
  "encoding/json"
  "time"
  "strconv"
  "github.com/bankole7782/zazabul"
)


func unpackBook(filename string) error {
  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    return err
  }

  bookName := strings.ReplaceAll(filename, ".pae1", "")
  obFolder := filepath.Join(rootPath, ".ob", bookName)
  os.MkdirAll(obFolder, 0777)

  inPath := filepath.Join(rootPath, "lib", filename)
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
  filename := vars["book_name"]

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

  rawDetails, err := os.ReadFile(filepath.Join(bookPath, "details.zconf"))
  if err != nil {
    errorPage(w, errors.Wrap(err, "os error"))
    return
  }
  conf, err := zazabul.ParseConfig(string(rawDetails))
  if err != nil {
    errorPage(w, errors.Wrap(err, "zazbul error"))
    return
  }


  newVersionStr := ""
  resp, err := http.Get( conf.Get("update_url"))
  if err != nil {
    fmt.Println(err)
  }
  if err == nil {
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err == nil && resp.StatusCode == 200 {
      newVersionStr = string(body)
    }
  }

  hnv := false
  if newVersionStr != "" && newVersionStr != conf.Get("version") {
    time1, err1 := time.Parse(paelito_shared.VersionFormat, newVersionStr)
    time2, err2 := time.Parse(paelito_shared.VersionFormat, conf.Get("version"))

    if err1 == nil && err2 == nil && time2.Before(time1) {
      hnv = true
    }
  }

  authors := strings.Split(conf.Get("authors"), ",")

  type Context struct {
    BookName string
    TOC []TableOfContent
    FirstFilename string
    Authors []string
    HasNewVersion bool
    NewVersion string
    SourceURL string
    BookVersion string
    BookDate string
  }
  tmpl := template.Must(template.ParseFS(content, "templates/view_book.html"))
  tmpl.Execute(w, Context{bookName, tocs, rawTOCObjs[0]["html_filename"], authors,
    hnv, newVersionStr, conf.Get("source_url"), conf.Get("version"), conf.Get("date")})
}


func getBookAsset(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  bookName := vars["book_name"]
  assetName := vars["asset"]

  if strings.HasSuffix(bookName, ".pae1") {
    bookName = bookName[: len(bookName) - 5]
  }

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    errorPage(w, err)
    return
  }

  obFolder := filepath.Join(rootPath, ".ob", string(bookName), "out")

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

  var paraNum int
  var chapNum int
  if len(r.FormValue("chapter_num")) != 0 {
    chapterNum, err := strconv.Atoi(r.FormValue("chapter_num"))
    pNum, err2 := strconv.Atoi(r.FormValue("para_num"))
    if err == nil && err2 == nil && chapterNum <= len(rawTOCObjs){
      chapterFilename = rawTOCObjs[chapterNum - 1]["html_filename"]
      paraNum = pNum
      chapNum = chapterNum
    }
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
  hasFont := false
  if paelito_shared.DoesPathExists(filepath.Join(obFolder, "font.ttf")) {
    hasFont = true
  }

  type Context struct {
    BookName string
    TOC []TableOfContent
    PageContents template.HTML
    PreviousChapter string
    NextChapter string
    HasBackground bool
    HasFont bool
    CurrentChapter string
    ParaNum int
    IsAGotoPage bool
    ChapterNum int
  }

  tmpl := template.Must(template.ParseFS(content, "templates/view_book_chapter.html"))
  tmpl2 := template.Must(template.ParseFS(content, "templates/view_book_chapter_custom_font.html"))

  if hasFont {
    tmpl2.Execute(w, Context{bookName, tocs, template.HTML(string(rawChapterHTML)), PreviousChapter, NextChapter, hasBG,
      hasFont, chapterFilename, paraNum, paraNum > 0, chapNum})
  } else {
    tmpl.Execute(w, Context{bookName, tocs, template.HTML(string(rawChapterHTML)), PreviousChapter, NextChapter, hasBG,
      hasFont, chapterFilename, paraNum, paraNum > 0, chapNum})
  }

}
