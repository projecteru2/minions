package etcd

import (
	"encoding/json"
	"fmt"

	"github.com/projecteru2/minions/types"
)

// ReservedAddressCodec .
type ReservedAddressCodec struct {
	Address *types.ReservedAddress
	version int64
}

// Key .
func (codec *ReservedAddressCodec) Key() string {
	if codec.Address.Address == "" {
		return ""
	}
	if codec.Address.PoolID == "" {
		return fmt.Sprintf("/barrel/addresses/%s", codec.Address.Address)
	}
	return fmt.Sprintf("/barrel/pools/%s/addresses/%s", codec.Address.PoolID, codec.Address.Address)
}

// Encode .
func (codec *ReservedAddressCodec) Encode() (string, error) {
	return marshal(codec.Address)
}

// SetVersion .
func (codec *ReservedAddressCodec) SetVersion(version int64) {
	codec.version = version
}

// Version .
func (codec *ReservedAddressCodec) Version() int64 {
	return codec.version
}

// Decode .
func (codec ReservedAddressCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Address)
}

// ContainerInfoCodec .
type ContainerInfoCodec struct {
	Info    *types.ContainerInfo
	version int64
}

// Key .
func (codec ContainerInfoCodec) Key() string {
	if codec.Info.ID == "" {
		return ""
	}
	return fmt.Sprintf("/barrel/containers/%s", codec.Info.ID)
}

// Encode .
func (codec ContainerInfoCodec) Encode() (string, error) {
	return marshal(codec.Info)
}

// SetVersion .
func (codec *ContainerInfoCodec) SetVersion(version int64) {
	codec.version = version
}

// Version .
func (codec *ContainerInfoCodec) Version() int64 {
	return codec.version
}

// Decode .
func (codec ContainerInfoCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Info)
}

// ReserveRequestCodec .
type ReserveRequestCodec struct {
	Request *types.ReserveRequest
	version int64
}

// Key .
func (codec ReserveRequestCodec) Key() string {
	if codec.Request.Address == "" {
		return ""
	}
	if codec.Request.PoolID == "" {
		return fmt.Sprintf("/barrel/reservereqs/%s", codec.Request.Address)
	}
	return fmt.Sprintf("/barrel/pools/%s/reservereqs/%s", codec.Request.PoolID, codec.Request.Address)
}

// Encode .
func (codec ReserveRequestCodec) Encode() (string, error) {
	return marshal(codec.Request)
}

// SetVersion .
func (codec *ReserveRequestCodec) SetVersion(version int64) {
	codec.version = version
}

// Version .
func (codec *ReserveRequestCodec) Version() int64 {
	return codec.version
}

// Decode .
func (codec ReserveRequestCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Request)
}

func marshal(src interface{}) (string, error) {
	bytes, err := json.Marshal(src)
	return string(bytes), err
}
