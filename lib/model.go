package lib

import (
	"context"
	"errors"

	etcdv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

// EtcdModel .
type EtcdModel interface {
	Key() string
	Read(ekv *mvccpb.KeyValue) error
	JSON() string
	Version() int64
}

// EtcdClient .
type EtcdClient struct {
	Etcd *etcdv3.Client
}

// Get .
func (cli EtcdClient) Get(model EtcdModel) (bool, error) {
	var err error
	key := model.Key()
	if key == "" {
		return false, errors.New("can't generate valid key from model")
	}

	var resp *etcdv3.GetResponse
	if resp, err = cli.Etcd.Get(context.Background(), key); err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, nil
	}

	return true, model.Read(resp.Kvs[0])
}

// Put .
func (cli EtcdClient) Put(model EtcdModel) (err error) {
	key := model.Key()
	if key == "" {
		err = errors.New("can't generate valid key from model")
		return
	}
	_, err = cli.Etcd.Put(context.Background(), key, model.JSON())
	return
}

// PutMulti .
func (cli EtcdClient) PutMulti(models ...EtcdModel) (err error) {
	var ops []etcdv3.Op
	for _, model := range models {
		key := model.Key()
		if key == "" {
			err = errors.New("can't generate valid key from model")
			return
		}
		ops = append(ops, etcdv3.OpPut(key, model.JSON()))
	}
	_, err = cli.Etcd.Txn(context.Background()).Then(ops...).Commit()
	return
}

// GetAndDelete .
func (cli EtcdClient) GetAndDelete(model EtcdModel) (get bool, err error) {
	key := model.Key()
	if key == "" {
		get = false
		err = errors.New("can't generate valid key from model")
		return
	}

	var resp *etcdv3.DeleteResponse
	if resp, err = cli.Etcd.Delete(context.Background(), key, etcdv3.WithPrevKV()); err != nil {
		return
	}
	if len(resp.PrevKvs) == 0 {
		return
	}
	get = true
	err = model.Read(resp.PrevKvs[0])
	return
}
