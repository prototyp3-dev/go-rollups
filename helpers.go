package rollups

import (
	"encoding/json"
	"encoding/hex"
	"net/http"
	"bytes"
  "os"
  "math/big"
)

var rollup_server = os.Getenv("ROLLUP_HTTP_SERVER_URL")

func SendPost(endpoint string, jsonData []byte) (*http.Response, error) {
  req, err := http.NewRequest(http.MethodPost, rollup_server + "/" + endpoint, bytes.NewBuffer(jsonData))
  if err != nil {
    return &http.Response{}, err
  }
  req.Header.Set("Content-Type", "application/json; charset=UTF-8")

  return http.DefaultClient.Do(req)
}

func SendFinish(finish *Finish) (*http.Response, error) {
  body, err := json.Marshal(finish)
  if err != nil {
    return &http.Response{}, err
  }
  
  return SendPost("finish", body)
}

func SendReport(report *Report) (*http.Response, error) {
  body, err := json.Marshal(report)
  if err != nil {
    return &http.Response{}, err
  }
  
  return SendPost("report", body)
}

func SendNotice(notice *Notice) (*http.Response, error) {
  body, err := json.Marshal(notice)
  if err != nil {
    return &http.Response{}, err
  }
  
  return SendPost("notice", body)
}

func SendVoucher(voucher *Voucher) (*http.Response, error) {
  body, err := json.Marshal(voucher)
  if err != nil {
    return &http.Response{}, err
  }
  
  return SendPost("voucher", body)
}

func SendException(exception *Exception) (*http.Response, error) {
  body, err := json.Marshal(exception)
  if err != nil {
    return &http.Response{}, err
  }
  
  return SendPost("exception", body)
}

func Hex2Str(hx string) (string, error) {
  bin, err := Hex2Bin(hx)
	if err != nil {
    return string(bin), err
	}
  return string(bin), nil
}

func Hex2Bin(hx string) ([]byte, error) {
  bin, err := hex.DecodeString(hx[2:])
	if err != nil {
    return bin, err
	}
  return bin, nil
}

func Str2Hex(str string) string {
  return Bin2Hex([]byte(str))
}

func Bin2Hex(bin []byte) string {
  hx := hex.EncodeToString(bin)
  return "0x"+string(hx)
}

func DecodeEtherDeposit(payloadHex string) (EtherDeposit,error) {
  bin, err := Hex2Bin(payloadHex)
	if err != nil {
    return EtherDeposit{}, err
	}
  
  amount := new(big.Int)
  amount.SetBytes(bin[20:52])

  return EtherDeposit{Depositor:Bin2Hex(bin[:20]), Amount:amount, Data:bin[52:]},nil
}

func DecodeErc20Deposit(payloadHex string) (Erc20Deposit,error) {
  bin, err := Hex2Bin(payloadHex)
	if err != nil {
    return Erc20Deposit{}, err
	}

  amount := new(big.Int)
  amount.SetBytes(bin[41:73])

  return Erc20Deposit{Depositor:Bin2Hex(bin[21:41]), TokenAddress:Bin2Hex(bin[1:21]), Amount:amount, Data:bin[73:]},nil
}

func DecodeErc721Deposit(payloadHex string) (Erc721Deposit,error) {
  bin, err := Hex2Bin(payloadHex)
	if err != nil {
    return Erc721Deposit{}, err
	}

  tokenId := new(big.Int)
  tokenId.SetBytes(bin[40:72])
  
  return Erc721Deposit{Depositor:Bin2Hex(bin[20:40]), TokenAddress:Bin2Hex(bin[:20]), TokenId:tokenId, Data:bin[72:]}, nil
}

func EtherWithdralVoucher(Sender string, Receiver string, Amount *big.Int) Voucher {
  payload := "0x522f6815" + hex.EncodeToString(append(Address2Bin(Receiver),PadBytes(Amount.Bytes(),32)...))
  return Voucher{Destination: Sender, Payload: payload}
}

func Erc20TransferVoucher(Receiver string, TokenAddress string, Amount *big.Int) Voucher {
  payload := "0xa9059cbb" + hex.EncodeToString(append(Address2Bin(Receiver),PadBytes(Amount.Bytes(),32)...))
  return Voucher{Destination: TokenAddress, Payload: payload}
}

func Erc721SafeTransferVoucher(Sender string, Receiver string, TokenAddress string, TokenId *big.Int) Voucher {
  payloadBytes := append(Address2Bin(Sender),Address2Bin(Receiver)...)
  payloadBytes = append(payloadBytes,PadBytes(TokenId.Bytes(),32)...)
  payload := "0x42842e0e" + hex.EncodeToString(payloadBytes)
  return Voucher{Destination: TokenAddress, Payload: payload}
}

func Address2Bin(address string) []byte {
  addressBin,_ := Hex2Bin(address)
  tmp := make([]byte, 32)
  copy(tmp[12:], addressBin)
  return tmp
}

func PadBytes(bin []byte,size int) []byte {
  tmp := make([]byte, size)
  l := len(bin)
  copy(tmp[(size - l):], bin)
  return tmp
}
