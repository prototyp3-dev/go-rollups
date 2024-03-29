package wallet

import (
  "os"
  "fmt"
  "math/big"
  "log"
  "encoding/json"
  "github.com/prototyp3-dev/go-rollups/rollups"
  hdl "github.com/prototyp3-dev/go-rollups/handler"
  "github.com/prototyp3-dev/go-rollups/handler/abi"
  "github.com/prototyp3-dev/go-rollups/handler/uri"
)

var DebugLogger *log.Logger

//
// Wallet
//

type Wallet struct {
  Ether *big.Int                                        `json:"ether"`
  Erc20 map[abihandler.Address]*big.Int                 `json:"erc20"`
  Erc721 map[abihandler.Address]map[[32]byte]struct{}   `json:"erc721"`
  Erc1155 map[abihandler.Address]map[[32]byte]*big.Int  `json:"erc1155"`
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
  erc72iMap := make(map[abihandler.Address][]*big.Int)
  for a := range w.Erc721 {
    erc72iMap[a] = w.Erc721TokenIdList(a)
  }
  erc1155Map := make(map[abihandler.Address][2][]*big.Int)
  for a := range w.Erc1155 {
    erc1155Map[a] = w.Erc1155TokenIdList(a)
  }
  return json.Marshal(struct{
    Ether *big.Int                                `json:"ether"`
    Erc20 map[abihandler.Address]*big.Int         `json:"erc20"`
    Erc721 map[abihandler.Address][]*big.Int      `json:"erc721"`
    Erc1155 map[abihandler.Address][2][]*big.Int  `json:"erc1155"`
  }{Ether: w.Ether, Erc20:w.Erc20, Erc721:erc72iMap, Erc1155:erc1155Map})
}

func (w *Wallet) Erc721TokenIdList(tokenAddres abihandler.Address) []*big.Int {
  idList := make([]*big.Int,0)
  for tokenIdBytes := range w.Erc721[tokenAddres] {
    idList = append(idList,new(big.Int).SetBytes(tokenIdBytes[:]))
  }
  return idList
}

func (w *Wallet) Erc1155TokenIdList(tokenAddres abihandler.Address) [2][]*big.Int {
  var idAmountList [2][]*big.Int
  idAmountList[0] = make([]*big.Int,0)
  idAmountList[1] = make([]*big.Int,0)
  for tokenIdBytes, amount := range w.Erc1155[tokenAddres] {
    idAmountList[0] = append(idAmountList[0],new(big.Int).SetBytes(tokenIdBytes[:]))
    idAmountList[1] = append(idAmountList[1],amount)
  }
  return idAmountList
}

func (w *Wallet) DepositEther(amount *big.Int) error {
  w.Ether.Add(w.Ether,amount)
  return nil
}

func (w *Wallet) DepositErc20(tokenAddress abihandler.Address, amount *big.Int) error {
  if w.Erc20[tokenAddress] == nil {
    w.Erc20[tokenAddress] = new(big.Int)
  }
  w.Erc20[tokenAddress].Add(w.Erc20[tokenAddress],amount)
  return nil
}

