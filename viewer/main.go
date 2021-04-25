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
	"encoding/json"
	"os/exec"
)

var wv webview.WebView

func init() {
	rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

	os.MkdirAll(filepath.Join(rootPath, "lib"), 0777)
	os.MkdirAll(filepath.Join(rootPath, "p"), 0777)

	includedBooks := []map[string]string {
		{
			"book_url": "http://pandolee.com/static/the_botanum.pae1",
			"book_file_name": "the_botanum.pae1",
		},
		{
			"book_url": "http://pandolee.com/static/the_baileia.pae1",
			"book_file_name": "the_baileia.pae1",
		},
	}

	for _, m := range includedBooks {
		downloadFile(m["book_url"], filepath.Join(rootPath, "lib", m["book_file_name"]))
	}
}


func main() {
	debug := false
	if os.Getenv("PANDOLEE_DEVELOPER") == "true" {
		debug = true
	}
	w := webview.New(debug)
	wv = w
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
			wv.SetTitle("Paelito: A book reader.")

      dirFIs, err := os.ReadDir(filepath.Join(rootPath, "lib"))
      if err != nil {
        errorPage(w, errors.Wrap(err, "os error"))
        return
      }
      booksMap := make([]map[string]string, 0)
      for _, dirFI := range dirFIs {
        if strings.HasSuffix(dirFI.Name(), ".pae1") {
					err = unpackBook(dirFI.Name())
					if err != nil {
						errorPage(w, err)
						return
					}
					bookName := strings.ReplaceAll(dirFI.Name(), ".pae1", "")

					rawDetails, err := os.ReadFile(filepath.Join(rootPath, ".ob", bookName, "out", "details.json"))
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

          bk := map[string]string {
            "filename" : dirFI.Name(),
            "filename_no_ext": bookName,
						"title": detailsObj["FullTitle"],
						"comment": detailsObj["Comment"],
						"authors": strings.Join(authors, ", "),
						"date": detailsObj["Date"],
						"source_url": detailsObj["BookSourceURL"],
						"version": detailsObj["Version"],
						"bookid": detailsObj["BookId"],
          }
          booksMap = append(booksMap, bk)
        }
      }
      type Context struct {
        Books []map[string]string
				LibPath string
      }
      tmpl := template.Must(template.ParseFS(content, "templates/home.html"))
      tmpl.Execute(w, Context{booksMap, filepath.Join(rootPath, "lib")})
	  })

    r.HandleFunc("/view_book/{book_name}", viewBook)
    r.HandleFunc("/gba/{bookid}/{asset}", getBookAsset)
    r.HandleFunc("/view_book_chapter/{book_name}/{ch_filename}", viewBookChapter)
		r.HandleFunc("/gs/{obj}", func (w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			rawObj, err := contentStatics.ReadFile("statics/" + vars["obj"])
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Disposition", "attachment; filename=" + vars["obj"])
			contentType := http.DetectContentType(rawObj)
			w.Header().Set("Content-Type", contentType)
			w.Write(rawObj)
		})

		r.HandleFunc("/xdg/", func (w http.ResponseWriter, r *http.Request) {
			exec.Command("xdg-open", r.FormValue("p")).Run()
		})

		r.HandleFunc("/search_book/{book_name}", searchBook)
		r.HandleFunc("/view_a_search_result/{book_name}/{word}/{search_index}", viewASearchResult)

	  http.ListenAndServe(fmt.Sprintf(":%s", port), r)

	}()

	defer func() {
		emptyDir(filepath.Join(rootPath, ".ob"))
	}()

	defer func() {
		emptyDir(filepath.Join(rootPath, ".maps"))
	}()

	w.Navigate(fmt.Sprintf("http://127.0.0.1:%s", port))
	w.Run()

}
