// based on the https://github.com/umbracle/ethgo/blob/main/abi/encode.go
package abi

import (
	"math/big"
	"reflect"
	"github.com/umbracle/ethgo/abi"
)

// batch of predefined reflect types
var (
	bigIntT       = reflect.TypeOf(new(big.Int))
)

type Type = abi.Type //[20]byte

func isDynamicType(t *Type) bool {
	if t.Kind() == abi.KindTuple {
		for _, elem := range t.TupleElems() {
			if isDynamicType(elem.Elem) {
				return true
			}
		}
		return false
	}
	return t.Kind() == abi.KindString || t.Kind() == abi.KindBytes || t.Kind() == abi.KindSlice || (t.Kind() == abi.KindArray && isDynamicType(t.Elem()))
}

// NewType parses a type in string format
func NewType(s string) (*Type, error) {
	return abi.NewType(s)
}

// MustNewType parses a type in string format or panics if its invalid
func MustNewType(s string) *Type {
	return abi.MustNewType(s)
}
