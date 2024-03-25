package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

var relayMessage string

func GenericHandler(payloadHex string) error {
  payload, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("GenericHandler: hex error decoding payload: %s", err)
  }
  infolog.Println("Generic request payload:", payload)

  report := rollups.Report{Payload: rollups.Str2Hex("Generic " + payload + relayMessage)}
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


func HandleRelay(metadata *rollups.Metadata, payloadHex string) error {
  infolog.Println("Hey, I know this address, sender is",metadata.MsgSender,"and the my address is", payloadHex)
  relayMessage = fmt.Sprint(" and the dapp address is ",payloadHex)
  report := rollups.Report{Payload: rollups.Str2Hex("Set address relay")}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleFixed: error making http request: %s", err)
  }

  return nil
}


func main() {
  handler.InitializeRollupsAddresses("localhost")

  handler.HandleDefault(GenericHandler)
  handler.HandleRollupsFixedAddresses(HandleFixed)
  handler.HandleFixedAddress(handler.RollupsAddresses.DappAddressRelay, HandleRelay)

  err := handler.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}