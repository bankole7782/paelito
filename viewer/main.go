package main

import (
	"github.com/webview/webview"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"html/template"
  "strings"
  "github.com/bankole7782/paelito/paelito_shared"
)


func main() {
	debug := false
	if os.Getenv("PANDOLEE_DEVELOPER") == "true" {
		debug = true
	}
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Paelito: A book reader.")
	w.SetSize(1200, 800, webview.HintNone)

	port := "45362"

  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

	go func() {

	  r := mux.NewRouter()

	  // r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(contentStatics))))

	  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      dirFIs, err := os.ReadDir(filepath.Join(rootPath, "out"))
      if err != nil {
        errorPage(w, errors.Wrap(err, "os error"))
        return
      }
      booksMap := make([]map[string]string, 0)
      for _, dirFI := range dirFIs {
        if strings.HasSuffix(dirFI.Name(), ".pae1") {
          bk := map[string]string {
            "filename" : dirFI.Name(),
            "title" : strings.ReplaceAll(dirFI.Name(), ".pae1", ""),
          }
          booksMap = append(booksMap, bk)
        }
      }
      type Context struct {
        Books []map[string]string
      }
      tmpl := template.Must(template.ParseFS(content, "templates/home.html"))
      tmpl.Execute(w, Context{booksMap})
	  })

    r.HandleFunc("/view_book/{filename}", viewBook)
    r.HandleFunc("/gba/{filename}/{filename2}", getBookAsset)
    r.HandleFunc("/view_book_chapter/{book_name}/{ch_filename}", viewBookChapter)

	  http.ListenAndServe(fmt.Sprintf(":%s", port), r)

	}()

	w.Navigate(fmt.Sprintf("http://127.0.0.1:%s", port))
	w.Run()

}
