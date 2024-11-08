package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler/uri"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

var valuesMap map[string]string

func HandleAdvanceGet(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  return HandleGet(payloadMap)
}

func HandleGet(payloadMap map[string]interface{}) error {
  infolog.Println("Route: get, payload:",payloadMap)
  key, ok := payloadMap["key"].(string)

  if !ok || key == "" {
    message := "HandleGet: Not enough parameters, you must provide string 'key'"
    report := rollups.Report{Payload: rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleGet: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  value := valuesMap[key]
  report := rollups.Report{Payload: rollups.Str2Hex(fmt.Sprint("Value of ",key," is ",value))}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleGet: error making http request: %s", err)
  }
  
  return nil
}

func HandleSet(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  infolog.Println("Route: set, payload:",payloadMap)
  key, okKey := payloadMap["key"].(string)
  value, okVal := payloadMap["value"].(string)

  if !okKey || !okVal || key == "" || value == "" {
    message := "HandleSet: Not enough parameters, you must provide string 'key' and 'value'"
    report := rollups.Report{Payload: rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleSet: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  valuesMap[key] = value
  report := rollups.Report{Payload: rollups.Str2Hex(fmt.Sprint("Value ",value," set for ",key))}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleSet: error making http request: %s", err)
  }

  return nil
}


func HandleWrongWay(payloadHex string) error {
  message := "Unrecognized input, you should send a valid uri"
  report := rollups.Report{Payload: rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleWrongWay: error making http request: %s", err)
  }
  return fmt.Errorf(message)
}

func main() {
  valuesMap = make(map[string]string)

  uriHandler := urihandler.NewUriHandler()

  uriHandler.HandleAdvanceRoute("/set/:key/:value", HandleSet)
  uriHandler.HandleAdvanceRoute("/get/:key", HandleAdvanceGet)
  uriHandler.HandleInspectRoute("/get/:key", HandleGet)

  uriHandler.HandleDefault(HandleWrongWay)

  uriHandler.SetDebug()
  err := uriHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}