func (w *Wallet) DepositErc721(tokenAddress abihandler.Address, tokenId *big.Int) error {
  if w.Erc721[tokenAddress] == nil {
    w.Erc721[tokenAddress] = make(map[[32]byte]struct{})
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := w.Erc721[tokenAddress][tokenIdBytes]; ok {
    return fmt.Errorf("Wallet already has id %d for erc721(%s)",tokenId,tokenAddress)
  }
  w.Erc721[tokenAddress][tokenIdBytes] = struct{}{}
  return nil
}

func (w *Wallet) DepositErc1155(tokenAddress abihandler.Address, tokenId *big.Int, amount *big.Int) error {
  if w.Erc1155[tokenAddress] == nil {
    w.Erc1155[tokenAddress] = make(map[[32]byte]*big.Int)
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := w.Erc1155[tokenAddress][tokenIdBytes]; !ok {
    w.Erc1155[tokenAddress][tokenIdBytes] = new(big.Int)
  }
  w.Erc1155[tokenAddress][tokenIdBytes].Add(w.Erc1155[tokenAddress][tokenIdBytes],amount)
  return nil
}

func (w *Wallet) WithdrawEther(amount *big.Int) error {
  if w.Ether.Cmp(amount) == -1 {
    return fmt.Errorf("Wallet has insufficient ether funds")
  }
  w.Ether.Sub(w.Ether,amount)
  return nil
}

func (w *Wallet) WithdrawErc20(tokenAddress abihandler.Address, amount *big.Int) error {
  if w.Erc20[tokenAddress] == nil || w.Erc20[tokenAddress].Cmp(amount) == -1 {
    return fmt.Errorf("Wallet has insufficient erc20(%s) funds",tokenAddress)
  }
  w.Erc20[tokenAddress].Sub(w.Erc20[tokenAddress],amount)
  return nil
}

func (w *Wallet) WithdrawErc721(tokenAddress abihandler.Address, tokenId *big.Int) error {
  if w.Erc721[tokenAddress] == nil {
    return fmt.Errorf("Wallet doesn't have any id for erc721(%s)",tokenAddress)
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := w.Erc721[tokenAddress][tokenIdBytes]; !ok {
    return fmt.Errorf("Wallet doesn't have id %d for erc721(%s)",tokenId,tokenAddress)
  }
  delete(w.Erc721[tokenAddress],tokenIdBytes)
  return nil
}

func (w *Wallet) WithdrawErc1155(tokenAddress abihandler.Address, tokenId *big.Int, amount *big.Int) error {
  if w.Erc1155[tokenAddress] == nil {
    return fmt.Errorf("Wallet doesn't have any id for erc1155%s)",tokenAddress)
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := w.Erc1155[tokenAddress][tokenIdBytes]; !ok {
    return fmt.Errorf("Wallet doesn't have id %d for Erc1155(%s)",tokenId,tokenAddress)
  }
  if w.Erc1155[tokenAddress][tokenIdBytes].Cmp(amount) == -1 {
    return fmt.Errorf("Wallet has insufficient Erc1155(%s) id %d funds",tokenId,tokenAddress)
  }
  w.Erc1155[tokenAddress][tokenIdBytes].Sub(w.Erc1155[tokenAddress][tokenIdBytes],amount)
  return nil
}

//
// WalletApp
//

type WalletRoute uint8

const (
  DappRelayAdvanceRoute WalletRoute = iota
  EtherCodecAdvanceRoutes
  Erc20CodecAdvanceRoutes
  Erc721CodecAdvanceRoutes
  Erc1155CodecAdvanceRoutes
  Erc1155SingleCodecAdvanceRoutes
  Erc1155BatchCodecAdvanceRoutes
  DepositEtherAdvanceRoute
  DepositErc20AdvanceRoute
  DepositErc721AdvanceRoute
  DepositErc1155SingleAdvanceRoute
  DepositErc1155BatchAdvanceRoute
  DepositErc1155AdvanceRoute
  WithdrawEtherAdvanceRoute
  WithdrawErc20AdvanceRoute
  WithdrawErc721AdvanceRoute
  WithdrawErc1155SingleAdvanceRoute
  WithdrawErc1155BatchAdvanceRoute
  WithdrawErc1155AdvanceRoute
  TransferEtherAdvanceRoute
  TransferErc20AdvanceRoute
  TransferErc721AdvanceRoute
  TransferErc1155SingleAdvanceRoute
  TransferErc1155BatchAdvanceRoute
  TransferErc1155AdvanceRoute
  BalanceInspectRoute
  WithdrawEtherCodecAdvanceRoute
  WithdrawErc20CodecAdvanceRoute
  WithdrawErc721CodecAdvanceRoute
  WithdrawErc1155SingleCodecAdvanceRoute
  WithdrawErc1155BatchCodecAdvanceRoute
  WithdrawErc1155CodecAdvanceRoute
  TransferEtherCodecAdvanceRoute
  TransferErc20CodecAdvanceRoute
  TransferErc721CodecAdvanceRoute
  TransferErc1155SingleCodecAdvanceRoute
  TransferErc1155BatchCodecAdvanceRoute
  TransferErc1155CodecAdvanceRoute
  BalanceCodecInspectRoute
  BalanceUriInspectRoute
)

type WalletApp struct {
  handler *hdl.Handler
  abiHandler *abihandler.AbiHandler
  uriHandler *urihandler.UriHandler
  DappAddress abihandler.Address
  Wallets map[abihandler.Address]*Wallet
}

func (w *WalletApp) GetWallet(address abihandler.Address) *Wallet {
  if w.Wallets[address] == nil {
    w.Wallets[address] = &Wallet{Ether: new(big.Int), Erc20: make(map[abihandler.Address]*big.Int), 
      Erc721: make(map[abihandler.Address]map[[32]byte]struct{}), Erc1155: make(map[abihandler.Address]map[[32]byte]*big.Int)}
  }
  return w.Wallets[address]
}

var etherNoticeCodec *abihandler.Codec
var erc20NoticeCodec *abihandler.Codec
var erc721NoticeCodec *abihandler.Codec
var erc1155NoticeCodec *abihandler.Codec

var erc1155BatchValueCodec *abihandler.Codec

var etherVoucherCodec *abihandler.Codec
var erc20VoucherCodec *abihandler.Codec
var erc721VoucherCodec *abihandler.Codec
var erc1155SingleVoucherCodec *abihandler.Codec
var erc1155BatchVoucherCodec *abihandler.Codec

func NewWalletApp() *WalletApp {
  if hdl.KnownRollupsAddresses == nil {
		panic("NewWalletApp: Rollups Addresses not initialized")
  }

  DebugLogger = log.New(os.Stderr, "[ debug ] ", log.Lshortfile)

  etherNoticeCodec = abihandler.NewCodec([]string{"address","int256","uint256"}) // address, amount, balance
  erc20NoticeCodec = abihandler.NewCodec([]string{"address","address","int256","uint256"}) // address, tokenAddress, amount, balance
  erc721NoticeCodec = abihandler.NewCodec([]string{"address","address","int256","uint256[]"}) // address, tokenAddress, tokenId, tokenIdList
  erc1155NoticeCodec = abihandler.NewCodec([]string{"address","address","int256[]","int256[]","uint256[]","uint256[]"}) // address, tokenAddress, tokenIdList, amountList, idBalance, amountsBalance

  erc1155BatchValueCodec = abihandler.NewCodec([]string{"uint256[]","uint256[]","bytes","bytes"}) // tokenIdList, amountList, base data, eec data

  etherVoucherCodec = abihandler.NewVoucherCodec("withdrawEther",[]string{"address","uint256"}) // receiver, balance
  erc20VoucherCodec = abihandler.NewVoucherCodec("transfer",[]string{"address","uint256"}) // receiver, amount
  erc721VoucherCodec = abihandler.NewVoucherCodec("safeTransferFrom",[]string{"address","address","uint256"}) // sender, receiver, tokenId
  erc1155SingleVoucherCodec = abihandler.NewVoucherCodec("safeTransferFrom",[]string{"address","address","uint256","uint256","bytes"}) // sender, receiver, tokenId, amount, data
  erc1155BatchVoucherCodec = abihandler.NewVoucherCodec("safeBatchTransferFrom",[]string{"address","address","uint256[]","uint256[]","bytes"}) // sender, receiver, tokenIds, amounts, data

  app := &WalletApp{Wallets: make(map[abihandler.Address]*Wallet)}

  return app
}

func (w *WalletApp) SetHandler(handler *hdl.Handler) {
  if handler == nil {
    panic("Nil handler")
  }
  w.handler = handler
}

func (w *WalletApp) SetAbiHandler(abiHdl *abihandler.AbiHandler) {
  if abiHdl == nil {
    panic("Nil handler")
  }
  if w.handler == nil {
    w.SetHandler(abiHdl.Handler)
  }
  w.abiHandler = abiHdl
}

func (w *WalletApp) SetUriHandler(uriHdl *urihandler.UriHandler) {
  if uriHdl == nil {
    panic("Nil handler")
  }
  if w.handler == nil {
    w.SetHandler(uriHdl.Handler)
  }
  w.uriHandler = uriHdl
}

func (w *WalletApp) Handler() *hdl.Handler {
  if w.handler == nil {
    w.SetHandler(hdl.NewSimpleHandler())
  }
  return w.handler
}

func (w *WalletApp) AbiHandler() *abihandler.AbiHandler {
  if w.abiHandler == nil {
    w.SetAbiHandler(abihandler.AddAbiHandler(w.Handler()))
  }
  return w.abiHandler
}

func (w *WalletApp) UriHandler() *urihandler.UriHandler {
  if w.uriHandler == nil {
    w.SetUriHandler(urihandler.AddUriHandler(w.Handler()))
  }
  return w.uriHandler
}

func (w *WalletApp) SetupRoutes(routes []WalletRoute) {

  var forceRelayRoute bool
  var relayRouteAdded bool

  for _, route := range routes {
    switch route {
    case DappRelayAdvanceRoute:
      relayRouteAdded = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.DappAddressRelay, abihandler.NewPackedCodec([]string{"address"}), w.HandleRelay)
    case EtherCodecAdvanceRoutes:
      forceRelayRoute = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.EtherPortalAddress, abihandler.NewPackedCodec([]string{"address","uint256","bytes"}), w.EtherPortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherWithdraw",[]string{"uint256","bytes"}), w.EtherWithdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherTransfer",[]string{"address","uint256","bytes"}), w.TransferEtherCodec)
    case DepositEtherAdvanceRoute:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.EtherPortalAddress, abihandler.NewPackedCodec([]string{"address","uint256","bytes"}), w.EtherPortalDeposit)
    case Erc20CodecAdvanceRoutes:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc20PortalAddress, abihandler.NewPackedCodec([]string{"bool","address","address","uint256","bytes"}), w.Erc20PortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc20Withdraw",[]string{"address","uint256","bytes"}), w.Erc20Withdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc20Transfer",[]string{"address","address","uint256","bytes"}), w.TransferErc20Codec)
    case DepositErc20AdvanceRoute:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc20PortalAddress, abihandler.NewPackedCodec([]string{"bool","address","address","uint256","bytes"}), w.Erc20PortalDeposit)
    case Erc721CodecAdvanceRoutes:
      forceRelayRoute = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc721PortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","bytes"}), w.Erc721PortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc721Withdraw",[]string{"address","uint256","bytes"}), w.Erc721Withdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc721Transfer",[]string{"address","address","uint256","bytes"}), w.TransferErc721Codec)
    case DepositErc721AdvanceRoute:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc721PortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","bytes"}), w.Erc721PortalDeposit)
    case Erc1155CodecAdvanceRoutes:
      forceRelayRoute = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155SinglePortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","uint256","bytes"}), w.Erc1155SinglePortalDeposit)
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155BatchPortalAddress, abihandler.NewPackedCodec([]string{"address","address","bytes"}), w.Erc1155BatchPortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchWithdraw",[]string{"address","uint256[]","uint256[]","bytes"}), w.Erc1155BatchWithdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleWithdraw",[]string{"address","uint256","uint256","bytes"}), w.Erc1155SingleWithdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleTransfer",[]string{"address","address","uint256","uint256","bytes"}), w.TransferErc1155SingleCodec)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchTransfer",[]string{"address","address","uint256[]","uint256[]","bytes"}), w.TransferErc1155BatchCodec)
    case Erc1155SingleCodecAdvanceRoutes:
      forceRelayRoute = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155SinglePortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","uint256","bytes"}), w.Erc1155SinglePortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleWithdraw",[]string{"address","uint256","uint256","bytes"}), w.Erc1155SingleWithdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleTransfer",[]string{"address","address","uint256","uint256","bytes"}), w.TransferErc1155SingleCodec)
    case DepositErc1155SingleAdvanceRoute:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155SinglePortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","uint256","bytes"}), w.Erc1155SinglePortalDeposit)
    case Erc1155BatchCodecAdvanceRoutes:
      forceRelayRoute = true
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155BatchPortalAddress, abihandler.NewPackedCodec([]string{"address","address","bytes"}), w.Erc1155BatchPortalDeposit)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchWithdraw",[]string{"address","uint256[]","uint256[]","bytes"}), w.Erc1155BatchWithdraw)
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchTransfer",[]string{"address","address","uint256[]","uint256[]","bytes"}), w.TransferErc1155BatchCodec)
    case DepositErc1155AdvanceRoute,DepositErc1155BatchAdvanceRoute:
      w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc1155BatchPortalAddress, abihandler.NewPackedCodec([]string{"address","address","bytes"}), w.Erc1155BatchPortalDeposit)
    case WithdrawEtherAdvanceRoute,WithdrawEtherCodecAdvanceRoute:
      forceRelayRoute = true
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherWithdraw",[]string{"uint256","bytes"}), w.EtherWithdraw)
    case WithdrawErc20AdvanceRoute,WithdrawErc20CodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc20Withdraw",[]string{"address","uint256","bytes"}), w.Erc20Withdraw)
    case WithdrawErc721AdvanceRoute,WithdrawErc721CodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc721Withdraw",[]string{"address","uint256","bytes"}), w.Erc721Withdraw)
      forceRelayRoute = true
    case WithdrawErc1155SingleAdvanceRoute,WithdrawErc1155SingleCodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleWithdraw",[]string{"address","uint256","uint256","bytes"}), w.Erc1155SingleWithdraw)
      forceRelayRoute = true
    case WithdrawErc1155AdvanceRoute,WithdrawErc1155BatchAdvanceRoute,WithdrawErc1155BatchCodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchWithdraw",[]string{"address","uint256[]","uint256[]","bytes"}), w.Erc1155BatchWithdraw)
      forceRelayRoute = true
    case TransferEtherAdvanceRoute,TransferEtherCodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherTransfer",[]string{"address","uint256","bytes"}), w.TransferEtherCodec)
    case TransferErc20AdvanceRoute,TransferErc20CodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc20Transfer",[]string{"address","address","uint256","bytes"}), w.TransferErc20Codec)
    case TransferErc721AdvanceRoute,TransferErc721CodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc721Transfer",[]string{"address","address","uint256","bytes"}), w.TransferErc721Codec)
    case TransferErc1155SingleAdvanceRoute,TransferErc1155SingleCodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155SingleTransfer",[]string{"address","address","uint256","uint256","bytes"}), w.TransferErc1155SingleCodec)
    case TransferErc1155AdvanceRoute,TransferErc1155BatchAdvanceRoute,TransferErc1155BatchCodecAdvanceRoute:
      w.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc1155BatchTransfer",[]string{"address","address","uint256[]","uint256[]","bytes"}), w.TransferErc1155BatchCodec)
    case BalanceInspectRoute,BalanceCodecInspectRoute:
      w.AbiHandler().HandleInspectRoute(abihandler.NewHeaderCodec("wallet","Balance",[]string{"address address"}), w.BalanceAbi)
    case BalanceUriInspectRoute:
      w.UriHandler().HandleInspectRoute("/balance/:address", w.BalanceUri)
    default:
      panic("Unrecognized route")
    }
  }
  if forceRelayRoute && !relayRouteAdded {
    w.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.DappAddressRelay, abihandler.NewPackedCodec([]string{"address"}), w.HandleRelay)
  }
}

