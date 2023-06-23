package main

import (
  "log"
  "os"
  "fmt"

  "github.com/prototyp3-dev/go-rollups"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

type CustomHandler struct {
  rollups.Handler
  NAdvances uint32
  NInspects uint32
}

func (handler *CustomHandler) Advance(metadata *rollups.Metadata, payloadHex string) error {
  handler.NAdvances += 1
  
  message := fmt.Sprint("Number of advances: ",handler.NAdvances)
  infolog.Println(message)

  _, err := handler.SendNotice(&rollups.Notice{rollups.Str2Hex(message)})
  if err != nil {
    return err
  }

  return nil
}

func (handler *CustomHandler) Inspect(payloadHex string) error {
  handler.NInspects += 1

  message := fmt.Sprint("Number of inspects: ",handler.NInspects)
  infolog.Println(message)

  err := handler.SendReport(&rollups.Report{rollups.Str2Hex(message)})
  if err != nil {
    return err
  }

  return nil
}

func main() {
  handler := CustomHandler{}
  handler.SetLogLevel(rollups.Trace)

  handler.HandleInspect(handler.Inspect)
  handler.HandleAdvance(handler.Advance)

  err := handler.Run()
  if err != nil {
    log.Panicln(err)
  }
}
