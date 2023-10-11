package rollups

import (
	"encoding/hex"
)

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

func PadBytesLeft(bin []byte,size int) []byte {
  return PadBytes(bin,size)
}

func PadBytesRight(bin []byte,size int) []byte {
  tmp := make([]byte, size)
  copy(tmp[:size], bin)
  return tmp
}