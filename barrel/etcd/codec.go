package etcd

import (
	"encoding/json"
	"fmt"

	"github.com/projecteru2/minions/types"
)

// ReservedAddressCodec .
type ReservedAddressCodec struct {
	Address *types.ReservedAddress
}

// ContainerInfoCodec .
type ContainerInfoCodec struct {
	Info *types.ContainerInfo
}

// ReserveRequestCodec .
type ReserveRequestCodec struct {
	Request *types.ReserveRequest
}

// Key .
func (codec ReservedAddressCodec) Key() string {
	if codec.Address.Address == "" {
		return ""
	}
	if codec.Address.PoolID == "" {
		return fmt.Sprintf("/barrel/addresses/%s", codec.Address.Address)
	}
	return fmt.Sprintf("/barrel/pools/%s/addresses/%s", codec.Address.PoolID, codec.Address.Address)
}

// Encode .
func (codec ReservedAddressCodec) Encode() (string, error) {
	return marshal(codec.Address)
}

// Decode .
func (codec ReservedAddressCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Address)
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

// Decode .
func (codec ContainerInfoCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Info)
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

// Decode .
func (codec ReserveRequestCodec) Decode(input string) error {
	return json.Unmarshal([]byte(input), codec.Request)
}

func marshal(src interface{}) (string, error) {
	bytes, err := json.Marshal(src)
	return string(bytes), err
}
