package types

// ReservedAddress .
type ReservedAddress struct {
	PoolID  string
	Address string
}

// ContainerInfo .
type ContainerInfo struct {
	ID        string
	Addresses []ReservedAddress
}

// ReserveRequest .
type ReserveRequest struct {
	ReservedAddress
}
