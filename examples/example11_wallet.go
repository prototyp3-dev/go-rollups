package main

import (
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/wallet"
)

var infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)

var dappWallet *wallet.wallet

func HandleWrongWay(payloadHex string) error {
  message := "Unrecognized input, you should send a valid input"
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleWrongWay: error making http request: %s", err)
  }
  return fmt.Errorf(message)
}

func main() {
  valuesMap = make(map[string]string)

  appHandler := handler.NewSimpleHandler()
  appHandler.SetDebug()

  // creates a new wallet
  dappWallet = wallet.NewWalletApp(appHandler);
  // setups the dapp relay and fixed portal deposit routes
  //   overrides any fixed address handler
  //   and extra routes to control assets
  dappWallet.SetupRoutes([
    wallet.DepositEtherAdvanceRoute,
    wallet.DepositErc20AdvanceRoute,
    wallet.BalanceInspectRoute,
    wallet.TransferErc20AdvanceRoute,
    wallet.WithdrawEtherAdvanceRoute
  ])

  appHandler.HandleDefault(HandleWrongWay)

  err := appHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}