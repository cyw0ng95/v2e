//go:build CONFIG_USE_SONIC

package jsonutil

import (
	"github.com/bytedance/sonic"
)

var sonicFast = sonic.ConfigFastest

func Marshal(v interface{}) ([]byte, error) {
	return sonicFast.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return sonicFast.Unmarshal(data, v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return sonicFast.MarshalIndent(v, prefix, indent)
}
