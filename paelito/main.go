package main

import (
  "github.com/bankole7782/paelito/paelito_internal"
  "fmt"
  "runtime"
  "os/exec"
  "os/signal"
  "os"

)


func main() {
  port := "45362"
  
  go paelito_internal.StartBackend()

  fmt.Printf("Running at http://127.0.0.1:%s\n", port)
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/C", "start", fmt.Sprintf("http://127.0.0.1:%s", port)).Output()
	} else if runtime.GOOS == "linux" {
		exec.Command("xdg-open", fmt.Sprintf("http://127.0.0.1:%s", port) ).Run()
	}
	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	}
	fmt.Println("Exiting")
}
