package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

func Handle(payload string) error {
  report := rollups.Report{payload}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("Handle: error making http request: %s", err)
  }
  return nil
}

func main() {
  rollups.HandleDefault(Handle)

  err := rollups.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}