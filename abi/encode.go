// based on the https://github.com/umbracle/ethgo/blob/main/abi/encode.go
package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

var (
	zero = big.NewInt(0)
	one  = big.NewInt(1)
)

func Encode(v interface{}, t *abi.Type) ([]byte, error) {
	return abi.Encode(v, t)
}

// Encode encodes a value
func EncodePacked(v interface{}, t *abi.Type) ([]byte, error) {
	return encode(reflect.ValueOf(v), t)
}

func encode(v reflect.Value, t *abi.Type) ([]byte, error) {
	fmt.Println(v,t)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch t.Kind() {
	case abi.KindSlice, abi.KindArray:
		return encodeSliceAndArray(v, t)

	case abi.KindTuple:
		return encodeTuple(v, t)

	case abi.KindString:
		return encodeString(v)

	case abi.KindBool:
		return encodeBool(v)

	case abi.KindAddress:
		return encodeAddress(v)

	case abi.KindInt, abi.KindUInt:
		return encodeNum(v,t)

	case abi.KindBytes:
		return encodeBytes(v)

	case abi.KindFixedBytes, abi.KindFunction:
		return encodeFixedBytes(v,t)

	default:
		return nil, fmt.Errorf("encoding not available for type '%s'", t.Kind())
	}
}

func getTypeSize(t *abi.Type) int {
	if t.Kind() == abi.KindArray && !isDynamicType(t.Elem()) {
		if t.Elem().Kind() == abi.KindArray || t.Elem().Kind() == abi.KindTuple {
			return t.Size() * getTypeSize(t.Elem())
		}
		return t.Size() // * 32
	} else if t.Kind() == abi.KindTuple && !isDynamicType(t.Elem()) {
		total := 0
		for _, elem := range t.TupleElems() {
			total += getTypeSize(elem.Elem)
		}
		return total
	}
	return t.Size() // 32
}

func encodeSliceAndArray(v reflect.Value, t *abi.Type) ([]byte, error) {
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, encodeErr(v, t.Kind().String())
	}

	if v.Kind() == reflect.Array && t.Kind() != abi.KindArray {
		return nil, fmt.Errorf("expected array")
	} else if v.Kind() == reflect.Slice && t.Kind() != abi.KindSlice {
		return nil, fmt.Errorf("expected slice")
	}

	if t.Kind() == abi.KindArray && t.Size() != v.Len() {
		return nil, fmt.Errorf("array len incompatible")
	}

	var ret, tail []byte

	for i := 0; i < v.Len(); i++ {
		val, err := encode(v.Index(i), t.Elem())
		if err != nil {
			return nil, err
		}
		ret = append(ret, val...)
	}
	return append(ret, tail...), nil
}

func encodeTuple(v reflect.Value, t *abi.Type) ([]byte, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var err error
	isList := true

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
	case reflect.Map:
		isList = false

	case reflect.Struct:
		isList = false
		v, err = mapFromStruct(v)
		if err != nil {
			return nil, err
		}

	default:
		return nil, encodeErr(v, "tuple")
	}

	if v.Len() < len(t.TupleElems()) {
		return nil, fmt.Errorf("expected at least the same length")
	}

	var ret, tail []byte
	var aux reflect.Value

	for i, elem := range t.TupleElems() {
		if isList {
			aux = v.Index(i)
		} else {
			name := elem.Name
			if name == "" {
				name = strconv.Itoa(i)
			}
			aux = v.MapIndex(reflect.ValueOf(name))
		}
		if aux.Kind() == reflect.Invalid {
			return nil, fmt.Errorf("cannot get key %s", elem.Name)
		}

		val, err := encode(aux, elem.Elem)
		if err != nil {
			return nil, err
		}
		ret = append(ret, val...)
	}

	return append(ret, tail...), nil
}

func convertArrayToBytes(value reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.TypeOf([]byte{}), value.Len(), value.Len())
	reflect.Copy(slice, value)
	return slice
}

