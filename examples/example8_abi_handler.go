package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler/abi"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

var valuesMap map[string]string

var noticeCodec *abihandler.Codec

func HandleAdvanceGet(metadata *rollups.Metadata, payloadSlice []interface{}) error {
  return HandleGet(payloadSlice)
}

func HandleGet(payloadSlice []interface{}) error {
  infolog.Println("Route: get, payload:",payloadSlice)
  key, ok := payloadSlice[0].(string)

  if !ok || key == "" {
    message := "HandleGet: Not enough parameters, you must provide " + key
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleGet: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  value := valuesMap[key]
  report := rollups.Report{rollups.Str2Hex(fmt.Sprint("Value of ",key," is ",value))}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleGet: error making http request: %s", err)
  }
  
  return nil
}

func HandleWrongWay(payloadHex string) error {
  message := "Unrecognized input, you should send a valid input"
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleWrongWay: error making http request: %s", err)
  }
  return fmt.Errorf(message)
}

func HandleSet(metadata *rollups.Metadata, payloadSlice []interface{}) error {
  infolog.Println("Route: set, payload:",payloadSlice)
  key, okKey := payloadSlice[0].(string)
  value, okVal := payloadSlice[1].(string)

  if !okKey || !okVal || key == "" || value == "" {
    message := "HandleSet: Not enough parameters, you must provide string 'key' and 'value'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleSet: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  valuesMap[key] = value

  report := rollups.Report{rollups.Str2Hex(fmt.Sprint("Value ",value," set for ",key))}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleSet: error making http request: %s", err)
  }

  noticePayload,err := noticeCodec.Encode([]interface{}{metadata.Timestamp,key,value})
  if err != nil {
    return fmt.Errorf("HandleSet: encoding notice: %s", err)
  }
  notice := rollups.Notice{noticePayload}
  _, err = rollups.SendNotice(&notice)
  if err != nil {
    return fmt.Errorf("HandleSet: error making http request: %s", err)
  }

  return nil
}

func main() {
  valuesMap = make(map[string]string)

  handler := abihandler.NewAbiHandler()
  handler.SetDebug()

  noticeCodec = abihandler.NewCodec([]string{"uint","string","string"})

  setCodec := abihandler.NewHeaderCodec("dapp","set",[]string{"string","string"})
  handler.HandleAdvanceRoute(setCodec, HandleSet)

  getCodec := abihandler.NewHeaderCodec("dapp","get",[]string{"string"})
  handler.HandleAdvanceRoute(getCodec, HandleAdvanceGet)
  handler.HandleInspectRoute(getCodec, HandleGet)

  handler.HandleDefault(HandleWrongWay)

  err := handler.Run()
  if err != nil {
    log.Panicln(err)
  }
}