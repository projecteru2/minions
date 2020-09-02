package etcd

import (
	"context"

	"github.com/projecteru2/minions/types"
)

// ReserveIPforContainer .
func (e *Etcd) ReserveIPforContainer(ctx context.Context, address *types.ReservedAddress, containerID string) error {
	container := &types.ContainerInfo{
		ID: containerID,
		Addresses: []types.ReservedAddress{{
			PoolID:  address.PoolID,
			Address: address.Address,
		}},
	}
	return e.PutMulti(ctx, &ContainerInfoCodec{Info: container}, &ReservedAddressCodec{Address: address})
}

// IPIsReserved .
func (e *Etcd) IPIsReserved(ctx context.Context, address *types.ReservedAddress) (bool, error) {
	return e.Get(ctx, &ReservedAddressCodec{Address: address})
}

// ConsumeRequestMarkIfPresent .
func (e *Etcd) ConsumeRequestMarkIfPresent(ctx context.Context, request *types.ReserveRequest) (bool, error) {
	return e.Delete(ctx, &ReserveRequestCodec{Request: request})
}

// AquireIfReserved .
func (e *Etcd) AquireIfReserved(ctx context.Context, address *types.ReservedAddress) (bool, error) {
	return e.Delete(ctx, &ReservedAddressCodec{Address: address})
}
