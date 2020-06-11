package lib

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

// ReservedIPAddress .
type ReservedIPAddress struct {
	Address string
	version int64
}

// Key .
func (addr *ReservedIPAddress) Key() string {
	if addr.Address == "" {
		return ""
	}
	return fmt.Sprintf("/barrel/addresses/%s", addr.Address)
}

// Read .
func (addr *ReservedIPAddress) Read(ekv *mvccpb.KeyValue) error {
	addr.version = ekv.Version
	return json.Unmarshal(ekv.Value, addr)
}

// JSON .
func (addr *ReservedIPAddress) JSON() string {
	return fmt.Sprintf(`{"Address":"%s"}`, addr.Address)
}

// Version .
func (addr *ReservedIPAddress) Version() int64 {
	return addr.version
}
