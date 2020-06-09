package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
)

const fixedIPLabel = "fixed-ip"
const inheritLabel = "ip-inherit"

type reservedIPAddress struct {
	Address   string
	Allocated bool
}

// IPAddressManager .
type ReservedIPManager interface {
	MarkReservedIP(ctx context.Context, ipAddress string) error
	CheckAndAquireReservedIP(ctx context.Context, ipAddress string) (bool, error)
	IsReservedIP(ctx context.Context, ipAddress string) (bool, error)
	ReleaseReservedIP(ctx context.Context, ipAddress string) error
}

type reservedIPManager struct {
	etcdV3 *clientv3.Client
}

func pathToIPAddress(ipAddress string) string {
	return fmt.Sprintf("/barrel/reserved/ip/%s", ipAddress)
}

func kvToModel(ekv *mvccpb.KeyValue) (reservedIPAddress, error) {
	result := reservedIPAddress{}
	err := json.Unmarshal(ekv.Value, &result)
	return result, err
}

func modelToKv(model reservedIPAddress) (string, string) {
	key := pathToIPAddress(model.Address)
	// adopt json.Marshal will introduce handling error, so use json template
	return key, fmt.Sprintf(`{"Address":"%s", "Allocated":%v}`, model.Address, model.Allocated)
}

// New .
func NewReservedIPManager(cfg clientv3.Config) (ReservedIPManager, error) {
	etcdV3Client, err := clientv3.New(cfg)
	if err != nil {
		return reservedIPManager{}, err
	}
	return reservedIPManager{etcdV3: etcdV3Client}, nil
}

// MarkReservedIP .
func (ripam reservedIPManager) MarkReservedIP(ctx context.Context, ipAddress string) error {
	key, value := modelToKv(reservedIPAddress{ipAddress, false})
	_, err := ripam.etcdV3.Put(context.Background(), key, value)
	return err
}

func (ripam reservedIPManager) IsReservedIP(ctx context.Context, ipAddress string) (bool, error) {
	resp, err := ripam.etcdV3.Get(context.Background(), pathToIPAddress(ipAddress))
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) == 0, nil
}

// IsReservedIP .
func (ripam reservedIPManager) checkIPForReservedStatus(ctx context.Context, ipAddress string) (bool, int64, error) {
	resp, err := ripam.etcdV3.Get(ctx, pathToIPAddress(ipAddress))
	if err != nil {
		return false, 0, errors.Wrap(err, "get IP reserved status error")
	}
	kvs := resp.Kvs
	if len(kvs) == 0 {
		return false, 0, nil
	}
	kv := kvs[0]
	model, err := kvToModel(kv)
	if err != nil {
		return false, 0, err
	}
	if model.Allocated {
		return false, 0, errors.Errorf("Specified reserved IP is already allocated, ip = %s", ipAddress)
	}
	return true, kv.Version, nil
}

// AquireReservedIP .
func (ripam reservedIPManager) CheckAndAquireReservedIP(ctx context.Context, ipAddress string) (bool, error) {
	reserved, version, err := ripam.checkIPForReservedStatus(ctx, ipAddress)
	if err != nil {
		return false, err
	}
	if !reserved {
		return false, nil
	}
	model := reservedIPAddress{Address: ipAddress, Allocated: true}
	key, value := modelToKv(model)
	conds := []clientv3.Cmp{clientv3.Compare(clientv3.Version(key), "=", version)}
	resp, err := ripam.etcdV3.Txn(context.Background()).If(
		conds...,
	).Then(
		clientv3.OpPut(key, value),
	).Commit()
	if err != nil {
		return true, errors.Wrap(err, "Commit etcd transaction error")
	}
	if resp.Succeeded {
		return true, nil
	}
	return true, errors.Errorf("Aquire reserved ip error, ip = %s, version = %d", ipAddress, version)
}

// ReleaseReservedIP .
func (ripam reservedIPManager) ReleaseReservedIP(ctx context.Context, ipAddress string) error {
	if _, err := ripam.etcdV3.Delete(context.Background(), pathToIPAddress(ipAddress)); err != nil {
		return err
	}
	return nil
}
