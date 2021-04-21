package main

import (
  "github.com/gomarkdown/markdown"
  // "github.com/gomarkdown/markdown/parser"
  "os"
  "github.com/gomarkdown/markdown/ast"
  "github.com/gomarkdown/markdown/html"
  "io"
  "fmt"
  "path/filepath"
  "github.com/otiai10/copy"
  "encoding/json"
  "strings"
  "github.com/bankole7782/mof"
  "compress/gzip"
  "time"
  "github.com/bankole7782/paelito/paelito_shared"
)


func DoesPathExists(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}


func main() {
  if len(os.Args) != 2 {
    panic("The program expects one arguments: folder name")
  }

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

  inPath := filepath.Join(rootPath, "p", os.Args[1])
  tmpFolder := filepath.Join(rootPath, ".mtmp-" + paelito_shared.UntestedRandomString(15))
  os.MkdirAll(tmpFolder, 0777)
  defer os.RemoveAll(tmpFolder)

  copy.Copy(filepath.Join(inPath, "cover.png"), filepath.Join(tmpFolder, "cover.png"))
  notNecessary := []string{"font1.ttf", "font2.ttf", "book.css"}
  for _, toCopy := range notNecessary {
    nnPath := filepath.Join(inPath, toCopy)
    if DoesPathExists(nnPath) {
      copy.Copy(nnPath, filepath.Join(tmpFolder, toCopy))
    }
  }

  rawTOC, err := os.ReadFile(filepath.Join(inPath, "rtoc.json"))
  if err != nil {
    panic(err)
  }
  rawTOCObjs := make([]map[string]string, 0)
  newTOCObjs := make([]map[string]string, 0)
  err = json.Unmarshal(rawTOC, &rawTOCObjs)
  if err != nil {
    panic(err)
  }
  opts := html.RendererOptions{
      Flags: html.CommonFlags,
      RenderNodeHook: renderHookDropCodeBlock,
  }
  renderer := html.NewRenderer(opts)

  for _, tocObj := range rawTOCObjs {
    rawChapter, err := os.ReadFile(filepath.Join(inPath, tocObj["filename"]))
    if err != nil {
      panic(err)
    }

    html := markdown.ToHTML([]byte(rawChapter), nil, renderer)
    outFileName := strings.Replace(tocObj["filename"], ".md", ".html", 1)
    os.WriteFile(filepath.Join(tmpFolder, outFileName), html, 0777)
    nTOCObj := map[string]string {
      "name": tocObj["name"],
      "html_filename": outFileName,
    }
    newTOCObjs = append(newTOCObjs, nTOCObj)
  }

  nTOCJson, err := json.Marshal(newTOCObjs)
  if err != nil {
    panic(err)
  }
  os.WriteFile(filepath.Join(tmpFolder, "rtoc.json"), nTOCJson, 0777)


  tmpFolder2 := filepath.Join(rootPath, ".mtmp-" + paelito_shared.UntestedRandomString(15))
  os.MkdirAll(tmpFolder2, 0777)
  defer os.RemoveAll(tmpFolder2)

  err = mof.MOF(tmpFolder, filepath.Join(tmpFolder2, "out.mof"))
  if err != nil {
    panic(err)
  }

  outFilePath := filepath.Join(rootPath, "out", os.Args[1] + ".pae1")
  os.MkdirAll(filepath.Join(rootPath, "out"), 0777)
  outFile, err := os.Create(outFilePath)
  if err != nil {
    panic(err)
  }
  defer outFile.Close()
  zw := gzip.NewWriter(outFile)
  zw.Name = os.Args[1] + ".pae1"
  zw.Comment = "A book"
  zw.ModTime = time.Now()

  mofBytes, err := os.ReadFile(filepath.Join(tmpFolder2, "out.mof"))
  if err != nil {
    panic(err)
  }
  _, err = zw.Write(mofBytes)
  if err != nil {
    panic(err)
  }

  if err := zw.Close(); err != nil {
    panic(err)
  }

  fmt.Println(outFilePath)
}


func renderHookDropCodeBlock(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
    // skip all nodes that are not CodeBlock nodes
	if _, ok := node.(*ast.CodeBlock); !ok {
		return ast.GoToNext, false
    }
    // custom rendering logic for ast.CodeBlock. By doing nothing it won't be
    // present in the output
	return ast.GoToNext, true
}
