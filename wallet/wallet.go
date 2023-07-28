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
}

func (this *Wallet) MarshalJSON() ([]byte, error) {
  erc72iMap := make(map[abihandler.Address][]*big.Int)
  for a, _ := range this.Erc721 {
    erc72iMap[a] = this.Erc721TokenIdList(a)
  }
  return json.Marshal(struct{
    Ether *big.Int                            `json:"ether"`
    Erc20 map[abihandler.Address]*big.Int     `json:"erc20"`
    Erc721 map[abihandler.Address][]*big.Int  `json:"erc721"`
  }{Ether: this.Ether, Erc20:this.Erc20, Erc721:erc72iMap})
}

func (this *Wallet) Erc721TokenIdList(tokenAddres abihandler.Address) []*big.Int {
  idList := make([]*big.Int,0)
  for tokenIdBytes, _ := range this.Erc721[tokenAddres] {
    idList = append(idList,new(big.Int).SetBytes(tokenIdBytes[:]))
  }
  return idList
}

func (this *Wallet) DepositEther(amount *big.Int) error {
  this.Ether.Add(this.Ether,amount)
  return nil
}

func (this *Wallet) DepositErc20(tokenAddress abihandler.Address, amount *big.Int) error {
  if this.Erc20[tokenAddress] == nil {
    this.Erc20[tokenAddress] = new(big.Int)
  }
  this.Erc20[tokenAddress].Add(this.Erc20[tokenAddress],amount)
  return nil
}

