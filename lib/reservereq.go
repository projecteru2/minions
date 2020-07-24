package lib

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

// ReserveRequest .
type ReserveRequest struct {
	Address string
	version int64
}

// Key .
func (req *ReserveRequest) Key() string {
	if req.Address == "" {
		return ""
	}
	return fmt.Sprintf("/barrel/reservereq/%s", req.Address)
}

// Read .
func (req *ReserveRequest) Read(ekv *mvccpb.KeyValue) error {
	req.version = ekv.Version
	return json.Unmarshal(ekv.Value, req)
}

// JSON .
func (req *ReserveRequest) JSON() string {
	return fmt.Sprintf(`{"Address":"%s"}`, req.Address)
}

// Version .
func (req *ReserveRequest) Version() int64 {
	return req.version
}
