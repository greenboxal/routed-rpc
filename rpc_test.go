package routedrpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWhoHas(t *testing.T) {
	hub := NewMockHub()

	handlerA := NewMockHandler("player_a")
	rpcA := Create(&Config{
		Handler:      handlerA,
		Provider:     hub.CreateClient(1),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	handlerB := NewMockHandler("player_b")
	rpcB := Create(&Config{
		Handler:      handlerB,
		Provider:     hub.CreateClient(2),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	assert.NotNil(t, rpcA)
	assert.NotNil(t, rpcB)

	// Test A
	who, err := rpcA.WhoHas("player_a")
	assert.Nil(t, err)
	assert.NotNil(t, who)
	assert.Equal(t, who.ID(), 1)

	who, err = rpcA.WhoHas("player_b")
	assert.Nil(t, err)
	assert.NotNil(t, who)
	assert.Equal(t, who.ID(), 2)

	// Test B
	who, err = rpcB.WhoHas("player_a")
	assert.Nil(t, err)
	assert.NotNil(t, who)
	assert.Equal(t, who.ID(), 1)

	who, err = rpcB.WhoHas("player_b")
	assert.Nil(t, err)
	assert.NotNil(t, who)
	assert.Equal(t, who.ID(), 2)
}

func TestCast(t *testing.T) {
	hub := NewMockHub()

	handlerA := NewMockHandler("player_a")
	rpcA := Create(&Config{
		Handler:      handlerA,
		Provider:     hub.CreateClient(1),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	handlerB := NewMockHandler("player_b")
	rpcB := Create(&Config{
		Handler:      handlerB,
		Provider:     hub.CreateClient(2),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	assert.NotNil(t, rpcA)
	assert.NotNil(t, rpcB)

	err := rpcA.Cast("player_b", "parameter")
	assert.Nil(t, err)

	handlerB.WaitCast()
	assert.Equal(t, handlerB.LastCastSender, nil)
	assert.Equal(t, handlerB.LastCastTarget, "player_b")
	assert.Equal(t, handlerB.LastCastMessage, "parameter")
}

func TestCall(t *testing.T) {
	hub := NewMockHub()

	handlerA := NewMockHandler("player_a")
	rpcA := Create(&Config{
		Handler:      handlerA,
		Provider:     hub.CreateClient(1),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	handlerB := NewMockHandler("player_b")
	rpcB := Create(&Config{
		Handler:      handlerB,
		Provider:     hub.CreateClient(2),
		ArpTimeout:   1 * time.Second,
		ArpCacheSize: 1000000,
	})

	assert.NotNil(t, rpcA)
	assert.NotNil(t, rpcB)

	ret := ""
	err := rpcA.Call("player_b", "parameter", &ret)
	assert.Nil(t, err)
	assert.Equal(t, ret, "player_b")
	assert.Equal(t, handlerB.LastCallSender, nil)
	assert.Equal(t, handlerB.LastCallTarget, "player_b")
	assert.Equal(t, handlerB.LastCallMessage, "parameter")
}