//
// Relay
//

func (w *WalletApp) HandleRelay(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  addr, ok1 := payloadMap["0"].(abihandler.Address)

  if !ok1 {
    message := "HandleRelay: parameters error"
    return fmt.Errorf(message)
  }

  w.DappAddress = addr

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Dapp Relay, dapp address is", addr)}

  return nil
}

//
// Deposit
//

func (w *WalletApp) EtherPortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  depositor, ok1 := payloadMap["0"].(abihandler.Address)
  amount, ok2 := payloadMap["1"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)

  if !ok1 || !ok2 {
    message := "EtherPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := w.GetWallet(depositor)

  // Deposit
  err := wallet.DepositEther(amount)
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: error adding funds: %s", err)
  }
  
  // Notice
  noticePayload,err := etherNoticeCodec.Encode([]interface{}{depositor,amount,wallet.Ether})
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: encoding notice: %s", err)
  }
  notice := rollups.Notice{Payload:noticePayload}
  _, err = rollups.SendNotice(&notice)
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amount,"native token deposit from",depositor)}
  
  return nil
}

func (w *WalletApp) Erc20PortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {

  // ret, ok1 := payloadMap["0"].(bool)
  tokenAddress, ok2 := payloadMap["1"].(abihandler.Address)
  depositor, ok3 := payloadMap["2"].(abihandler.Address)
  amount, ok4 := payloadMap["3"].(*big.Int)
  // dataBytes, ok5 := payloadMap["4"].([]byte)

  if !ok2 || !ok3 || !ok4 {
    message := "Erc20PortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := w.GetWallet(depositor)

  // Deposit
  err := wallet.DepositErc20(tokenAddress, amount)
  if err != nil {
    return fmt.Errorf("Erc20PortalDeposit: error adding funds: %s", err)
  }
  
  // Notice
  noticePayload,err := erc20NoticeCodec.Encode([]interface{}{depositor,tokenAddress,amount,wallet.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amount,"tokens",tokenAddress,"Erc20 deposit from",depositor)}

  return nil
}

func (w *WalletApp) Erc721PortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  depositor, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok4 := payloadMap["3"].([]byte)

  if !ok1 || !ok2 || !ok3 {
    message := "Erc721PortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := w.GetWallet(depositor)

  // Deposit
  err := wallet.DepositErc721(tokenAddress, tokenId)
  if err != nil {
    return fmt.Errorf("Erc721PortalDeposit: error adding id: %s", err)
  }
  
  // Notice
  noticePayload,err := erc721NoticeCodec.Encode([]interface{}{depositor,tokenAddress,tokenId,wallet.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("Erc721PortalDeposit: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc721PortalDeposit: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received id",tokenId,tokenAddress,"Erc721 deposit from",depositor)}

  return nil
}

func (w *WalletApp) Erc1155SinglePortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  depositor, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  amount, ok4 := payloadMap["3"].(*big.Int)
  // dataBytes, ok4 := payloadMap["3"].([]byte)

  if !ok1 || !ok2 || !ok3 || !ok4 {
    message := "Erc1155SinglePortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := w.GetWallet(depositor)

  // Deposit
  err := wallet.DepositErc1155(tokenAddress, tokenId, amount)
  if err != nil {
    return fmt.Errorf("Erc1155SinglePortalDeposit: error adding id: %s", err)
  }
  
  // Notice
  tokenBalance := wallet.Erc1155TokenIdList(tokenAddress)
  noticePayload,err := erc1155NoticeCodec.Encode([]interface{}{depositor,tokenAddress,[]*big.Int{tokenId},[]*big.Int{amount},tokenBalance[0],tokenBalance[1]})
  if err != nil {
    return fmt.Errorf("Erc1155SinglePortalDeposit: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc1155SinglePortalDeposit: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amount,"tokens from id",tokenId,tokenAddress,"Erc1155 deposit from",depositor)}

  return nil
}

func (w *WalletApp) Erc1155BatchPortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  depositor, ok2 := payloadMap["1"].(abihandler.Address)
  valueBytes, ok3 := payloadMap["2"].([]byte)

  if !ok1 || !ok2 || !ok3 {
    message := "Erc1155BatchPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  valueMap, err := erc1155BatchValueCodec.Decode(rollups.Bin2Hex(valueBytes))
  if err != nil {
    return fmt.Errorf("Erc1155BatchPortalDeposit: parameters error: %s", err)
  }
  tokenIds, ok4 := valueMap["0"].([]*big.Int)
  amounts, ok5 := valueMap["1"].([]*big.Int)
  
  if !ok4 || !ok5 {
    message := "Erc1155BatchPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  if len(tokenIds) != len(amounts) {
    message := "Erc1155BatchPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }
  numTokens := len(tokenIds)

  wallet := w.GetWallet(depositor)

  // Deposit
  for i := 0 ; i < numTokens ; i++ {
    err = wallet.DepositErc1155(tokenAddress, tokenIds[i], amounts[i])
    if err != nil {
      return fmt.Errorf("Erc1155BatchPortalDeposit: error adding id: %s", err)
    }
  }

  // Notice
  tokenBalance := wallet.Erc1155TokenIdList(tokenAddress)
  noticePayload,err := erc1155NoticeCodec.Encode([]interface{}{depositor,tokenAddress,tokenIds,amounts,tokenBalance[0],tokenBalance[1]})
  if err != nil {
    return fmt.Errorf("Erc1155BatchPortalDeposit: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc1155BatchPortalDeposit: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amounts,"tokens from ids",tokenIds,tokenAddress,"Erc1155 deposit from",depositor)}

  return nil
}

//
// Withdraw
//

func (w *WalletApp) EtherWithdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if w.DappAddress == (abihandler.Address{}) {
    return fmt.Errorf("EtherWithdraw: Can not generate voucher as there is no dapp address configured")
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("EtherWithdraw: payload:",payloadMap)}
  amount, ok1 := payloadMap["0"].(*big.Int)
  dataBytes, ok2 := payloadMap["1"].([]byte)

  if !ok1 || !ok2 {
    message := "EtherWithdraw: parameters error"
    return fmt.Errorf(message)
  }

  addr,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("EtherWithdraw: error converting address: %s", err)
  }

  wallet := w.GetWallet(addr)

  // Withdrawal
  err = wallet.WithdrawEther(amount)
  if err != nil {
    return fmt.Errorf("EtherWithdraw: error withdrawing Ether: %s", err)
  }

  // Voucher
  voucherPayload,err := etherVoucherCodec.Encode([]interface{}{addr,amount})
  if err != nil {
    return fmt.Errorf("EtherWithdraw: encoding voucher: %s", err)
  }
  _, err = w.handler.SendVoucher(w.DappAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := etherNoticeCodec.Encode([]interface{}{addr,new(big.Int).Neg(amount),wallet.Ether})
  if err != nil {
    return fmt.Errorf("EtherWithdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("EtherWithdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amount,"ETH from",addr,"data:",dataBytes)}

  return nil
}

func (w *WalletApp) Erc20Withdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("EtherWithdraw: payload:",payloadMap)}
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  amount, ok2 := payloadMap["1"].(*big.Int)
  dataBytes, ok3 := payloadMap["2"].([]byte)

  if !ok1 || !ok2 || !ok3 {
    message := "Erc20Withdraw: parameters error"
    return fmt.Errorf(message)
  }

  addr,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error converting address: %s", err)
  }

  wallet := w.GetWallet(addr)

  // Withdrawal
  err = wallet.WithdrawErc20(tokenAddress,amount)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error withdrawing Erc20: %s", err)
  }
    
  // Voucher
  voucherPayload,err := erc20VoucherCodec.Encode([]interface{}{addr,amount})
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: encoding voucher: %s", err)
  }
  _, err = w.handler.SendVoucher(tokenAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := erc20NoticeCodec.Encode([]interface{}{addr,tokenAddress,new(big.Int).Neg(amount),wallet.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amount,"of Erc20",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

func (w *WalletApp) Erc721Withdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Erc721Withdraw: payload:",payloadMap)}

  if w.DappAddress == (abihandler.Address{}) {
    return fmt.Errorf("Erc721Withdraw: Can not generate voucher as there is no dapp address configured")
  }

  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  tokenId, ok2 := payloadMap["1"].(*big.Int)
  dataBytes, ok3 := payloadMap["2"].([]byte)

  if !ok1 || !ok2 || !ok3 {
    message := "Erc721Withdraw: parameters error"
    return fmt.Errorf(message)
  }

  addr,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error converting address: %s", err)
  }

  wallet := w.GetWallet(addr)

  // Withdrawal
  err = wallet.WithdrawErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error withdrawing Erc721: %s", err)
  }
    
  // Voucher
  voucherPayload,err := erc721VoucherCodec.Encode([]interface{}{w.DappAddress,addr,tokenId})
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: encoding voucher: %s", err)
  }
  _, err = w.handler.SendVoucher(tokenAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := erc721NoticeCodec.Encode([]interface{}{addr,tokenAddress,new(big.Int).Neg(tokenId),wallet.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn id",tokenId,"of Erc721",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

func (w *WalletApp) Erc1155SingleWithdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Erc1155SingleWithdraw: payload:",payloadMap)}

  if w.DappAddress == (abihandler.Address{}) {
    return fmt.Errorf("Erc1155SingleWithdraw: Can not generate voucher as there is no dapp address configured")
  }

  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  tokenId, ok2 := payloadMap["1"].(*big.Int)
  amount, ok3 := payloadMap["2"].(*big.Int)
  dataBytes, ok4 := payloadMap["3"].([]byte)

  if !ok1 || !ok2 || !ok3 || !ok4 {
    message := "Erc1155SingleWithdraw: parameters error"
    return fmt.Errorf(message)
  }

  addr,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: error converting address: %s", err)
  }

  wallet := w.GetWallet(addr)

  // Withdrawal
  err = wallet.WithdrawErc1155(tokenAddress,tokenId,amount)
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: error withdrawing Erc721: %s", err)
  }

  // Voucher
  voucherPayload,err := erc1155SingleVoucherCodec.Encode([]interface{}{w.DappAddress,addr,tokenId,amount,[]byte{}})
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: encoding voucher: %s", err)
  }
  _, err = w.handler.SendVoucher(tokenAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: error making http request: %s", err)
  }

  // Notice
  tokenBalance := wallet.Erc1155TokenIdList(tokenAddress)
  noticePayload,err := erc1155NoticeCodec.Encode([]interface{}{addr,tokenAddress,[]*big.Int{new(big.Int).Neg(tokenId)},[]*big.Int{new(big.Int).Neg(amount)},tokenBalance[0],tokenBalance[1]})
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc1155SingleWithdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amount,"tokens of id",tokenId,"of Erc1155",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

func (w *WalletApp) Erc1155BatchWithdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Erc1155SingleWithdraw: payload:",payloadMap)}

  if w.DappAddress == (abihandler.Address{}) {
    return fmt.Errorf("Erc1155BatchWithdraw: Can not generate voucher as there is no dapp address configured")
  }

  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  tokenIds, ok2 := payloadMap["1"].([]*big.Int)
  amounts, ok3 := payloadMap["2"].([]*big.Int)
  dataBytes, ok4 := payloadMap["3"].([]byte)

  if !ok1 || !ok2 || !ok3 || !ok4 {
    message := "Erc1155BatchWithdraw: parameters error"
    return fmt.Errorf(message)
  }

  if len(tokenIds) != len(amounts) {
    message := "Erc1155BatchPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  addr,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("Erc1155BatchWithdraw: error converting address: %s", err)
  }

  wallet := w.GetWallet(addr)

  // Withdrawal
  negAmounts := make([]*big.Int,0)
  negIds := make([]*big.Int,0)
  numTokens := len(tokenIds)
  for i := 0 ; i < numTokens ; i++ {
    err = wallet.WithdrawErc1155(tokenAddress, tokenIds[i], amounts[i])
    if err != nil {
      return fmt.Errorf("Erc1155BatchPortalDeposit: error adding id: %s", err)
    }
    negIds = append(negIds,new(big.Int).Neg(tokenIds[i]))
    negAmounts = append(negAmounts,new(big.Int).Neg(amounts[i]))
  }

  // Voucher
  voucherPayload,err := erc1155BatchVoucherCodec.Encode([]interface{}{w.DappAddress,addr,tokenIds,amounts,[]byte{}})
  if err != nil {
    return fmt.Errorf("Erc1155BatchWithdraw: encoding voucher: %s", err)
  }
  _, err = w.handler.SendVoucher(tokenAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc1155BatchWithdraw: error making http request: %s", err)
  }

  // Notice
  tokenBalance := wallet.Erc1155TokenIdList(tokenAddress)
  noticePayload,err := erc1155NoticeCodec.Encode([]interface{}{addr,tokenAddress,negIds,negAmounts,tokenBalance[0],tokenBalance[1]})
  if err != nil {
    return fmt.Errorf("Erc1155BatchWithdraw: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc1155BatchWithdraw: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amounts,"tokens of ids",tokenIds,"of Erc1155",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

//
// Transfer
//

func (w *WalletApp) TransferEtherCodec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  receiver, ok1 := payloadMap["0"].(abihandler.Address)
  amount, ok2 := payloadMap["1"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 {
    message := "TransferEtherCodec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferEtherCodec: error converting address: %s", err)
  }

  return w.TransferEther(sender,receiver,amount)
}

func (w *WalletApp) TransferEther(sender abihandler.Address, receiver abihandler.Address, amount *big.Int) error {

  walletSender := w.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawEther(amount)
  if err != nil {
    return fmt.Errorf("TransferEther: error withdrawing Ether: %s", err)
  }

  walletReceiver := w.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositEther(amount)
  if err != nil {
    return fmt.Errorf("TransferEther: error depositing Ether: %s", err)
  }

  // Notice
  noticePayload,err := etherNoticeCodec.Encode([]interface{}{sender,new(big.Int).Neg(amount),walletSender.Ether})
  if err != nil {
    return fmt.Errorf("TransferEther: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferEther: error making http request: %s", err)
  }

  noticePayload,err = etherNoticeCodec.Encode([]interface{}{receiver,amount,walletReceiver.Ether})
  if err != nil {
    return fmt.Errorf("TransferEther: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferEther: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",amount,"ETH from",sender,"to",receiver)}

  return nil
}

func (w *WalletApp) TransferErc20Codec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  amount, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 {
    message := "TransferErc20Codec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErc20Codec: error converting address: %s", err)
  }

  return w.TransferErc20(tokenAddress,sender,receiver,amount)
}

func (w *WalletApp) TransferErc20(tokenAddress abihandler.Address, sender abihandler.Address, receiver abihandler.Address, amount *big.Int) error {

  walletSender := w.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawErc20(tokenAddress,amount)
  if err != nil {
    return fmt.Errorf("TransferErc20: error withdrawing Erc20: %s", err)
  }

  walletReceiver := w.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositErc20(tokenAddress,amount)
  if err != nil {
    return fmt.Errorf("TransferErc20: error depositing Erc20: %s", err)
  }

  // Notice
  noticePayload,err := erc20NoticeCodec.Encode([]interface{}{tokenAddress,sender,new(big.Int).Neg(amount),walletSender.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("TransferErc20: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc20: error making http request: %s", err)
  }

  noticePayload,err = erc20NoticeCodec.Encode([]interface{}{tokenAddress,receiver,amount,walletReceiver.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("TransferErc20: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc20: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",amount,"Erc20",tokenAddress,"from",sender,"to",receiver)}

  return nil
}

func (w *WalletApp) TransferErc721Codec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 {
    message := "TransferErc721Codec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErc721Codec: error converting address: %s", err)
  }

  return w.TransferErc721(tokenAddress,sender,receiver,tokenId)
}

func (w *WalletApp) TransferErc721(tokenAddress abihandler.Address, sender abihandler.Address, receiver abihandler.Address, tokenId *big.Int) error {

  walletSender := w.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("TransferErc721: error withdrawing Erc721: %s", err)
  }

  walletReceiver := w.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("TransferErc721: error depositing Erc721: %s", err)
  }

  // Notice
  noticePayload,err := erc721NoticeCodec.Encode([]interface{}{tokenAddress,sender,new(big.Int).Neg(tokenId),walletSender.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("TransferErc721: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc721: error making http request: %s", err)
  }

  noticePayload,err = erc721NoticeCodec.Encode([]interface{}{tokenAddress,receiver,tokenId,walletReceiver.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("TransferErc721: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc721: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",tokenId,"Erc721",tokenAddress,"from",sender,"to",receiver)}

  return nil
}

func (w *WalletApp) TransferErc1155SingleCodec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  amount, ok4 := payloadMap["3"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 || !ok4 {
    message := "TransferErc1155SingleCodec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErc1155SingleCodec: error converting address: %s", err)
  }

  return w.TransferErc1155Batch(tokenAddress,sender,receiver,[]*big.Int{tokenId},[]*big.Int{amount})
}

