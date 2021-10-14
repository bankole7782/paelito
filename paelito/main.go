package main

import (
  "os/exec"
  "runtime"
  "github.com/bankole7782/paelito/paelito_shared"
  "fmt"
)

func main() {
  if runtime.GOOS == "windows" {
    exec.Command("cmd", "/C", "start", fmt.Sprintf("http://127.0.0.1:%s", paelito_shared.Port)).Run()
  } else if runtime.GOOS == "linux" {
    exec.Command("xdg-open", fmt.Sprintf("http://127.0.0.1:%s", paelito_shared.Port)).Run()
  }
}
