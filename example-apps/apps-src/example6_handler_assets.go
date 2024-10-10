package main

import (
  "strconv"
  "fmt"
  "log"
  "os"
  "math/big"
  
  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

func HandleEther(metadata *rollups.Metadata, payloadHex string) error {
  deposit, err := rollups.DecodeEtherDeposit(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleEther: error decoding deposit: %s", err)
  }
  
  infolog.Println("Received",new(big.Float).Quo(new(big.Float).SetInt(deposit.Amount),big.NewFloat(1e18)),"native token deposit from",deposit.Depositor,"data:",string(deposit.Data))

  voucher := rollups.EtherWithdralVoucher(deposit.Depositor, deposit.Amount)
  infolog.Println("Sending voucher destination",voucher.Destination,"payload",voucher.Payload)
  res, err := rollups.SendVoucher(&voucher)
  if err != nil {
    return fmt.Errorf("HandleEther: error making http request: %s", err)
  }
  infolog.Println("Received voucher status", strconv.Itoa(res.StatusCode))
  return nil
}

func HandleErc20(metadata *rollups.Metadata, payloadHex string) error {
  deposit, err := rollups.DecodeErc20Deposit(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleErc20: error decoding deposit: %s", err)
  }
  infolog.Println("Received",deposit.Amount,"tokens",deposit.TokenAddress,"Erc20 deposit from",deposit.Depositor,"data:",string(deposit.Data))

  voucher := rollups.Erc20TransferVoucher(deposit.Depositor, deposit.TokenAddress, deposit.Amount)
  infolog.Println("Sending voucher destination",voucher.Destination,"payload",voucher.Payload)
  res, err := rollups.SendVoucher(&voucher)
  if err != nil {
    return fmt.Errorf("HandleErc20: error making http request: %s", err)
  }
  infolog.Println("Received voucher status", strconv.Itoa(res.StatusCode))
  return nil
}

func HandleErc721(metadata *rollups.Metadata, payloadHex string) error {
  deposit, err := rollups.DecodeErc721Deposit(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleErc721: error decoding deposit:", err)
  }

  infolog.Println("Received id",deposit.TokenId,deposit.TokenAddress,"Erc721 deposit from",deposit.Depositor,"data:",string(deposit.Data))

  voucher := rollups.Erc721SafeTransferVoucher(metadata.AppContract, deposit.Depositor, deposit.TokenAddress, deposit.TokenId)
  infolog.Println("Sending voucher destination",voucher.Destination,"payload",voucher.Payload)
  res, err := rollups.SendVoucher(&voucher)
  if err != nil {
    return fmt.Errorf("HandleErc721: error making http request: %s", err)
  }
  infolog.Println("Received voucher status", strconv.Itoa(res.StatusCode))
  return nil
}

func HandleErc1155Single(metadata *rollups.Metadata, payloadHex string) error {
  deposit, err := rollups.DecodeErc1155SingleDeposit(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleErc1155Single: error decoding deposit:", err)
  }

  infolog.Println("Received ",deposit.Amount,"tokens of id",deposit.TokenId,deposit.TokenAddress,"Erc1155 Single deposit from",deposit.Depositor,"base layer data:",string(deposit.BaseLayerData),"and exec layer data:",string(deposit.ExecLayerData))

  voucher := rollups.Erc1155SafeTransferFromVoucher(metadata.AppContract, deposit.Depositor, deposit.TokenAddress, deposit.TokenId, deposit.Amount, make([]byte,0))
  infolog.Println("Sending voucher destination",voucher.Destination,"payload",voucher.Payload)
  res, err := rollups.SendVoucher(&voucher)
  if err != nil {
    return fmt.Errorf("HandleErc721: error making http request: %s", err)
  }
  infolog.Println("Received voucher status", strconv.Itoa(res.StatusCode))
  return nil
}

func HandleErc1155Batch(metadata *rollups.Metadata, payloadHex string) error {
  deposit, err := rollups.DecodeErc1155BatchDeposit(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleErc1155Batch: error decoding deposit:", err)
  }

  infolog.Println("Received ",deposit.Amounts,"amounts of ids",deposit.TokenIds,deposit.TokenAddress,"Erc1155 Batch deposit from",deposit.Depositor,"base layer data:",string(deposit.BaseLayerData),"and exec layer data:",string(deposit.ExecLayerData))

  voucher := rollups.Erc1155SafeBatchTransferFromVoucher(metadata.AppContract, deposit.Depositor, deposit.TokenAddress, deposit.TokenIds, deposit.Amounts, make([]byte,0))
  infolog.Println("Sending voucher destination",voucher.Destination,"payload",voucher.Payload)
  res, err := rollups.SendVoucher(&voucher)
  if err != nil {
    return fmt.Errorf("HandleErc721: error making http request: %s", err)
  }
  infolog.Println("Received voucher status", strconv.Itoa(res.StatusCode))
  return nil
}

func main() {
  handler.InitializeRollupsAddresses("localhost")
  
  handler.HandleFixedAddress(handler.RollupsAddresses.EtherPortalAddress, HandleEther)
  handler.HandleFixedAddress(handler.RollupsAddresses.Erc20PortalAddress, HandleErc20)
  handler.HandleFixedAddress(handler.RollupsAddresses.Erc721PortalAddress, HandleErc721)
  handler.HandleFixedAddress(handler.RollupsAddresses.Erc1155SinglePortalAddress, HandleErc1155Single)
  handler.HandleFixedAddress(handler.RollupsAddresses.Erc1155BatchPortalAddress, HandleErc1155Batch)

  err := handler.RunDebug()
  if err != nil {
    log.Panicln(err)
  }
}