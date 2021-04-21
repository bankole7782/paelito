package main

import (
  "fmt"
  "html/template"
  "strings"
  "net/http"
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
	tmpl := template.Must(template.ParseFS(content, "templates/base.html", "templates/error.html"))
	tmpl.Execute(w, Context{template.HTML(msg)})
}
