package main

import (
  "fmt"
  "log"
  "os"
  "math/big"
  "encoding/json"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler/abi"
  "github.com/prototyp3-dev/go-rollups/wallet"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

type MyApp struct {
  dappWallet *wallet.WalletApp
  feesMap map[abihandler.Address]*big.Int
  developerAddress abihandler.Address
  feeValue *big.Int
  feePayedCodec *abihandler.Codec
}

func (this *MyApp) GetPayedFee(addr abihandler.Address) *big.Int {
  if this.feesMap[addr] == nil {
    this.feesMap[addr] = new(big.Int)
  }
  return this.feesMap[addr]
}

func (this *MyApp) PayFee(metadata *rollups.Metadata, payloadSlice []interface{}) error {
  infolog.Println("Route: PayFee, payload:",payloadSlice)
  
  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("PayFee: error converting address: %s", err)
  }

  this.dappWallet.TransferEther(sender,this.developerAddress,this.feeValue)
  payedFee := this.GetPayedFee(sender)
  payedFee.Add(payedFee,this.feeValue)

  noticePayload,err := this.feePayedCodec.Encode([]interface{}{sender,this.GetPayedFee(sender)})
  if err != nil {
    return fmt.Errorf("PayFee: encoding notice: %s", err)
  }
  _, err = this.dappWallet.Handler().SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("PayFee: error making http request: %s", err)
  }

  return nil
}

func (this *MyApp) ChangeFee(metadata *rollups.Metadata, payloadSlice []interface{}) error {
  infolog.Println("Route: ChangeFee, payload:",payloadSlice)
  newFee, ok1 := payloadSlice[0].(*big.Int)
  
  if !ok1 {
    message := "ChangeFee: parameters error"
    err := this.dappWallet.Handler().SendReport(rollups.Str2Hex(message))
    if err != nil {
      return fmt.Errorf("ChangeFee: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  this.feeValue.Set(newFee)
  
  return nil
}

func (this *MyApp) GetFeeUri(payloadMap map[string]interface{}) error {
  addrStr, ok1 := payloadMap["address"].(string)
  if !ok1 {
    return fmt.Errorf("GetFee: parameters error")
  }

  addr,err := abihandler.Hex2Address(addrStr)
  if err != nil {
    return fmt.Errorf("GetFee: parameters error: %s", err)
  }

  return this.GetFee([]interface{}{addr})
}

func (this *MyApp) GetFee(payloadSlice []interface{}) error {
  infolog.Println("Route: GetFee, payload:",payloadSlice)
  addr, ok1 := payloadSlice[0].(abihandler.Address)
  
  if !ok1 {
    message := "GetFee: parameters error"
    err := this.dappWallet.Handler().SendReport(rollups.Str2Hex(message))
    if err != nil {
      return fmt.Errorf("GetFee: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  
  payedFeeJson, err := json.Marshal(struct{
    Address string  `json:"address"`
    PayedFee uint64 `json:"payedFee"`
  }{Address: addr.String(), PayedFee: this.GetPayedFee(addr).Uint64()})
  if err != nil {
    return fmt.Errorf("GetFee: error converting wallet to json: %s", err)
  }

  err = this.dappWallet.Handler().SendReport(rollups.Str2Hex(string(payedFeeJson)))
  if err != nil {
    return fmt.Errorf("GetFee: error making http request: %s", err)
  }

  return nil
}

func (this *MyApp) HandleWrongWay(payloadHex string) error {
  message := "Unrecognized input, you should send a valid input"
  err := this.dappWallet.Handler().SendReport(rollups.Str2Hex(message))
  if err != nil {
    return fmt.Errorf("HandleWrongWay: error making http request: %s", err)
  }
  return fmt.Errorf(message)
}

func main() {
  myApp := MyApp{}
  myApp.feesMap = make(map[abihandler.Address]*big.Int)
  developerAddress, err := abihandler.Hex2Address("0x70997970C51812dc3A010C7d01b50e0d17dc79C8") // test wallet #1
  if err != nil {
    panic(err)
  }
  myApp.developerAddress = developerAddress
  myApp.feeValue = new(big.Int)
  f,_ := big.NewFloat(1e16).Uint64()
  myApp.feeValue.SetUint64(f)

  myApp.feePayedCodec = abihandler.NewCodec([]string{"address","uint256"})

  appHandler := abihandler.NewAbiHandler()
  appHandler.SetDebug()
  appHandler.InitializeRollupsAddresses("localhost")

  // creates a new wallet
  myApp.dappWallet = wallet.NewWalletApp();
  myApp.dappWallet.SetAbiHandler(appHandler)

  // setups the dapp relay and fixed portal deposit routes
  //   overrides any fixed address handler
  //   and extra routes to control assets
  myApp.dappWallet.SetupRoutes([]wallet.WalletRoute{
    wallet.DepositEtherAdvanceRoute,
    wallet.WithdrawEtherAdvanceRoute,
    wallet.TransferEtherAdvanceRoute,
    wallet.BalanceInspectRoute,wallet.BalanceUriInspectRoute})

  appHandler.HandleAdvanceRoute(abihandler.NewHeaderCodec("dapp","fee",[]string{}), myApp.PayFee)
  appHandler.HandleFixedAddressAdvance(abihandler.Address2Hex(developerAddress),abihandler.NewHeaderCodec("dapp","changeFee",[]string{"uint256"}), myApp.ChangeFee)
  appHandler.HandleInspectRoute(abihandler.NewHeaderCodec("dapp","fee",[]string{"address"}), myApp.GetFee)
  myApp.dappWallet.UriHandler().HandleInspectRoute("/fee/:address", myApp.GetFeeUri)
  
  appHandler.HandleDefault(myApp.HandleWrongWay)

  err = appHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}