func (this *Wallet) DepositErc721(tokenAddress abihandler.Address, tokenId *big.Int) error {
  if this.Erc721[tokenAddress] == nil {
    this.Erc721[tokenAddress] = make(map[[32]byte]struct{})
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := this.Erc721[tokenAddress][tokenIdBytes]; ok {
    return fmt.Errorf("Wallet already has id %d for erc721(%s)",tokenId,tokenAddress)
  }
  this.Erc721[tokenAddress][tokenIdBytes] = struct{}{}
  return nil
}

func (this *Wallet) WithdrawEther(amount *big.Int) error {
  if this.Ether.Cmp(amount) == -1 {
    return fmt.Errorf("Wallet has insufficient ether funds")
  }
  this.Ether.Sub(this.Ether,amount)
  return nil
}

func (this *Wallet) WithdrawErc20(tokenAddress abihandler.Address, amount *big.Int) error {
  if this.Erc20[tokenAddress] == nil || this.Erc20[tokenAddress].Cmp(amount) == -1 {
    return fmt.Errorf("Wallet has insufficient erc20(%s) funds",tokenAddress)
  }
  this.Erc20[tokenAddress].Sub(this.Erc20[tokenAddress],amount)
  return nil
}

func (this *Wallet) WithdrawErc721(tokenAddress abihandler.Address, tokenId *big.Int) error {
  if this.Erc721[tokenAddress] == nil {
    return fmt.Errorf("Wallet doesn't have any id for erc721(%s)",tokenAddress)
  }
  var tokenIdBytes [32]byte
  buf := make([]byte,32,32)
  buf = tokenId.FillBytes(buf)
  copy(tokenIdBytes[:],buf[:32])
  if _, ok := this.Erc721[tokenAddress][tokenIdBytes]; !ok {
    return fmt.Errorf("Wallet doesn't have id %d for erc721(%s)",tokenId,tokenAddress)
  }
  delete(this.Erc721[tokenAddress],tokenIdBytes)
  return nil
}

//
// WalletApp
//

type WalletRoute uint8

const (
  DappRelayAdvanceRoute WalletRoute = iota
  DepositEtherAdvanceRoute
  DepositErc20AdvanceRoute
  DepositErc721AdvanceRoute
  WithdrawEtherAdvanceRoute
  WithdrawErc20AdvanceRoute
  WithdrawErc721AdvanceRoute
  TransferEtherAdvanceRoute
  TransferErc20AdvanceRoute
  TransferErc721AdvanceRoute
  BalanceInspectRoute
  WithdrawEtherCodecAdvanceRoute
  WithdrawErc20CodecAdvanceRoute
  WithdrawErc721CodecAdvanceRoute
  TransferEtherCodecAdvanceRoute
  TransferErc20CodecAdvanceRoute
  TransferErc721CodecAdvanceRoute
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

func (this *WalletApp) GetWallet(address abihandler.Address) *Wallet {
  if this.Wallets[address] == nil {
    this.Wallets[address] = &Wallet{Ether: new(big.Int), Erc20: make(map[abihandler.Address]*big.Int), Erc721: make(map[abihandler.Address]map[[32]byte]struct{})}
  }
  return this.Wallets[address]
}

var etherNoticeCodec *abihandler.Codec
var erc20NoticeCodec *abihandler.Codec
var erc721NoticeCodec *abihandler.Codec

var etherVoucherCodec *abihandler.Codec
var erc20VoucherCodec *abihandler.Codec
var erc721VoucherCodec *abihandler.Codec

func NewWalletApp() *WalletApp {
  if hdl.KnownRollupsAddresses == nil {
		panic("NewWalletApp: Rollups Addresses not initialized")
  }

  DebugLogger = log.New(os.Stderr, "[ debug ] ", log.Lshortfile)

  etherNoticeCodec = abihandler.NewCodec([]string{"address","uint256","uint256"}) // address, amount, balance
  erc20NoticeCodec = abihandler.NewCodec([]string{"address","address","uint256","uint256"}) // address, tokenAddress, amount, balance
  erc721NoticeCodec = abihandler.NewCodec([]string{"address","address","uint256","uint256[]"}) // address, tokenAddress, tokenId, tokenIdList

  etherVoucherCodec = abihandler.NewVoucherCodec("withdrawEther",[]string{"address","uint256"}) // receiver, balance
  erc20VoucherCodec = abihandler.NewVoucherCodec("transfer",[]string{"address","uint256"}) // receiver, amount
  erc721VoucherCodec = abihandler.NewVoucherCodec("safeTransferFrom",[]string{"address","address","uint256"}) // sender, receive, tokenId

  app := &WalletApp{Wallets: make(map[abihandler.Address]*Wallet)}

  return app
}

func (this *WalletApp) SetHandler(handler *hdl.Handler) {
  if handler == nil {
    panic("Nil handler")
  }
  this.handler = handler
}

func (this *WalletApp) SetAbiHandler(abiHdl *abihandler.AbiHandler) {
  if abiHdl == nil {
    panic("Nil handler")
  }
  if this.handler == nil {
    this.SetHandler(abiHdl.Handler)
  }
  this.abiHandler = abiHdl
}

func (this *WalletApp) SetUriHandler(uriHdl *urihandler.UriHandler) {
  if uriHdl == nil {
    panic("Nil handler")
  }
  if this.handler == nil {
    this.SetHandler(uriHdl.Handler)
  }
  this.uriHandler = uriHdl
}

func (this *WalletApp) Handler() *hdl.Handler {
  if this.handler == nil {
    this.SetHandler(hdl.NewSimpleHandler())
  }
  return this.handler
}

func (this *WalletApp) AbiHandler() *abihandler.AbiHandler {
  if this.abiHandler == nil {
    this.SetAbiHandler(abihandler.AddAbiHandler(this.Handler()))
  }
  return this.abiHandler
}

func (this *WalletApp) UriHandler() *urihandler.UriHandler {
  if this.uriHandler == nil {
    this.SetUriHandler(urihandler.AddUriHandler(this.Handler()))
  }
  return this.uriHandler
}

func (this *WalletApp) SetupRoutes(routes []WalletRoute) {

  var forceRelayRoute bool
  var relayRouteAdded bool

  for _, route := range routes {
    switch route {
    case DappRelayAdvanceRoute:
      relayRouteAdded = true
      this.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.DappAddressRelay, abihandler.NewPackedCodec([]string{"address"}), this.HandleRelay)
    case DepositEtherAdvanceRoute:
      this.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.EtherPortalAddress, abihandler.NewPackedCodec([]string{"address","uint256","bytes"}), this.EtherPortalDeposit)
    case DepositErc20AdvanceRoute:
      this.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc20PortalAddress, abihandler.NewPackedCodec([]string{"bool","address","address","uint256","bytes"}), this.Erc20PortalDeposit)
    case DepositErc721AdvanceRoute:
      this.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.Erc721PortalAddress, abihandler.NewPackedCodec([]string{"address","address","uint256","bytes"}), this.Erc721PortalDeposit)
    case WithdrawEtherAdvanceRoute,WithdrawEtherCodecAdvanceRoute:
      forceRelayRoute = true
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherWithdraw",[]string{"uint256","bytes"}), this.EtherWithdraw)
    case WithdrawErc20AdvanceRoute,WithdrawErc20CodecAdvanceRoute:
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","ERC20_Withdraw",[]string{"address","uint256","bytes"}), this.Erc20Withdraw)
    case WithdrawErc721AdvanceRoute,WithdrawErc721CodecAdvanceRoute:
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","ERC721_Withdraw",[]string{"address","uint256","bytes"}), this.Erc721Withdraw)
      forceRelayRoute = true
    case TransferEtherAdvanceRoute,TransferEtherCodecAdvanceRoute:
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","EtherTransfer",[]string{"address","uint256","bytes"}), this.TransferEtherCodec)
    case TransferErc20AdvanceRoute,TransferErc20CodecAdvanceRoute:
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc20Transfer",[]string{"address","address","uint256","bytes"}), this.TransferErc20Codec)
    case TransferErc721AdvanceRoute,TransferErc721CodecAdvanceRoute:
      this.AbiHandler().HandleAdvanceRoute(abihandler.NewHeaderCodec("wallet","Erc721Transfer",[]string{"address","address","uint256","bytes"}), this.TransferErc721Codec)
    case BalanceInspectRoute,BalanceCodecInspectRoute:
      this.AbiHandler().HandleInspectRoute(abihandler.NewHeaderCodec("wallet","Balance",[]string{"address address"}), this.BalanceAbi)
    case BalanceUriInspectRoute:
      this.UriHandler().HandleInspectRoute("/balance/:address", this.BalanceUri)
    default:
      panic("Unrecognized route")
    }
  }
  if forceRelayRoute && !relayRouteAdded {
    this.AbiHandler().HandleFixedAddressAdvance(hdl.RollupsAddresses.DappAddressRelay, abihandler.NewPackedCodec([]string{"address"}), this.HandleRelay)
  }
}

