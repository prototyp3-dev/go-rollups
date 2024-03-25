package main

import (
  "encoding/json"
  "io"
  "strconv" 
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

func HandleAdvance(metadata *rollups.Metadata, payloadHex string) error {
  payload, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleAdvance: hex error decoding payload:", err)
  }
  infolog.Println("Advance request payload:", payload)

  notice := rollups.Notice{Payload: rollups.Str2Hex("Advanced " + payload)}
  res, err := rollups.SendNotice(&notice)
  if err != nil {
    return fmt.Errorf("HandleAdvance: error making http request: %s", err)
  }
 
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("HandleAdvance: could not read response body: %s", err)
  }
  
  var indexRes rollups.IndexResponse
  err = json.Unmarshal(body, &indexRes)
  if err != nil {
    return fmt.Errorf("HandleAdvance: Error unmarshaling body: %s", err)
  }
  infolog.Println("Received notice status", strconv.Itoa(res.StatusCode), "body", string(body), "index", strconv.FormatUint(indexRes.Index,10))

  return nil
}

func HandleInspect(payloadHex string) error {
  payload, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleInspect: hex error decoding payload: %s", err)
  }
  infolog.Println("Inspect request payload:", payload)

  report := rollups.Report{Payload: rollups.Str2Hex("Inspected " + payload)}
  res, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleInspect: error making http request: %s", err)
  }

  body, err := io.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("HandleInspect: could not read response body:", err)
  }
  infolog.Println("Received report status", strconv.Itoa(res.StatusCode), "body", string(body))
  
  return nil
}


func main() {
  rollupsHandler := handler.NewSimpleHandler()
  rollupsHandler.SetLogLevel(handler.Trace)
  
  rollupsHandler.HandleInspect(HandleInspect)
  rollupsHandler.HandleAdvance(HandleAdvance)

  err := rollupsHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}