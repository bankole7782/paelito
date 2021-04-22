package main

import (
  "fmt"
  "html/template"
  "strings"
  "net/http"
  "os"
  "path/filepath"
  "github.com/pkg/errors"
)


func errorPage(w http.ResponseWriter, err error) {
	type Context struct {
		Msg template.HTML
	}
	msg := fmt.Sprintf("%+v", err)
	fmt.Println(msg)
	msg = strings.ReplaceAll(msg, "\n", "<br>")
	msg = strings.ReplaceAll(msg, " ", "&nbsp;")
	msg = strings.ReplaceAll(msg, "\t", "&nbsp;&nbsp;")
	tmpl := template.Must(template.ParseFS(content, "templates/error.html"))
	tmpl.Execute(w, Context{template.HTML(msg)})
}


func emptyDir(path string) error {
	objFIs, err := os.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "ioutil error")
	}
	for _, objFI := range objFIs {
		os.RemoveAll(filepath.Join(path, objFI.Name()))
	}
	return nil
}