//
// Relay
//

func (this *WalletApp) HandleRelay(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  addr, ok1 := payloadMap["0"].(abihandler.Address)

  if !ok1 {
    message := "HandleRelay: parameters error"
    return fmt.Errorf(message)
  }

  this.DappAddress = addr

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Dapp Relay, dapp address is", addr)}

  return nil
}

//
// Deposit
//

func (this *WalletApp) EtherPortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  depositor, ok1 := payloadMap["0"].(abihandler.Address)
  amount, ok2 := payloadMap["1"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)

  if !ok1 || !ok2 {
    message := "EtherPortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := this.GetWallet(depositor)

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
  notice := rollups.Notice{noticePayload}
  _, err = rollups.SendNotice(&notice)
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amount,"native token deposit from",depositor)}
  
  return nil
}

func (this *WalletApp) Erc20PortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {

  // ret, ok1 := payloadMap["0"].(bool)
  tokenAddress, ok2 := payloadMap["1"].(abihandler.Address)
  depositor, ok3 := payloadMap["2"].(abihandler.Address)
  amount, ok4 := payloadMap["3"].(*big.Int)
  // dataBytes, ok5 := payloadMap["4"].([]byte)

  if !ok2 || !ok3 || !ok4 {
    message := "Erc20PortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := this.GetWallet(depositor)

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
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received",amount,"tokens",tokenAddress,"Erc20 deposit from",depositor)}

  return nil
}

func (this *WalletApp) Erc721PortalDeposit(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  depositor, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok4 := payloadMap["3"].([]byte)

  if !ok1 || !ok2 || !ok3 {
    message := "Erc721PortalDeposit: parameters error"
    return fmt.Errorf(message)
  }

  wallet := this.GetWallet(depositor)

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
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc721PortalDeposit: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Received id",tokenId,tokenAddress,"Erc721 deposit from",depositor)}

  return nil
}

//
// Withdraw
//

func (this *WalletApp) EtherWithdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if this.DappAddress == (abihandler.Address{}) {
    return fmt.Errorf("EtherWithdraw: Can not generate voucher as there is no dapp address configured")
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("EtherWithdraw: payload:",payloadMap)}
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

  wallet := this.GetWallet(addr)

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
  _, err = this.handler.SendVoucher(this.DappAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("EtherPortalDeposit: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := etherNoticeCodec.Encode([]interface{}{addr,amount,wallet.Ether})
  if err != nil {
    return fmt.Errorf("EtherWithdraw: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("EtherWithdraw: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amount,"ETH from",addr,"data:",dataBytes)}

  return nil
}

func (this *WalletApp) Erc20Withdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("EtherWithdraw: payload:",payloadMap)}
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

  wallet := this.GetWallet(addr)

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
  _, err = this.handler.SendVoucher(this.DappAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := erc20NoticeCodec.Encode([]interface{}{addr,tokenAddress,amount,wallet.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc20Withdraw: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn",amount,"of Erc20",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

func (this *WalletApp) Erc721Withdraw(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Erc721Withdraw: payload:",payloadMap)}

  if this.DappAddress == (abihandler.Address{}) {
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

  wallet := this.GetWallet(addr)

  // Withdrawal
  err = wallet.WithdrawErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error withdrawing Erc721: %s", err)
  }
    
  // Voucher
  voucherPayload,err := erc721VoucherCodec.Encode([]interface{}{this.DappAddress,addr,tokenId})
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: encoding voucher: %s", err)
  }
  _, err = this.handler.SendVoucher(this.DappAddress.String(),voucherPayload)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error making http request: %s", err)
  }

  // Notice
  noticePayload,err := erc721NoticeCodec.Encode([]interface{}{addr,tokenAddress,tokenId,wallet.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("Erc721Withdraw: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Withdrawn id",tokenId,"of Erc721",tokenAddress,"from",addr,"data:",dataBytes)}

  return nil
}

//
// Transfer
//

func (this *WalletApp) TransferEtherCodec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
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

  return this.TransferEther(sender,receiver,amount)
}

func (this *WalletApp) TransferEther(sender abihandler.Address, receiver abihandler.Address, amount *big.Int) error {

  walletSender := this.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawEther(amount)
  if err != nil {
    return fmt.Errorf("TransferEther: error withdrawing Ether: %s", err)
  }

  walletReceiver := this.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositEther(amount)
  if err != nil {
    return fmt.Errorf("TransferEther: error depositing Ether: %s", err)
  }

  // Notice
  noticePayload,err := etherNoticeCodec.Encode([]interface{}{sender,amount,walletSender.Ether})
  if err != nil {
    return fmt.Errorf("TransferEther: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferEther: error making http request: %s", err)
  }

  noticePayload,err = etherNoticeCodec.Encode([]interface{}{receiver,amount,walletReceiver.Ether})
  if err != nil {
    return fmt.Errorf("TransferEther: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferEther: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",amount,"ETH from",sender,"to",receiver)}

  return nil
}

func (this *WalletApp) TransferErc20Codec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  amount, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 {
    message := "TransferErec20Codec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErec20Codec: error converting address: %s", err)
  }

  return this.TransferErc20(tokenAddress,sender,receiver,amount)
}

