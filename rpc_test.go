package routedrpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/stretchr/testify/assert"
)

type nullHandler struct{}

func (n *nullHandler) HasTarget(target Address) bool {
	return true
}

func (n *nullHandler) HandleCast(sender, target Address, message interface{}) error {
	return nil
}

func (n *nullHandler) HandleCall(sender, target Address, message interface{}) (interface{}, error) {
	return nil, nil
}

func setupRpc(port int, handler Handler) (*Rpc, error) {
	config := memberlist.DefaultLocalConfig()

	config.Name = fmt.Sprintf("%d", port)
	config.BindPort = port
	config.AdvertisePort = port

	return Create(&Options{
		Config:  config,
		Handler: handler,

		ArpCacheSize: 1000,
		ArpTimeout:   1 * time.Second,
	})
}

func TestRpc(t *testing.T) {
	_, err := setupRpc(10000, &nullHandler{})
	assert.Nil(t, err)

	rpc2, err := setupRpc(10001, &nullHandler{})
	assert.Nil(t, err)

	rpc2.Join([]string{"127.0.0.1:10000"})

	_, err = rpc2.WhoHas("lol123")
	assert.Nil(t, err)
}
