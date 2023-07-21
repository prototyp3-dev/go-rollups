package abihandler

import (
  "strings"
  "fmt"

  "github.com/prototyp3-dev/go-rollups/rollups"

  "github.com/prototyp3-dev/go-rollups/abi"
  crypto "github.com/umbracle/ethgo" //"github.com/ethereum/go-ethereum/crypto"
)

// // go-ethereum 
// https://stackoverflow.com/questions/50772811/how-can-i-get-the-same-return-value-as-solidity-abi-encodepacked-in-golang
// https://github.com/ethereum/go-ethereum/blob/master/accounts/abi/pack_test.go

// // eth go
// https://ethereum.stackexchange.com/questions/117060/abi-decode-raw-types-with-go
// https://github.com/umbracle/ethgo/blob/main/abi/abi_test.go
// https://github.com/umbracle/ethgo/blob/main/keccak.go

type Codec struct {
  Framework string
  Method string
  Header string
  PackedFields []string
  Fields []string
}

func (c *Codec) String() string {
  atts := make([]string,0)
  if c.Header != "" && c.Framework != "" &&c.Method != "" {
    atts = append(atts, fmt.Sprintf("Header(%s)",c.Header))
    atts = append(atts, fmt.Sprintf("Framework(%s)",c.Framework))
    atts = append(atts, fmt.Sprintf("Method(%s)",c.Method))
  }
  if len(c.Fields) > 0 {
    atts = append(atts, fmt.Sprintf("Fields(%s)",c.Fields))
  }
  if len(c.PackedFields) > 0 {
    atts = append(atts, fmt.Sprintf("PackedFields(%s)",c.PackedFields))
  }
  return fmt.Sprintf("Codec{%s}",strings.Join(atts,","))
}

func NewPackedCodec(fields []string) *Codec {
  typ := abi.MustNewType("tuple("+ strings.Join(fields, ",") +")")
  var cleanFields []string
  for _, elem := range typ.TupleElems() {
    cleanFields = append(cleanFields, elem.Elem.String())
  }
  return &Codec{PackedFields: cleanFields}
}

func NewHeaderPackedCodec(framework string, method string, fields []string) *Codec {
  codec := NewPackedCodec(fields)
  header := CodecHeader(framework, method, codec.PackedFields)
  codec.Header = header
  codec.Framework = framework
  codec.Method = method
  return codec
}

func NewCodec(fields []string) *Codec {
  typ := abi.MustNewType("tuple("+ strings.Join(fields, ",") +")")
  var cleanFields []string
  for _, elem := range typ.TupleElems() {
    cleanFields = append(cleanFields, elem.Elem.String())
  }
  return &Codec{Fields: cleanFields}
}

func NewHeaderCodec(framework string, method string, fields []string) *Codec {
  codec := NewCodec(fields)
  header := CodecHeader(framework, method, codec.Fields)
  codec.Header = header
  codec.Framework = framework
  codec.Method = method
  return codec
}

func CodecHeader(framework string, method string, fields []string) string {
  frameworkeccak := crypto.Keccak256([]byte(framework))
  methodkeccak := crypto.Keccak256([]byte(method))
  fieldskeccak := crypto.Keccak256([]byte("("+strings.Join(fields, ",")+")"))
  headerAllbytes := append(frameworkeccak, methodkeccak...)
  headerAllbytes = append(headerAllbytes, fieldskeccak...)
  return rollups.Bin2Hex(crypto.Keccak256(headerAllbytes))
}

func NewVoucherCodec(method string, fields []string) *Codec {
  codec := NewCodec(fields)
  header := VoucherHeader(method, codec.Fields)
  codec.Header = header
  codec.Method = method
  return codec
}

func VoucherHeader(method string, fields []string) string {
  headerKeccak := crypto.Keccak256([]byte(method+"("+strings.Join(fields, ",")+")"))
  return rollups.Bin2Hex(crypto.Keccak256(headerKeccak[:4]))
}
func (this *Codec) Decode(payloadHex string) ([]interface{},error) {
	var result []interface{}
  payloadBytes, err := rollups.Hex2Bin(payloadHex)
  if err != nil {
    return result,fmt.Errorf("Decode: %s", err)
  }
  if len(this.Header) > 0 {
    if payloadHex[:66] != this.Header {
      return result,fmt.Errorf("Decode: Header does not match")
    }
    payloadBytes = payloadBytes[32:]
  }

  var fields []string
  if this.PackedFields != nil {
    fields = this.PackedFields
  } else if this.Fields != nil {
    fields = this.Fields
  }

  if len(fields) == 0 {
    return result,nil
  }
  tupleFields := make([]string,0)
  for i := 0; i < len(fields); i += 1 {
    tupleFields = append(tupleFields,fmt.Sprintf("%s f%d",fields[i],i))
  }
  typ, _ := abi.NewType("tuple("+ strings.Join(tupleFields, ",") +")")

  var decoded interface{}
  if this.PackedFields != nil {
    decoded,err = abi.DecodePacked(typ, payloadBytes)
  } else if this.Fields != nil {
    decoded,err = abi.Decode(typ, payloadBytes)
  }
  if err != nil {
    return result,fmt.Errorf("Decode: %s",err)
  }

  mapResult, ok := decoded.(map[string]interface{})
  if !ok {
    return result,fmt.Errorf("Convert decoded payload to map error")
  }
  result = make([]interface{},len(mapResult))
  for i := 0; i < len(mapResult); i += 1 {
    key := fmt.Sprintf("f%d",i)
    result[i] = mapResult[key]
  }
  return result,nil
}

func (this *Codec) Encode(payloadSlice []interface{}) (string,error) {
	var result string

  var fields []string
  if this.PackedFields != nil {
    fields = this.PackedFields
  } else if this.Fields != nil {
    fields = this.Fields
  }

  if len(fields) != len(payloadSlice) {
    return result,fmt.Errorf("Encode: Wrong values length")
  }

  tupleFields := make([]string,0)
  payloadMap := make(map[string]interface{})
  for i := 0; i < len(fields); i += 1 {
    key := fmt.Sprintf("f%d",i)
    tupleFields = append(tupleFields,fmt.Sprintf("%s %s",fields[i],key))
    payloadMap[key] = payloadSlice[i]
  }
  typ, _ := abi.NewType("tuple("+ strings.Join(tupleFields, ",") +")")

  var encoded []byte
  var err error
  if this.PackedFields != nil {
    encoded,err = abi.EncodePacked(payloadMap,typ)
  } else if this.Fields != nil {
    encoded,err = abi.Encode(payloadMap,typ)
  }
  if err != nil {
    return result,fmt.Errorf("Encode: %s",err)
  }

  result = rollups.Bin2Hex(encoded)
  if this.Header != "" {
    result = this.Header + result[2:]
  }
  return result,nil
}
