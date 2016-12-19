package routedrpc

// An AddressNotFound is returned when an address isn't found in the RPC system
type AddressNotFound struct {
}

// Returns a string representation of the error
func (a AddressNotFound) Error() string {
	return "address not found"
}