func (w *WalletApp) TransferErc1155BatchCodec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  tokenIds, ok3 := payloadMap["2"].([]*big.Int)
  amounts, ok4 := payloadMap["3"].([]*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 || !ok4 {
    message := "TransferErc1155BatchCodec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErc1155BatchCodec: error converting address: %s", err)
  }

  return w.TransferErc1155Batch(tokenAddress,sender,receiver,tokenIds,amounts)
}

func (w *WalletApp) TransferErc1155Batch(tokenAddress abihandler.Address, sender abihandler.Address, receiver abihandler.Address, tokenIds []*big.Int, amounts []*big.Int) error {

  numTokens := len(tokenIds)

  // Withdrawal
  negAmounts := make([]*big.Int,0)
  negIds := make([]*big.Int,0)
  walletSender := w.GetWallet(sender)
  for i := 0 ; i < numTokens ; i++ {
    err := walletSender.WithdrawErc1155(tokenAddress, tokenIds[i], amounts[i])
    if err != nil {
      return fmt.Errorf("TransferErc1155: error adding id: %s", err)
    }
    negIds = append(negIds,new(big.Int).Neg(tokenIds[i]))
    negAmounts = append(negAmounts,new(big.Int).Neg(amounts[i]))
  }

  // Deposit
  walletReceiver := w.GetWallet(receiver)
  for i := 0 ; i < numTokens ; i++ {
    err := walletReceiver.DepositErc1155(tokenAddress, tokenIds[i], amounts[i])
    if err != nil {
      return fmt.Errorf("Erc1155BatchPortalDeposit: error adding id: %s", err)
    }
  }

  // Notice
  tokenBalanceSender := walletSender.Erc1155TokenIdList(tokenAddress)
  noticePayload,err := erc1155NoticeCodec.Encode([]interface{}{sender,tokenAddress,negIds,negAmounts,tokenBalanceSender[0],tokenBalanceSender[1]})
  if err != nil {
    return fmt.Errorf("TransferErc1155: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc1155: error making http request: %s", err)
  }

  tokenBalanceReceiver := walletReceiver.Erc1155TokenIdList(tokenAddress)
  noticePayload,err = erc1155NoticeCodec.Encode([]interface{}{receiver,tokenAddress,tokenIds,amounts,tokenBalanceReceiver[0],tokenBalanceReceiver[1]})
  if err != nil {
    return fmt.Errorf("TransferErc1155: encoding notice: %s", err)
  }
  _, err = w.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc1155: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",amounts,"tokens with ids",tokenIds,"Erc155",tokenAddress,"from",sender,"to",receiver)}

  return nil
}

//
// Balance
//

func (w *WalletApp) BalanceAbi(payloadMap map[string]interface{}) error {
  addr, ok1 := payloadMap["address"].(abihandler.Address)
  if !ok1 {
    message := "Balance: parameters error"
    return fmt.Errorf("%s",message)
  }

  wallet := w.GetWallet(addr)

  balanceJson, err := json.Marshal(wallet)
  if err != nil {
    return fmt.Errorf("balance: error converting wallet to json: %s", err)
  }

  err = w.handler.SendReport(rollups.Str2Hex(string(balanceJson)))
  if err != nil {
    return fmt.Errorf("balance: error making http request: %s", err)
  }

  if w.handler.LogLevel >= hdl.Debug {DebugLogger.Println(addr,"balance",string(balanceJson))}

  return nil
}

func (w *WalletApp) BalanceUri(payloadMap map[string]interface{}) error {
  addrStr, ok1 := payloadMap["address"].(string)
  if !ok1 {
    return fmt.Errorf("balance: parameters error")
  }

  addr,err := abihandler.Hex2Address(addrStr)
  if err != nil {
    return fmt.Errorf("balance: parameters error: %s", err)
  }

  return w.BalanceAbi(map[string]interface{}{"address":addr})
}
