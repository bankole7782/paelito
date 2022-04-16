package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"html/template"
  "strings"
  "github.com/bankole7782/paelito/paelito_shared"
	"os/exec"
	"runtime"
  "github.com/bankole7782/zazabul"
	"os/signal"
)


func init() {
	rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

	os.MkdirAll(filepath.Join(rootPath, "lib"), 0777)
	os.MkdirAll(filepath.Join(rootPath, "p"), 0777)

	includedBooks := []map[string]string {
		{
			"book_url": "http://sae.ng/static/the_botanum.pae1",
			"book_file_name": "the_botanum.pae1",
		},
		{
			"book_url": "http://sae.ng/static/the_baileia.pae1",
			"book_file_name": "the_baileia.pae1",
		},
	}

	for _, m := range includedBooks {
		err := downloadFile(m["book_url"], filepath.Join(rootPath, "lib", m["book_file_name"]))
		if err != nil {
			fmt.Printf("%+v\n", err)
			panic(err)
		}
	}


}


func main() {
	port := "45362"
  rootPath, err := paelito_shared.GetRootPath()
  if err != nil {
    panic(err)
  }

  defer func() {
		emptyDir(filepath.Join(rootPath, ".ob"))
	}()

	go func() {

	  r := mux.NewRouter()

	  // r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(contentStatics))))

	  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

					rawDetails, err := os.ReadFile(filepath.Join(rootPath, ".ob", bookName, "out", "details.zconf"))
					if err != nil {
						errorPage(w, errors.Wrap(err, "Invalid format. Please redownload an updated version."))
						return
					}
					conf, err := zazabul.ParseConfig(string(rawDetails))
					if err != nil {
						errorPage(w, errors.Wrap(err, "zazbul error"))
						return
					}

          bk := map[string]string {
            "filename" : dirFI.Name(),
            "filename_no_ext": bookName,
						"title": conf.Get("title"),
						"comment": conf.Get("comment"),
						"authors": conf.Get("authors"),
						"date": conf.Get("date"),
						"source_url": conf.Get("source_url"),
						"version": conf.Get("version"),
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
    r.HandleFunc("/gba/{book_name}/{asset}", getBookAsset)
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
			if runtime.GOOS == "windows" {
				exec.Command("start", r.FormValue("p")).Run()
			} else if runtime.GOOS == "linux" {
				exec.Command("xdg-open", r.FormValue("p")).Run()
			}
		})

		r.HandleFunc("/favicon.ico", func (w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/gs/paelito.ico", 301)
		})

		r.HandleFunc("/search_book/{book_name}", searchBook)
		r.HandleFunc("/view_a_search_result/{book_name}/{word}/{search_index}", viewASearchResult)

	  err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

	}()

	fmt.Printf("Running at http://127.0.0.1:%s\n", port)
	exec.Command("xdg-open", fmt.Sprintf("http://127.0.0.1:%s", port) ).Run()
	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	}
	fmt.Println("Exiting")

}
