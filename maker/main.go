package main

import (
  "os"
  // "io"
  "fmt"
  "path/filepath"
  "github.com/otiai10/copy"
  "encoding/json"
  "strings"
  "time"
  "github.com/bankole7782/paelito/paelito_shared"
  // "bytes"
  "github.com/PuerkitoBio/goquery"
  "strconv"
  "github.com/russross/blackfriday/v2"
  "github.com/bankole7782/zazabul"
)


func main() {
  if len(os.Args) != 2 {
    panic("The program expects one arguments: folder name")
  }

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

  inPath := filepath.Join(rootPath, "p", os.Args[1])
  if ! paelito_shared.DoesPathExists(inPath) {
    panic(fmt.Sprintf("The book dir '%s' is not in '%s'", os.Args[1], rootPath))
  }
  tmpFolderOut := filepath.Join(rootPath, ".mtmp-" + paelito_shared.UntestedRandomString(15))
  tmpFolder := filepath.Join(tmpFolderOut, os.Args[1])
  os.MkdirAll(tmpFolder, 0777)
  defer os.RemoveAll(tmpFolderOut)

  if ! paelito_shared.DoesPathExists(filepath.Join(inPath, "cover.png")) {
    panic("Your book must have a cover.png")
  }

  copy.Copy(filepath.Join(inPath, "cover.png"), filepath.Join(tmpFolder, "cover.png"))

  if paelito_shared.DoesPathExists(filepath.Join(inPath, "font.ttf")) {
    copy.Copy(filepath.Join(inPath, "font.ttf"), filepath.Join(tmpFolder, "font.ttf"))
  }

  // copy all the image files into the program.
  allDirFIS, _ := os.ReadDir(inPath)
  for _, dirFI := range allDirFIS {
    if strings.HasSuffix(dirFI.Name(), ".png") || strings.HasSuffix(dirFI.Name(), ".jpg") {
      copy.Copy(filepath.Join(inPath, dirFI.Name()), filepath.Join(tmpFolder, dirFI.Name()))
    }
  }

  // validate and update book details.zconf
  raw, err := os.ReadFile(filepath.Join(inPath, "details.zconf"))
  if err != nil {
    panic("Your book must have a details.zconf")
  }
  conf, err := zazabul.ParseConfig(string(raw))
  if err != nil {
    panic("Your conf is invalid.")
  }

  compulsoryKeys := []string{"title", "comment", "authors", "update_url", "source_url", "contact_email"}
  for _, key := range compulsoryKeys {
    if conf.Get(key) == "" {
      panic("Your details.zconf doesn't have the following field: " + key)
    }
  }

  conf.Update(map[string]string {
    "version": time.Now().Format(paelito_shared.VersionFormat),
    "date": time.Now().Format("2006-01-02"),
  })

  err = conf.Write(filepath.Join(tmpFolder, "details.zconf"))
  if err != nil {
    panic("Could not write details.zconf")
  }

  // load stop words
  stopWordsList := make([]string, 0)
  err = json.Unmarshal(englishStopwords, &stopWordsList)
  if err != nil {
    panic(err)
  }

  // convert markdowns to html files.
  rawTOC, err := os.ReadFile(filepath.Join(inPath, "toc.txt"))
  if err != nil {
    panic("You book must have toc.txt, this is the root table of content file.")
  }

  mapOfWordPositions := make(map[string][]paelito_shared.WordPosition)

  newTOCObjs := make([]map[string]string, 0)
  for _, part := range strings.Split(string(rawTOC), "\r\n\r\n") {
    parts := strings.Split(strings.TrimSpace(part), "\r\n")
    if len(parts) != 2 {
      continue
    }

    rawChapter, err := os.ReadFile(filepath.Join(inPath, parts[1]))
    if err != nil {
      panic(err)
    }

    html := string( blackfriday.Run(rawChapter) )

    // update the images in the document
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
      panic(err)
    }

    doc.Find("img").Each(func(i int, s *goquery.Selection) {
      src, _ := s.Attr("src")
      if ! strings.HasPrefix(src, "http") {
        s.SetAttr("src", filepath.Join("/gba/" + os.Args[1] + ".pae1/" + src))
      }
    })

    newHtml, _ := doc.Html()

    outFileName := strings.Replace(parts[1], ".md", ".html", 1)
    os.WriteFile(filepath.Join(tmpFolder, outFileName), []byte(newHtml), 0777)

    subTOC := make([]map[string]string, 0)
    doc.Find("h2").Each(func (i int, s *goquery.Selection) {
      aTOC := map[string]string {
        "index": strconv.Itoa(i + 1),
        "title": s.Text(),
      }
      subTOC = append(subTOC, aTOC)
    })
    subTOCJson, _ := json.Marshal(subTOC)
    subTOCFileName := strings.ReplaceAll(parts[1], ".md", "_toc.json")
    os.WriteFile(filepath.Join(tmpFolder, subTOCFileName), subTOCJson, 0777)

    // make an index for search sakes.
    doc.Find("p").Each(func (i int, s *goquery.Selection) {
      words := strings.Fields(s.Text())

      for _, word := range words {
        wordLower := strings.ToLower(word)

        if FindIn(stopWordsList, wordLower) != -1 {
          continue
        }

        wordPositions, ok := mapOfWordPositions[wordLower]
        if ! ok {
          mapOfWordPositions[wordLower] = []paelito_shared.WordPosition{
            {
              ParagraphIndex: i,
              HtmlFilename: outFileName,
              Word: wordLower,
            },
          }
        } else {
          wordPositions = append(wordPositions, paelito_shared.WordPosition{
            ParagraphIndex: i,
            HtmlFilename: outFileName,
            Word: wordLower,
          })
          mapOfWordPositions[wordLower] = wordPositions
        }
      }
    })

    nTOCObj := map[string]string {
      "name": parts[0],
      "html_filename": outFileName,
    }
    newTOCObjs = append(newTOCObjs, nTOCObj)
  }

  nTOCJson, err := json.Marshal(newTOCObjs)
  if err != nil {
    panic(err)
  }
  os.WriteFile(filepath.Join(tmpFolder, "rtoc.json"), nTOCJson, 0777)

  mapOfWordPositionsJson, err := json.Marshal(mapOfWordPositions)
  if err != nil {
    panic(err)
  }
  os.WriteFile(filepath.Join(tmpFolder, "index.json"), mapOfWordPositionsJson, 0777)

  outFilePath := filepath.Join(rootPath, "out", os.Args[1] + ".zip")
  os.MkdirAll(filepath.Join(rootPath, "out"), 0777)
  if paelito_shared.DoesPathExists(outFilePath) {
    os.Remove(outFilePath)
  }

  err = paelito_shared.ZipSource(tmpFolder, outFilePath)
  if err != nil {
    panic(err)
  }

  versionFilePath := filepath.Join(rootPath, "out", os.Args[1] + "_version.txt")
  os.WriteFile(versionFilePath, []byte(conf.Get("version")), 0777)
  fmt.Println("book path: ", outFilePath)
  fmt.Println("book version path: ", versionFilePath)

  fmt.Println("Upload the two generated files to your server.")
}


func FindIn(container []string, elem string) int {
	for i, o := range container {
		if o == elem {
			return i
		}
	}
	return -1
}
