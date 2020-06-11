package lib

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

// Container .
type Container struct {
	ID      string
	Address string
	version int64
}

// Key .
func (container *Container) Key() string {
	if container.ID == "" {
		return ""
	}
	return fmt.Sprintf("/barrel/containers/%s", container.ID)
}

// Read .
func (container *Container) Read(ekv *mvccpb.KeyValue) error {
	container.version = ekv.Version
	return json.Unmarshal(ekv.Value, container)
}

// JSON .
func (container *Container) JSON() string {
	// adopt json.Marshal will introduce handling error, so use json template
	return fmt.Sprintf(`{"ID":"%s", "Address":"%s"}`, container.ID, container.Address)
}

// Version .
func (container *Container) Version() int64 {
	return container.version
}
