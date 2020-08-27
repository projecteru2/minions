package driver

import (
	"context"
	"fmt"
	"net"

	pluginIPAM "github.com/docker/go-plugins-helpers/ipam"
	"github.com/pkg/errors"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	caliconet "github.com/projectcalico/libcalico-go/lib/net"
	logutils "github.com/projectcalico/libnetwork-plugin/utils/log"

	barrelMeta "github.com/projecteru2/minions/barrel"
	calDriver "github.com/projecteru2/minions/driver/calico"
	calIpamDriver "github.com/projecteru2/minions/driver/calico/ipam"
	"github.com/projecteru2/minions/types"

	log "github.com/sirupsen/logrus"
)

// IPAMDriver .
type IPAMDriver struct {
	calicoIPAM *calIpamDriver.CalicoIPAM
	meta       barrelMeta.Meta
}

// NewIPAMDriver .
func NewIPAMDriver(
	clientv3 clientv3.Interface,
	meta barrelMeta.Meta,
) pluginIPAM.Ipam {
	return &IPAMDriver{
		calicoIPAM: calIpamDriver.NewCalicoIPAM(clientv3),
		meta:       meta,
	}
}

// GetCapabilities .
func (i IPAMDriver) GetCapabilities() (*pluginIPAM.CapabilitiesResponse, error) {
	resp := pluginIPAM.CapabilitiesResponse{}
	logutils.JSONMessage("GetCapabilities response", resp)
	return &resp, nil
}

// GetDefaultAddressSpaces .
func (i IPAMDriver) GetDefaultAddressSpaces() (*pluginIPAM.AddressSpacesResponse, error) {
	resp := &pluginIPAM.AddressSpacesResponse{
		LocalDefaultAddressSpace:  calDriver.CalicoLocalAddressSpace,
		GlobalDefaultAddressSpace: calDriver.CalicoGlobalAddressSpace,
	}
	logutils.JSONMessage("GetDefaultAddressSpace response", resp)
	return resp, nil
}

// RequestPool .
func (i IPAMDriver) RequestPool(request *pluginIPAM.RequestPoolRequest) (*pluginIPAM.RequestPoolResponse, error) {
	logutils.JSONMessage("RequestPool", request)

	// Calico IPAM does not allow you to request a SubPool.
	if request.SubPool != "" {
		err := errors.New(
			"Calico IPAM does not support sub pool configuration " +
				"on 'docker create network'. Calico IP Pools " +
				"should be configured first and IP assignment is " +
				"from those pre-configured pools.",
		)
		log.Errorln(err)
		return nil, err
	}

	if len(request.Options) != 0 {
		err := errors.New("Arbitrary options are not supported")
		log.Errorln(err)
		return nil, err
	}

	var (
		pool *types.Pool
		err  error
	)

	// If a pool (subnet on the CLI) is specified, it must match one of the
	// preconfigured Calico pools.
	if request.Pool != "" {
		if pool, err = i.calicoIPAM.RequestPool(request.Pool); err != nil {
			log.Errorf("[IPAMDriver::RequestPool] request calico pool error, %v", err)
			return nil, err
		}
	} else {
		pool = i.calicoIPAM.RequestDefaultPool(request.V6)
	}

	// We use static pool ID and CIDR. We don't need to signal the
	// The meta data includes a dummy gateway address. This prevents libnetwork
	// from requesting a gateway address from the pool since for a Calico
	// network our gateway is set to a special IP.
	resp := &pluginIPAM.RequestPoolResponse{
		PoolID: pool.Name,
		Pool:   pool.CIDR,
		Data:   map[string]string{"com.docker.network.gateway": pool.Gateway},
	}
	logutils.JSONMessage("RequestPool response", resp)
	return resp, nil
}

// ReleasePool .
func (i IPAMDriver) ReleasePool(request *pluginIPAM.ReleasePoolRequest) error {
	logutils.JSONMessage("ReleasePool", request)
	return nil
}

// RequestAddress .
func (i IPAMDriver) RequestAddress(request *pluginIPAM.RequestAddressRequest) (*pluginIPAM.RequestAddressResponse, error) {
	logutils.JSONMessage("RequestAddress", request)

	// Calico IPAM does not allow you to choose a gateway.
	if err := checkOptions(request.Options); err != nil {
		log.Errorf("[IpamDriver::RequestAddress] check request options failed, %v", err)
		return nil, err
	}

	var address caliconet.IP
	var err error
	if address, err = i.requestIP(request); err != nil {
		return nil, err
	}

	// we should remove the request mark
	ip := fmt.Sprintf("%v", address)
	if _, err := i.meta.ConsumeRequestMarkIfPresent(
		context.Background(),
		&types.ReserveRequest{
			ReservedAddress: types.ReservedAddress{
				PoolID:  request.PoolID,
				Address: ip,
			},
		}); err != nil {
		// Do not continue, or else the mark will cause some undefined behavior
		log.Errorf("[IPAM.RequestAddress] remove request mark of ip(%v) error, %v", ip, err)
		return nil, err
	}
	log.Infof("[IPAM.RequestAddress] removed request mark on ip(%v) allocated", ip)

	resp := &pluginIPAM.RequestAddressResponse{
		// Return the IP as a CIDR.
		Address: formatIPAddress(address),
	}
	logutils.JSONMessage("RequestAddress response", resp)
	return resp, nil
}

// ReleaseAddress .
func (i IPAMDriver) ReleaseAddress(request *pluginIPAM.ReleaseAddressRequest) error {
	logutils.JSONMessage("ReleaseAddress", request)
	reserved, err := i.meta.IPIsReserved(
		context.Background(),
		&types.ReservedAddress{
			PoolID:  request.PoolID,
			Address: request.Address,
		},
	)
	if err != nil {
		log.Errorf("Get reserved ip status error, ip: %v", request.Address)
		return err
	}

	if reserved {
		log.Infof("Ip is reserved, will not release to pool, ip: %v\n", request.Address)
		return nil
	}
	return i.calicoIPAM.ReleaseIP(request.PoolID, request.Address)
}

func (i IPAMDriver) requestIP(request *pluginIPAM.RequestAddressRequest) (caliconet.IP, error) {
	if request.Address == "" {
		return i.calicoIPAM.AutoAssign(request.PoolID)
	}
	var err error

	// specified address requested, so will try assign from reserved pool, then calico pool
	log.Info("Assigning specified IP from reserved pool first, then calico pools")

	// try to acquire ip from reserved ip pool
	var acquired bool
	if acquired, err = i.meta.AquireIfReserved(
		context.Background(),
		&types.ReservedAddress{
			PoolID:  request.PoolID,
			Address: request.Address,
		}); err != nil {
		return caliconet.IP{}, err
	}
	if acquired {
		return caliconet.IP{IP: net.ParseIP(request.Address)}, nil
	}
	// assign IP from calico
	return i.calicoIPAM.AssignIP(request.PoolID, request.Address)
}
