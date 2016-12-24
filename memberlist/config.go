package memberlist

type Config struct {
	Name string

	BindAddr      string
	WhispBindPort int
	RpcBindPort   int

	AdvertiseAddr      string
	WhispAdvertisePort int
	RpcAdvertisePort   int
}
