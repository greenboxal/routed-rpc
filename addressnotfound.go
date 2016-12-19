package routedrpc

type AddressNotFound struct {
}

func (a AddressNotFound) Error() string {
	return "address not found"
}
