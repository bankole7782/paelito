package main

import (
  "fmt"
  "os"
  "github.com/jchv/go-webview2"
  "log"
)


func main() {
  port := "45362"
  debug := false
	if os.Getenv("SAENUMA_DEVELOPER") == "true" {
		debug = true
	}

  go StartBackend()

  w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     debug,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: "Paelito: A book maker and reader",
		},
	})
	if w == nil {
		log.Fatalln("Failed to load webview.")
	}
	defer w.Destroy()
	w.SetSize(1200, 700, webview2.HintNone)
	w.Navigate(fmt.Sprintf("http://127.0.0.1:%s", port))
	w.Run()
}
