package memberlist

import (
	"fmt"
	"testing"
	"time"

	"github.com/greenboxal/routed-rpc"
	"github.com/stretchr/testify/assert"
)

func setupProvider(port int) (*Provider, error) {
	cfg := &Config{}

	cfg.Name = fmt.Sprintf("test_%d", port)
	cfg.WhispBindEndpoint = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.RPCBindEndpoint = fmt.Sprintf("127.0.0.1:%d", port+1)
	cfg.WhispAdvertiseEndpoint = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.RPCAdvertiseEndpoint = fmt.Sprintf("127.0.0.1:%d", port+1)

	return Create(cfg)
}

func TestProvider(t *testing.T) {
	providerA, err := setupProvider(10000)
	assert.Nil(t, err)
	assert.NotNil(t, providerA)

	providerB, err := setupProvider(20000)
	assert.Nil(t, err)
	assert.NotNil(t, providerB)

	_, err = providerB.Join([]string{"127.0.0.1:10000"})
	assert.Nil(t, err)

	handlerA := routedrpc.NewMockHandler("player_a")
	rpcA := routedrpc.Create(&routedrpc.Config{
		Handler:      handlerA,
		Provider:     providerA,
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
		CallTimeout:  2 * time.Second,
	})

	handlerB := routedrpc.NewMockHandler("player_b")
	rpcB := routedrpc.Create(&routedrpc.Config{
		Handler:      handlerB,
		Provider:     providerB,
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
		CallTimeout:  2 * time.Second,
	})

	assert.NotNil(t, rpcA)
	assert.NotNil(t, rpcB)

	ret := ""
	err = rpcA.Call("pepe", "player_b", "parameter", &ret)
	assert.Nil(t, err)
	assert.Equal(t, ret, "player_b")
	assert.Equal(t, handlerB.LastCallSender, "pepe")
	assert.Equal(t, handlerB.LastCallTarget, "player_b")
	assert.Equal(t, handlerB.LastCallMessage, "parameter")
}
