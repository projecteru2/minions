package barrel

import (
	"context"

	"github.com/projecteru2/minions/types"
)

// Meta .
type Meta interface {
	ReserveIPforContainer(ctx context.Context, address *types.ReservedAddress, ID string) error
	IPIsReserved(ctx context.Context, address *types.ReservedAddress) (bool, error)
	ConsumeRequestMarkIfPresent(ctx context.Context, request *types.ReserveRequest) (bool, error)
	AquireIfReserved(ctx context.Context, address *types.ReservedAddress) (bool, error)
}
