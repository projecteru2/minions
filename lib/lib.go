package lib

import (
	"context"
	"net"

	"github.com/coreos/etcd/clientv3"
	libcalico "github.com/projectcalico/libcalico-go/lib/clientv3"
	caliconet "github.com/projectcalico/libcalico-go/lib/net"
	log "github.com/sirupsen/logrus"
)

// Client .
type Client interface {
	ReleaseReservedIPByTiedContainerIDIfIdle(containerID string) error
}

type client struct {
	etcd   EtcdClient
	calico libcalico.Interface
}

// NewClient .
func NewClient(etcdV3 *clientv3.Client, calico libcalico.Interface) Client {
	return client{EtcdClient{etcdV3}, calico}
}

func (client client) ReleaseReservedIPByTiedContainerIDIfIdle(containerID string) error {
	log.Infof("Release reserved IP by tied containerID(%s)\n", containerID)

	container := Container{ID: containerID}
	if present, err := client.etcd.GetAndDelete(&container); err != nil {
		return err
	} else if !present {
		log.Infof("the container(%s) is not exists, will do nothing\n", containerID)
		return nil
	}
	if container.Address == "" {
		log.Infof("the ip of container(%s) is empty, will do nothing\n", containerID)
		return nil
	}
	address := ReservedIPAddress{Address: container.Address}
	log.Infof("aquiring reserved address(%s)\n", address.Address)
	if present, err := client.etcd.GetAndDelete(&address); err != nil {
		log.Errorf("aquiring reserved address(%s) error", address.Address)
		return err
	} else if !present {
		log.Infof("reserved address(%s) has already been released or reallocated\n", address.Address)
		return nil
	}

	log.Infof("release ip(%s) to calico pools\n", address.Address)
	ip := caliconet.IP{IP: net.ParseIP(address.Address)}
	if _, err := client.calico.IPAM().ReleaseIPs(context.Background(), []caliconet.IP{ip}); err != nil {
		log.Errorf("IP releasing error, ip: %v\n", ip)
		return err
	}

	return nil
}
