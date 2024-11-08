package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

func Handle(payload string) error {
  report := rollups.Report{Payload: payload}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("Handle: error making http request: %s", err)
  }
  return nil
}

func main() {
  handler.HandleDefault(Handle)

  err := handler.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}