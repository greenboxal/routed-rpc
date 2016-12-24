package memberlist

type Config struct {
	// Node name
	Name string

	// Endpoint to bind the whisp service
	WhispBindEndpoint string

	// Endpoint to bind the RPC service
	RpcBindEndpoint string

	// Endpoint to advertise to whisp service to other peers
	WhispAdvertiseEndpoint string

	// Endpoint to advertise the RPC service to other peers
	RpcAdvertiseEndpoint string
}
