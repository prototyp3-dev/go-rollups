package main

import (
  "log"
  "os"
  "fmt"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

type CustomHandler struct {
  handler.Handler
  NAdvances uint32
  NInspects uint32
}

func (ch *CustomHandler) Advance(metadata *rollups.Metadata, payloadHex string) error {
  ch.NAdvances += 1
  
  message := fmt.Sprint("Number of advances: ",ch.NAdvances)
  infolog.Println(message)

  _, err := ch.SendNotice(rollups.Str2Hex(message))
  if err != nil {
    return err
  }

  return nil
}

func (ch *CustomHandler) Inspect(payloadHex string) error {
  ch.NInspects += 1

  message := fmt.Sprint("Number of inspects: ",ch.NInspects, "(shouldn't change)")
  infolog.Println(message)

  err := ch.SendReport(rollups.Str2Hex(message))
  if err != nil {
    return err
  }

  return nil
}

func main() {
  cutomHandler := CustomHandler{}
  cutomHandler.SetLogLevel(handler.Trace)

  cutomHandler.HandleInspect(cutomHandler.Inspect)
  cutomHandler.HandleAdvance(cutomHandler.Advance)

  err := cutomHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}
