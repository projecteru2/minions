package lib

import (
	"context"
	"fmt"
	"net"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	libcalico "github.com/projectcalico/libcalico-go/lib/clientv3"
	calicoipam "github.com/projectcalico/libcalico-go/lib/ipam"
	caliconet "github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/options"
	osutils "github.com/projectcalico/libnetwork-plugin/utils/os"
	log "github.com/sirupsen/logrus"
)

// Client .
type Client interface {
	RequestFixedIP(poolID string) (string, error)
	MarkFixedIPForContainer(containerID string, address string) error
	ReleaseReservedIPByTiedContainerIDIfIdle(containerID string) error
	MarkReserveRequestForIP(ip string) error
}

type client struct {
	etcd     EtcdClient
	calico   libcalico.Interface
	PoolIDV4 string
	PoolIDV6 string
}

// NewClient .
func NewClient(etcdV3 *clientv3.Client, calico libcalico.Interface) Client {
	return client{EtcdClient{etcdV3}, calico, "CalicoPoolIPv4", "CalicoPoolIPv6"}
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

func (client client) MarkReserveRequestForIP(ip string) (err error) {
	var reserved bool
	request := ReservedIPAddress{Address: ip}
	if reserved, err = client.etcd.Get(&request); reserved || err != nil {
		return
	}
	return client.etcd.Put(&request)
}

func (client client) RequestFixedIP(poolID string) (address string, err error) {
	log.Infof("Auto assigning IP from Calico pools, poolID = %s", poolID)

	var hostname string
	if hostname, err = osutils.GetHostname(); err != nil {
		return
	}

	var IPs []caliconet.IP

	// If the poolID isn't the fixed one then find the pool to assign from.
	// poolV4 defaults to nil to assign from across all pools.
	var poolV4 []caliconet.IPNet

	var poolV6 []caliconet.IPNet
	var numIPv4, numIPv6 int
	if poolID == client.PoolIDV4 {
		numIPv4 = 1
		numIPv6 = 0
	} else if poolID == client.PoolIDV6 {
		numIPv4 = 0
		numIPv6 = 1
	} else {
		var version int
		poolsClient := client.calico.IPPools()
		var ipPool *apiv3.IPPool
		if ipPool, err = poolsClient.Get(context.Background(), poolID, options.GetOptions{}); err != nil {
			log.Errorf("Invalid Pool - %v", poolID)
			return
		}

		var ipNet *caliconet.IPNet
		if _, ipNet, err = caliconet.ParseCIDR(ipPool.Spec.CIDR); err != nil {
			log.Errorf("Invalid CIDR - %v", poolID)
			return
		}

		version = ipNet.Version()
		if version == 4 {
			poolV4 = []caliconet.IPNet{{IPNet: ipNet.IPNet}}
			numIPv4 = 1
			log.Debugln("Using specific pool ", poolV4)
		} else if version == 6 {
			poolV6 = []caliconet.IPNet{{IPNet: ipNet.IPNet}}
			numIPv6 = 1
			log.Debugln("Using specific pool ", poolV6)
		}
	}

	// Auto assign an IP address.
	// IPv4/v6 pool will be nil if the docker network doesn't have a subnet associated with.
	// Otherwise, it will be set to the Calico pool to assign from.
	var IPsV4 []caliconet.IP
	var IPsV6 []caliconet.IP
	if IPsV4, IPsV6, err = client.calico.IPAM().AutoAssign(
		context.Background(),
		calicoipam.AutoAssignArgs{
			Num4:      numIPv4,
			Num6:      numIPv6,
			Hostname:  hostname,
			IPv4Pools: poolV4,
			IPv6Pools: poolV6,
		},
	); err != nil {
		log.Errorln("IP assignment error")
		return
	}
	IPs = append(IPsV4, IPsV6...)

	// We should only have one IP address assigned at this point.
	if len(IPs) != 1 {
		err = errors.Errorf("Unexpected number of assigned IP addresses. "+
			"A single address should be assigned. Got %v", IPs)
		return
	}
	return client.reserveAndFormatIPAddress(IPs[0])
}

func (client client) reserveAndFormatIPAddress(ip caliconet.IP) (result string, err error) {
	if ip.Version() == 4 {
		// IPv4 address
		result = fmt.Sprintf("%v/%v", ip, "32")
	} else {
		// IPv6 address
		result = fmt.Sprintf("%v/%v", ip, "128")
	}
	address := fmt.Sprintf("%v", ip)
	log.Infof("[MinionsClient.reserveAndFormatIPAddress] request ip %s success, reserving...", address)
	err = client.etcd.Put(&ReservedIPAddress{Address: address})
	return
}

func (client client) MarkFixedIPForContainer(containerID string, address string) (err error) {
	return client.etcd.Put(&Container{ID: containerID, Address: address})
}
