package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

var operatorMessage string

func GenericHandler(payloadHex string) error {
  payload, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("GenericHandler: hex error decoding payload: %s", err)
  }
  infolog.Println("Generic request payload:", payload)

  report := rollups.Report{Payload: rollups.Str2Hex("Generic " + payload + operatorMessage)}
  _, err = rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("GenericHandler: error making http request: %s", err)
  }

  return nil
}

func HandleFixed(metadata *rollups.Metadata, payloadHex string) error {
  infolog.Println("Received deposit")
  report := rollups.Report{Payload: rollups.Str2Hex("Warning: Ignored any desposits")}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleFixed: error making http request: %s", err)
  }

  return nil
}


func HandleOperator(metadata *rollups.Metadata, payloadHex string) error {
  infolog.Println("Hey, I know this address, sender is",metadata.MsgSender,"and the message is", payloadHex)
  payloadStr, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleOperator: hex error decoding payload:", err)
  }
  operatorMessage = fmt.Sprint(" and the message is ",payloadStr)
  report := rollups.Report{Payload: rollups.Str2Hex("Set operator message")}
  _, err = rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleFixed: error making http request: %s", err)
  }

  return nil
}


func main() {
  handler.InitializeRollupsAddresses("localhost")

  operator := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

  handler.HandleDefault(GenericHandler)
  handler.HandleRollupsFixedAddresses(HandleFixed)
  handler.HandleFixedAddress(operator, HandleOperator)

  err := handler.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}