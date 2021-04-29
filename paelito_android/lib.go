package paelito_android

import (
  "github.com/bankole7782/paelito/paelito_shared"
  "strings"
  "github.com/bankole7782/mof"
  "compress/gzip"
  "path/filepath"
  "os"
  "github.com/pkg/errors"
  "io"
  "encoding/json"
)


func UnpackBook(filename string) error {
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

  rawDetails, err := os.ReadFile(filepath.Join(obFolder, "out", "details.json"))
  if err != nil {
    return errors.Wrap(err, "os error")
  }
  detailsObj := make(map[string]string)
  err = json.Unmarshal(rawDetails, &detailsObj)
  if err != nil {
    return errors.Wrap(err, "json error")
  }

  os.MkdirAll(filepath.Join(rootPath, ".maps"), 0777)
  os.WriteFile(filepath.Join(rootPath, ".maps", detailsObj["BookId"]), []byte(bookName), 0777)
  return nil
}
