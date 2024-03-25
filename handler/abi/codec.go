package abihandler

import (
  "strings"
  "fmt"

  "github.com/prototyp3-dev/go-rollups/rollups"

  "github.com/lynoferraz/abigo"
  "github.com/umbracle/ethgo" //"github.com/ethereum/go-ethereum/crypto"
)

type Address = ethgo.Address //[20]byte

func Address2Hex(bin Address) string {
  return rollups.Bin2Hex(bin[:])
}

func Hex2Address(hexAddress string) (Address,error) {
  addrBytes, err := rollups.Hex2Bin(hexAddress)
  if err != nil {
    return Address{}, fmt.Errorf("Hex2Address: Error converting address from hex")
  }
  return Bin2Address(addrBytes)
}

func Bin2Address(addrBytes []byte) (Address,error) {
  if len(addrBytes) != 20 {
    return Address{}, fmt.Errorf("Bin2Address: Wrong address length")
  }
  var addrArr [20]byte
  copy(addrArr[:], addrBytes[:20])
  return addrArr,nil
}

type Codec struct {
  Framework string
  Method string
  Header string
  PackedFields []string
  Fields []string
  typ *abi.Type
}

func (c Codec) String() string {
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
  typ := GetType(fields)
  return &Codec{PackedFields: fields, typ: typ}
}

func NewHeaderPackedCodec(framework string, method string, fields []string) *Codec {
  codec := NewPackedCodec(fields)
  cleanFields := CleanFields(codec.typ)
  header := CodecHeader(framework, method, cleanFields)
  codec.Header = header
  codec.Framework = framework
  codec.Method = method
  return codec
}

func NewCodec(fields []string) *Codec {
  typ := GetType(fields)
  return &Codec{Fields: fields, typ: typ}
}

func NewHeaderCodec(framework string, method string, fields []string) *Codec {
  codec := NewCodec(fields)
  cleanFields := CleanFields(codec.typ)
  header := CodecHeader(framework, method, cleanFields)
  codec.Header = header
  codec.Framework = framework
  codec.Method = method
  return codec
}

func GetType(fields []string) *abi.Type {
  return abi.MustNewType("tuple("+ strings.Join(fields, ",") +")")
}

func CleanFields(typ *abi.Type) []string {
  var cleanFields []string
  for _, elem := range typ.TupleElems() {
    cleanFields = append(cleanFields, elem.Elem.String())
  }
  return cleanFields
}

func CodecHeader(framework string, method string, fields []string) string {
  frameworkeccak := ethgo.Keccak256([]byte(framework))
  methodkeccak := ethgo.Keccak256([]byte(method))
  fieldskeccak := ethgo.Keccak256([]byte("("+strings.Join(fields, ",")+")"))
  headerAllbytes := append(frameworkeccak, methodkeccak...)
  headerAllbytes = append(headerAllbytes, fieldskeccak...)
  return rollups.Bin2Hex(ethgo.Keccak256(headerAllbytes))
}

func NewVoucherCodec(method string, fields []string) *Codec {
  codec := NewCodec(fields)
  header := VoucherHeader(method, codec.Fields)
  codec.Header = header
  codec.Method = method
  return codec
}

func VoucherHeader(method string, fields []string) string {
  headerKeccak := ethgo.Keccak256([]byte(method+"("+strings.Join(fields, ",")+")"))
  return rollups.Bin2Hex(headerKeccak[:4])
}
func (c *Codec) Decode(payloadHex string) (map[string]interface{},error) {
	var result map[string]interface{}
  payloadBytes, err := rollups.Hex2Bin(payloadHex)
  if err != nil {
    return result,fmt.Errorf("Decode: %s", err)
  }
  if len(c.Header) > 0 {
    if payloadHex[:66] != c.Header {
      return result,fmt.Errorf("Decode: Header does not match")
    }
    payloadBytes = payloadBytes[32:]
  }

  var fields []string
  if c.PackedFields != nil {
    fields = c.PackedFields
  } else if c.Fields != nil {
    fields = c.Fields
  }

  if len(fields) == 0 {
    return result,nil
  }
  // tupleFields := make([]string,0)
  // for i := 0; i < len(fields); i += 1 {
  //   tupleFields = append(tupleFields,fmt.Sprintf("%s f%d",fields[i],i))
  // }
  // typ, _ := abi.NewType("tuple("+ strings.Join(tupleFields, ",") +")")

  var decoded interface{}
  if c.PackedFields != nil {
    decoded,err = abi.DecodePacked(c.typ, payloadBytes)
  } else if c.Fields != nil {
    decoded,err = abi.Decode(c.typ, payloadBytes)
  }
  if err != nil {
    return result,fmt.Errorf("Decode: %s",err)
  }

  mapResult, ok := decoded.(map[string]interface{})
  if !ok {
    return result,fmt.Errorf("convert decoded payload to map error")
  }
  // result = make([]interface{},len(mapResult))
  // for i := 0; i < len(mapResult); i += 1 {
  //   key := fmt.Sprintf("f%d",i)
  //   result[i] = mapResult[key]
  // }

  return mapResult,nil
}

func (c *Codec) Encode(payload interface{}) (string,error) {
	var result string

  var fields []string
  if c.PackedFields != nil {
    fields = c.PackedFields
  } else if c.Fields != nil {
    fields = c.Fields
  }

  var payloadMap map[string]interface{}
  typ := c.typ

  payloadSlice, ok := payload.([]interface{})
  if ok {
    if len(fields) != len(payloadSlice) {
      return result,fmt.Errorf("Encode: Wrong values length")
    }
    
    tupleFields := make([]string,0)
    payloadMap = make(map[string]interface{})
    for i := 0; i < len(fields); i += 1 {
      key := fmt.Sprintf("f%d",i)
      tupleFields = append(tupleFields,fmt.Sprintf("%s %s",fields[i],key))
      payloadMap[key] = payloadSlice[i]
    }
    typ, _ = abi.NewType("tuple("+ strings.Join(tupleFields, ",") +")")
  
  } else {
    payloadMap, ok = payload.(map[string]interface{})
    if !ok {
      return result,fmt.Errorf("Encode: Wrong payload")
    }
  }

  var encoded []byte
  var err error
  if c.PackedFields != nil {
    encoded,err = abi.EncodePacked(payloadMap,typ)
  } else if c.Fields != nil {
    encoded,err = abi.Encode(payloadMap,typ)
  }
  if err != nil {
    return result,fmt.Errorf("Encode: %s",err)
  }

  result = rollups.Bin2Hex(encoded)
  if c.Header != "" {
    result = c.Header + result[2:]
  }
  return result,nil
}
