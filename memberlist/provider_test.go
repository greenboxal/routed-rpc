package memberlist

import (
	"fmt"
	"testing"
	"time"

	"github.com/greenboxal/routed-rpc"
	mb "github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
)

func setupProvider(port int) (*Memberlist, error) {
	cfg := mb.DefaultLocalConfig()

	cfg.Name = fmt.Sprintf("test_%d", port)
	cfg.BindPort = port
	cfg.AdvertisePort = port

	return Create(cfg)
}

func TestProvider(t *testing.T) {
	providerA, err := setupProvider(10000)
	assert.Nil(t, err)
	assert.NotNil(t, providerA)

	providerB, err := setupProvider(10001)
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
	err = rpcA.Call("player_b", "parameter", &ret)
	assert.Nil(t, err)
	assert.Equal(t, ret, "player_b")
	assert.Equal(t, handlerB.LastCallSender, nil)
	assert.Equal(t, handlerB.LastCallTarget, "player_b")
	assert.Equal(t, handlerB.LastCallMessage, "parameter")
}
