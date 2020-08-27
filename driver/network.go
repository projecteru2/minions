package driver

import (
	"context"

	"github.com/docker/go-plugins-helpers/network"
	"github.com/pkg/errors"
	api "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	log "github.com/sirupsen/logrus"

	dockerTypes "github.com/docker/docker/api/types"
	dockerNetworkTypes "github.com/docker/docker/api/types/network"

	dockerClient "github.com/docker/docker/client"
	logutils "github.com/projectcalico/libnetwork-plugin/utils/log"
	"github.com/projecteru2/minions/barrel"
	calNetDriver "github.com/projecteru2/minions/driver/calico/network"
	"github.com/projecteru2/minions/types"
)

// NetworkDriver .
type NetworkDriver struct {
	calNetDriver calNetDriver.Driver
	dockerCli    *dockerClient.Client
	meta         barrel.Meta
}

// NewNetworkDriver .
func NewNetworkDriver(
	client clientv3.Interface,
	dockerCli *dockerClient.Client,
	meta barrel.Meta,
) network.Driver {
	return NetworkDriver{
		calNetDriver: calNetDriver.NewNetworkDriver(client, dockerCli),
		dockerCli:    dockerCli,
		meta:         meta,
	}
}

// GetCapabilities .
func (driver NetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	return driver.calNetDriver.GetCapabilities()
}

// AllocateNetwork .
func (driver NetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	return driver.calNetDriver.AllocateNetwork(request)
}

// FreeNetwork is used for swarm-mode support in remote plugins, which
// Calico's libnetwork-plugin doesn't currently support.
func (driver NetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	return driver.calNetDriver.FreeNetwork(request)
}

// CreateNetwork .
func (driver NetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	return driver.calNetDriver.CreateNetwork(request)
}

// DeleteNetwork .
func (driver NetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	return driver.calNetDriver.DeleteNetwork(request)
}

// CreateEndpoint .
func (driver NetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	return driver.calNetDriver.CreateEndpoint(request)
}

// DeleteEndpoint .
func (driver NetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	return driver.calNetDriver.DeleteEndpoint(request)
}

// EndpointInfo .
func (driver NetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	return driver.calNetDriver.EndpointInfo(request)
}

// Join .
func (driver NetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	return driver.calNetDriver.Join(request)
}

// Leave .
func (driver NetworkDriver) Leave(request *network.LeaveRequest) error {
	logutils.JSONMessage("Leave response", request)
	var (
		container        dockerTypes.Container
		endpointSettings *dockerNetworkTypes.EndpointSettings
		pool             *api.IPPool
		shouldReserveIP  bool
		err              error
	)
	if container, endpointSettings, err = driver.findDockerContainerByEndpointID(request.EndpointID); err != nil {
		return err
	}

	if pool, err = driver.calNetDriver.FindPoolByNetworkID(endpointSettings.NetworkID); err != nil {
		return err
	}

	if shouldReserveIP, err = driver.shouldReserveIP(
		container,
		&types.ReserveRequest{
			ReservedAddress: types.ReservedAddress{
				PoolID:  pool.Name,
				Address: endpointSettings.IPAddress,
			},
		}); err != nil {
		// we move on when trying to find out whether should reserve by reserve request mark
		log.Errorln(err)
	}
	if shouldReserveIP {
		if err = driver.meta.ReserveIPforContainer(
			context.Background(),
			&types.ReservedAddress{
				PoolID:  pool.Name,
				Address: endpointSettings.IPAddress,
			}, container.ID); err != nil {
			// we move on when reserve is failed
			log.Errorln(err)
		}
	}
	return driver.calNetDriver.Leave(request)
}

func (driver NetworkDriver) findDockerContainerByEndpointID(endpointID string) (dockerTypes.Container, *dockerNetworkTypes.EndpointSettings, error) {
	containers, err := driver.dockerCli.ContainerList(context.Background(), dockerTypes.ContainerListOptions{})
	if err != nil {
		log.Errorf("dockerCli ContainerList Error, %v", err)
		return dockerTypes.Container{}, nil, err
	}
	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			if endpointID == network.EndpointID {
				return container, network, nil
			}
		}
	}
	return dockerTypes.Container{}, nil, errors.Errorf("find no container with endpintID = %s", endpointID)
}

func (driver NetworkDriver) shouldReserveIP(container dockerTypes.Container, address *types.ReserveRequest) (shouldReserve bool, err error) {
	// reserve ip here by container label
	if containerHasFixedIPLabel(container) {
		shouldReserve = true
		// we should consume the mark
		if _, err := driver.meta.ConsumeRequestMarkIfPresent(context.Background(), address); err != nil {
			log.Errorf("[Network.ConsumeRequestMarkIfPresent] remove request mark error, %v", err)
		}
		log.Infof("[Network.ConsumeRequestMarkIfPresent] container has fixed-ip label, shouldReserve ip(%v) = %v", address, shouldReserve)
		return
	}
	// reserve ip here by reserve request mark
	if shouldReserve, err = driver.meta.ConsumeRequestMarkIfPresent(context.Background(), address); err != nil {
		// ensure shouldReserve is false here when err is not nil
		log.Errorf("[Network.ConsumeRequestMarkIfPresent] error, %v", err)
		shouldReserve = false
		return
	}
	var msg string
	if shouldReserve {
		msg = "marked as requested"
	} else {
		msg = "not marked as requested"
	}
	log.Infof("[Network.ConsumeRequestMarkIfPresent] address is %s, shouldReserve ip(%v) = %v", msg, address, shouldReserve)
	return
}

// DiscoverNew .
func (driver NetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	return driver.calNetDriver.DiscoverNew(request)
}

// DiscoverDelete .
func (driver NetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	return driver.calNetDriver.DiscoverDelete(request)
}

// ProgramExternalConnectivity .
func (driver NetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	return driver.calNetDriver.ProgramExternalConnectivity(request)
}

// RevokeExternalConnectivity .
func (driver NetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	return driver.calNetDriver.RevokeExternalConnectivity(request)
}