func (this *WalletApp) TransferErc20(tokenAddress abihandler.Address, sender abihandler.Address, receiver abihandler.Address, amount *big.Int) error {

  walletSender := this.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawErc20(tokenAddress,amount)
  if err != nil {
    return fmt.Errorf("TransferErc20: error withdrawing Erc20: %s", err)
  }

  walletReceiver := this.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositErc20(tokenAddress,amount)
  if err != nil {
    return fmt.Errorf("TransferErc20: error depositing Erc20: %s", err)
  }

  // Notice
  noticePayload,err := erc20NoticeCodec.Encode([]interface{}{tokenAddress,sender,amount,walletSender.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("TransferErc20: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc20: error making http request: %s", err)
  }

  noticePayload,err = erc20NoticeCodec.Encode([]interface{}{tokenAddress,receiver,amount,walletReceiver.Erc20[tokenAddress]})
  if err != nil {
    return fmt.Errorf("TransferErc20: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc20: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",amount,"Erc20",tokenAddress,"from",sender,"to",receiver)}

  return nil
}

func (this *WalletApp) TransferErc721Codec(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  tokenAddress, ok1 := payloadMap["0"].(abihandler.Address)
  receiver, ok2 := payloadMap["1"].(abihandler.Address)
  tokenId, ok3 := payloadMap["2"].(*big.Int)
  // dataBytes, ok3 := payloadMap["2"].([]byte)
  if !ok1 || !ok2 || !ok3 {
    message := "TransferErec721Codec: parameters error"
    return fmt.Errorf(message)
  }

  sender,err := abihandler.Hex2Address(metadata.MsgSender)
  if err != nil {
    return fmt.Errorf("TransferErec721Codec: error converting address: %s", err)
  }

  return this.TransferErc721(tokenAddress,sender,receiver,tokenId)
}

func (this *WalletApp) TransferErc721(tokenAddress abihandler.Address, sender abihandler.Address, receiver abihandler.Address, tokenId *big.Int) error {

  walletSender := this.GetWallet(sender)

  // Withdrawal
  err := walletSender.WithdrawErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("TransferErc721: error withdrawing Erc721: %s", err)
  }

  walletReceiver := this.GetWallet(receiver)

  // Deposit
  err = walletReceiver.DepositErc721(tokenAddress,tokenId)
  if err != nil {
    return fmt.Errorf("TransferErc721: error depositing Erc721: %s", err)
  }

  // Notice
  noticePayload,err := erc721NoticeCodec.Encode([]interface{}{tokenAddress,sender,tokenId,walletSender.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("TransferErc721: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc721: error making http request: %s", err)
  }

  noticePayload,err = erc721NoticeCodec.Encode([]interface{}{tokenAddress,receiver,tokenId,walletReceiver.Erc721TokenIdList(tokenAddress)})
  if err != nil {
    return fmt.Errorf("TransferErc721: encoding notice: %s", err)
  }
  _, err = this.handler.SendNotice(noticePayload)
  if err != nil {
    return fmt.Errorf("TransferErc721: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println("Transfered",tokenId,"Erc721",tokenAddress,"from",sender,"to",receiver)}

  return nil
}


//
// Balance
//

func (this *WalletApp) BalanceAbi(payloadMap map[string]interface{}) error {
  addr, ok1 := payloadMap["address"].(abihandler.Address)
  if !ok1 {
    message := "Balance: parameters error"
    return fmt.Errorf(message)
  }

  wallet := this.GetWallet(addr)

  balanceJson, err := json.Marshal(wallet)
  if err != nil {
    return fmt.Errorf("Balance: error converting wallet to json: %s", err)
  }

  err = this.handler.SendReport(rollups.Str2Hex(string(balanceJson)))
  if err != nil {
    return fmt.Errorf("Balance: error making http request: %s", err)
  }

  if this.handler.LogLevel >= hdl.Debug {DebugLogger.Println(addr,"balance",string(balanceJson))}

  return nil
}

func (this *WalletApp) BalanceUri(payloadMap map[string]interface{}) error {
  addrStr, ok1 := payloadMap["address"].(string)
  if !ok1 {
    return fmt.Errorf("Balance: parameters error")
  }

  addr,err := abihandler.Hex2Address(addrStr)
  if err != nil {
    return fmt.Errorf("Balance: parameters error: %s", err)
  }

  return this.BalanceAbi(map[string]interface{}{"address":addr})
}
