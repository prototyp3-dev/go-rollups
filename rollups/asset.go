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

type Erc1155SingleDeposit struct {
  Depositor string
  TokenAddress string
  TokenId *big.Int
  Amount *big.Int
  BaseLayerData []byte
  ExecLayerData []byte
}

type Erc1155BatchDeposit struct {
  Depositor string
  TokenAddress string
  TokenIds []*big.Int
  Amounts []*big.Int
  BaseLayerData []byte
  ExecLayerData []byte
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
  amount.SetBytes(bin[40:72])

  return Erc20Deposit{Depositor:Bin2Hex(bin[20:40]), TokenAddress:Bin2Hex(bin[:20]), Amount:amount, Data:bin[72:]},nil
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

func DecodeErc1155SingleDeposit(payloadHex string) (Erc1155SingleDeposit,error) {
  bin, err := Hex2Bin(payloadHex)
	if err != nil {
    return Erc1155SingleDeposit{}, err
	}

  tokenId := new(big.Int)
  tokenId.SetBytes(bin[40:72])
  
  amount := new(big.Int)
  amount.SetBytes(bin[72:104])

  allData := bin[104:]

  blDataPosition := int(new(big.Int).SetBytes(allData[0:32]).Int64())
  blDataSize := int(new(big.Int).SetBytes(allData[blDataPosition:blDataPosition+32]).Int64())
  blData := allData[blDataPosition+32:blDataPosition+32+blDataSize]

  elDataPosition := int(new(big.Int).SetBytes(allData[32:64]).Int64())
  elDataSize := int(new(big.Int).SetBytes(allData[elDataPosition:elDataPosition+32]).Int64())
  elData := allData[elDataPosition+32:elDataPosition+32+elDataSize]

  return Erc1155SingleDeposit{Depositor:Bin2Hex(bin[20:40]), TokenAddress:Bin2Hex(bin[:20]), TokenId:tokenId, Amount:amount, BaseLayerData:blData, ExecLayerData:elData}, nil
}

func DecodeErc1155BatchDeposit(payloadHex string) (Erc1155BatchDeposit,error) {
  bin, err := Hex2Bin(payloadHex)
	if err != nil {
    return Erc1155BatchDeposit{}, err
	}

  token := Bin2Hex(bin[:20])
  depositor := Bin2Hex(bin[20:40])

  allData := bin[40:]

  idsPosition := int(new(big.Int).SetBytes(allData[0:32]).Int64())
  idsSize := int(new(big.Int).SetBytes(allData[idsPosition:idsPosition+32]).Int64())
  idList := make([]*big.Int,0)
  for i := 0; i < idsSize ; i++ {
    idList = append(idList,new(big.Int).SetBytes( allData[idsPosition+32*(i+1):idsPosition+32*(i+2)] ))
  }

  amountsPosition := int(new(big.Int).SetBytes(allData[32:64]).Int64())
  amountsSize := int(new(big.Int).SetBytes(allData[amountsPosition:amountsPosition+32]).Int64())
  amountList := make([]*big.Int,0)
  for i := 0; i < amountsSize ; i++ {
    amountList = append(amountList,new(big.Int).SetBytes( allData[amountsPosition+32*(i+1):amountsPosition+32*(i+2)] ))
  }

  blDataPosition := int(new(big.Int).SetBytes(allData[64:96]).Int64())
  blDataSize := int(new(big.Int).SetBytes(allData[blDataPosition:blDataPosition+32]).Int64())
  blData := allData[blDataPosition+32:blDataPosition+32+blDataSize]

  elDataPosition := int(new(big.Int).SetBytes(allData[96:128]).Int64())
  elDataSize := int(new(big.Int).SetBytes(allData[elDataPosition:elDataPosition+32]).Int64())
  elData := allData[elDataPosition+32:elDataPosition+32+elDataSize]

  return Erc1155BatchDeposit{Depositor:depositor, TokenAddress:token, TokenIds:idList, Amounts:amountList, BaseLayerData:blData, ExecLayerData:elData}, nil
}

func EtherWithdralVoucher(Receiver string, Amount *big.Int) Voucher {
  return Voucher{Destination: Receiver, Payload: "0x", Value: Amount}
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

// safeTransferFrom(address from, address to, uint256 id, uint256 amount, bytes data)
func Erc1155SafeTransferFromVoucher(Sender string, Receiver string, TokenAddress string, TokenId *big.Int, Amount *big.Int, Data []byte) Voucher {
  payloadBytes := append(Address2Bin(Sender),Address2Bin(Receiver)...)
  payloadBytes = append(payloadBytes,PadBytes(TokenId.Bytes(),32)...)
  payloadBytes = append(payloadBytes,PadBytes(Amount.Bytes(),32)...)
  dataPosition := 160 // 32 * 5
  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(dataPosition)).Bytes(),32)...)
  lenData := len(Data)
  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(lenData)).Bytes(),32)...)
  if lenData > 0 {
    slots := lenData/32
    if lenData%32 > 0 {
      slots = slots + 1
    }
    payloadBytes = append(payloadBytes,PadBytesRight(Data,slots)...)
  }

  payload := "0xf242432a" + hex.EncodeToString(payloadBytes)
  return Voucher{Destination: TokenAddress, Payload: payload}
}

// safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] amounts, bytes data)
func Erc1155SafeBatchTransferFromVoucher(Sender string, Receiver string, TokenAddress string, TokenIds []*big.Int, Amounts []*big.Int, Data []byte) Voucher {
  payloadBytes := append(Address2Bin(Sender),Address2Bin(Receiver)...)

  lenIds := len(TokenIds)
  lenAmounts := len(Amounts)
  lenData := len(Data)

  idsPosition := 160 // 32 * 5
  amountsPosition := idsPosition + 32 + lenIds * 32
  dataPosition := amountsPosition + 32 + lenAmounts * 32

  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(idsPosition)).Bytes(),32)...)
  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(amountsPosition)).Bytes(),32)...)
  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(dataPosition)).Bytes(),32)...)

  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(lenIds)).Bytes(),32)...)
  for _, id := range TokenIds {
    payloadBytes = append(payloadBytes,PadBytes(id.Bytes(),32)...)
  }

  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(lenAmounts)).Bytes(),32)...)
  for _,amount := range Amounts {
    payloadBytes = append(payloadBytes,PadBytes(amount.Bytes(),32)...)
  }

  payloadBytes = append(payloadBytes,PadBytes(new(big.Int).SetInt64(int64(lenData)).Bytes(),32)...)
  if lenData > 0 {
    slots := lenData/32
    if lenData%32 > 0 {
      slots = slots + 1
    }
    payloadBytes = append(payloadBytes,PadBytesRight(Data,slots)...)
  }
  
  payload := "0x2eb2c2d6" + hex.EncodeToString(payloadBytes)
  return Voucher{Destination: TokenAddress, Payload: payload}
}
