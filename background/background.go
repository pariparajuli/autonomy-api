package background

import (
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/vmihailenco/msgpack/v4"
)

// Background is a struct to maintain common clients
// and functions for all background workers
type Background struct {
	Onesignal *onesignal.OneSignalClient
}

// MsgPackDataConverter is DataConverter that uses msgpack to
// encode and decode data between workflows and activities
type MsgPackDataConverter struct{}

func NewMsgPackDataConverter() *MsgPackDataConverter {
	return &MsgPackDataConverter{}
}

// ToData implements conversion of a list of values.
func (c *MsgPackDataConverter) ToData(value ...interface{}) ([]byte, error) {
	data := [][]byte{}
	for _, v := range value {
		if b, err := msgpack.Marshal(v); err != nil {
			return nil, err
		} else {
			data = append(data, b)
		}
	}

	return msgpack.Marshal(data)
}

// FromData implements conversion of an array of values of different types.
// Useful for deserializing arguments of function invocations.
func (c *MsgPackDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	data := [][]byte{}
	if err := msgpack.Unmarshal(input, &data); err != nil {
		return err
	}

	for i, d := range data {
		if err := msgpack.Unmarshal(d, valuePtr[i]); err != nil {
			return err
		}
	}
	return nil
}
