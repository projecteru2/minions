package types

import "github.com/pkg/errors"

var (
	ErrNoOps         = errors.New("No ops")
	ErrCIDRNotInPool = errors.New("The requested subnet must match the CIDR of a configured Calico IP Pool")
)