func encodeFixedBytes(v reflect.Value, t *abi.Type) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	if v.Kind() == reflect.String {
		value, err := decodeHex(v.String())
		if err != nil {
			return nil, err
		}

		v = reflect.ValueOf(value)
	}
	return rightPad(v.Bytes(), t.Size()), nil
}

func encodeAddress(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	if v.Kind() == reflect.String {
		var addr ethgo.Address
		if err := addr.UnmarshalText([]byte(v.String())); err != nil {
			return nil, err
		}
		v = reflect.ValueOf(addr.Bytes())
	}
	return v.Bytes(), nil
}

func encodeBytes(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	if v.Kind() == reflect.String {
		value, err := decodeHex(v.String())
		if err != nil {
			return nil, err
		}

		v = reflect.ValueOf(value)
	}
	return v.Bytes(), nil
}

func encodeString(v reflect.Value) ([]byte, error) {
	if v.Kind() != reflect.String {
		return nil, encodeErr(v, "string")
	}
	return []byte(v.String()), nil
}

func encodeNum(v reflect.Value, t *abi.Type) ([]byte, error) {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return toUSize(new(big.Int).SetUint64(v.Uint()),t.Size()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return toUSize(big.NewInt(v.Int()),t.Size()), nil

	case reflect.Ptr:
		if v.Type() != bigIntT {
			return nil, encodeErr(v.Elem(), "number")
		}
		return toUSize(v.Interface().(*big.Int),256), nil

	case reflect.Float64:
		return encodeNum(reflect.ValueOf(int64(v.Float())),t)

	case reflect.String:
		n, ok := new(big.Int).SetString(v.String(), 10)
		if !ok {
			n, ok = new(big.Int).SetString(v.String()[2:], 16)
			if !ok {
				return nil, encodeErr(v, "number")
			}
		}
		return encodeNum(reflect.ValueOf(n),t)

	default:
		return nil, encodeErr(v, "number")
	}
}

func encodeBool(v reflect.Value) ([]byte, error) {
	if v.Kind() != reflect.Bool {
		return nil, encodeErr(v, "bool")
	}
	if v.Bool() {
		return one.Bytes(), nil
	}
	return zero.Bytes(), nil
}

func encodeErr(v reflect.Value, t string) error {
	return fmt.Errorf("failed to encode %s as %s", v.Kind().String(), t)
}

func mapFromStruct(v reflect.Value) (reflect.Value, error) {
	res := map[string]interface{}{}
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue
		}

		tagValue := f.Tag.Get("abi")
		if tagValue == "-" {
			continue
		}

		name := strings.ToLower(f.Name)
		if tagValue != "" {
			name = tagValue
		}
		if _, ok := res[name]; !ok {
			res[name] = v.Field(i).Interface()
		}
	}
	return reflect.ValueOf(res), nil
}

var (
	tt256   = new(big.Int).Lsh(big.NewInt(1), 256)   // 2 ** 256
	tt256m1 = new(big.Int).Sub(tt256, big.NewInt(1)) // 2 ** 256 - 1
)

func toUSize(n *big.Int, size int) []byte {
	b := new(big.Int)
	b = b.Set(n)

	if b.Sign() < 0 || b.BitLen() > size {
		tt   := new(big.Int).Lsh(big.NewInt(1), uint(size))   // 2 ** 256
		ttm1 := new(big.Int).Sub(tt, big.NewInt(1)) // 2 ** 256 - 1
		b.And(b, ttm1)
	}

	return leftPad(b.Bytes(), size / 8)
}


func padBytes(b []byte, size int, left bool) []byte {
	l := len(b)
	if l == size {
		return b
	}
	if l > size {
		return b[l-size:]
	}
	tmp := make([]byte, size)
	if left {
		copy(tmp[size-l:], b)
	} else {
		copy(tmp, b)
	}
	return tmp
}

func leftPad(b []byte, size int) []byte {
	return padBytes(b, size, true)
}

func rightPad(b []byte, size int) []byte {
	return padBytes(b, size, false)
}

func encodeHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func decodeHex(str string) ([]byte, error) {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	buf, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("could not decode hex: %v", err)
	}
	return buf, nil
}