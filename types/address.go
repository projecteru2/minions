package types

// ReservedAddress .
type ReservedAddress struct {
	PoolID  string
	Address string
}

// Container .
type ContainerInfo struct {
	ID string
	ReservedAddress
}

// ReserveRequest .
type ReserveRequest struct {
	ReservedAddress
}
