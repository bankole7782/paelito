package main

import (
  "fmt"
  "html/template"
  "strings"
  "net/http"
  "os"
  "path/filepath"
  "github.com/pkg/errors"
  "github.com/bankole7782/paelito/paelito_shared"
  "io"
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


func downloadFile(url, outPath string) error {
	if paelito_shared.DoesPathExists(outPath) {
		return nil
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "http error")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
  if err != nil {
    return errors.Wrap(err, "io error")
  }

	if resp.StatusCode != 200 {
		return errors.New(string(body))
	}
	// // Create the file
	// parts := strings.Split(outPath, "/")
	// pathDir := strings.Join(parts[: len(parts) - 1], "/")
	// fmt.Println(pathDir)
	// err = os.MkdirAll(pathDir, 0777)
	// if err != nil {
	// 	return errors.Wrap(err, "os error")
	// }

	out, err := os.Create(outPath)
	if err != nil {
		return errors.Wrap(err, "os error")
	}
	defer out.Close()

	// Write the body to file
	_, err = out.Write(body)
	if err != nil {
		return errors.Wrap(err, "io error")
	}

	return nil
}
