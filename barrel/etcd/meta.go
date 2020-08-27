package etcd

import (
	"context"

	"github.com/projecteru2/minions/types"
)

// ReserveIPforContainer .
func (e *Etcd) ReserveIPforContainer(ctx context.Context, address *types.ReservedAddress, containerID string) error {
	container := &types.ContainerInfo{
		ID: containerID,
		ReservedAddress: types.ReservedAddress{
			PoolID:  address.PoolID,
			Address: address.Address,
		},
	}
	return e.PutMulti(ctx, ContainerInfoCodec{container}, ReservedAddressCodec{address})
}

// IPIsReserved .
func (e *Etcd) IPIsReserved(ctx context.Context, address *types.ReservedAddress) (bool, error) {
	return e.Get(ctx, ReservedAddressCodec{address})
}

// ConsumeRequestMarkIfPresent .
func (e *Etcd) ConsumeRequestMarkIfPresent(ctx context.Context, request *types.ReserveRequest) (bool, error) {
	return e.Delete(ctx, ReserveRequestCodec{request})
}

// AquireIfReserved .
func (e *Etcd) AquireIfReserved(ctx context.Context, address *types.ReservedAddress) (bool, error) {
	return e.Delete(ctx, ReservedAddressCodec{address})
}
