package cadence

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack/v4"
)

// MsgPackDataConverter is DataConverter that uses msgpack to
// encode and decode data between workflows and activities
type MsgPackDataConverter struct{}

func NewMsgPackDataConverter() *MsgPackDataConverter {
	return &MsgPackDataConverter{}
}

// ToData implements conversion of a list of values.
func (c *MsgPackDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	for i, obj := range value {
		if err := enc.Encode(obj); err != nil {
			return nil, fmt.Errorf(
				"unable to encode argument: %d, %v, with msgpack error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return buf.Bytes(), nil
}

// FromData implements conversion of an array of values of different types.
// Useful for deserializing arguments of function invocations.
func (c *MsgPackDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewBuffer(input))
	for i, obj := range valuePtr {
		if err := dec.Decode(obj); err != nil {
			return fmt.Errorf(
				"unable to decode argument: %d, %v, with msgpack error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}
