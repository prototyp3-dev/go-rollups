package rollups

import (
	"encoding/hex"
  "math/big"
)

type EtherDeposit struct {
  Depositor string
  Amount *big.Int
  Data []byte
}

type Erc20Deposit struct {
  Depositor string
  TokenAddress string
  Amount *big.Int
  Data []byte
}

type Erc721Deposit struct {
  Depositor string
  TokenAddress string
  TokenId *big.Int
  Data []byte
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
