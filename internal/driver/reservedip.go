package driver

import (
	log "github.com/sirupsen/logrus"

	"github.com/coreos/etcd/clientv3"
	"github.com/projecteru2/minions/lib"
)

const fixedIPLabel = "fixed-ip"

type reservedIPManager struct {
	etcd lib.EtcdClient
}

// ReservedIPManager .
type ReservedIPManager interface {
	Reserve(ip string, containerID string) error
	AquireIfReserved(ip string) (bool, error)
	IsReserved(ip string) (bool, error)
}

// NewReservedIPManager .
func NewReservedIPManager(etcdV3Client *clientv3.Client) ReservedIPManager {
	return reservedIPManager{etcd: lib.EtcdClient{Etcd: etcdV3Client}}
}

// Reserve .
func (ripam reservedIPManager) Reserve(ip string, containerID string) error {
	log.Infof("Reserve ip = %s", ip)
	return ripam.etcd.PutMulti(&lib.ReservedIPAddress{Address: ip}, &lib.Container{ID: containerID, Address: ip})
}

// IsReserved .
func (ripam reservedIPManager) IsReserved(ip string) (bool, error) {
	return ripam.etcd.Get(&lib.ReservedIPAddress{Address: ip})
}

// AquireIfReserved .
func (ripam reservedIPManager) AquireIfReserved(ip string) (bool, error) {
	log.Infof("AquireIPIfReserved ip = %s", ip)
	return ripam.etcd.GetAndDelete(&lib.ReservedIPAddress{Address: ip})
}
