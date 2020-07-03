package lib

import (
	"time"

	"github.com/coreos/etcd/clientv3"
)

const (
	clientTimeout    = 10 * time.Second
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
)

// NewEtcdClient .
func NewEtcdClient(endpoints []string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          clientTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
	})
}
