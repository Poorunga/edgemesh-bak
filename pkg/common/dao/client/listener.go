package client

import (
	"fmt"

	"github.com/kubeedge/edgemesh/pkg/common/constants"
	"github.com/kubeedge/edgemesh/pkg/common/dao"
)

// Listener is only used by EdgeMesh. It stores
// the fakeIP of EdgeMesh into edge db. One fakeIP for
// one servicediscovery.
const (
	DefaultNamespace = "default"
)

type ListenerGetter interface {
	Listener() ListenInterface
}

// ListenInterface is an interface
// the key is servicediscovery name
type ListenInterface interface {
	Add(key string, value string) error
	Del(key string) error
	Get(key string) (string, error)
}

type listener struct {
}

func newListener() *listener {
	return &listener{}
}

func (ln *listener) Add(key string, value string) error {
	key = genKeyFromServiceName(key)
	// compatible for the old edgemesh's json.Marshal(content), see metaManager.processInsert()
	value = fmt.Sprintf("\"%s\"", value)
	meta := &dao.Meta{
		Key:   key,
		Type:  constants.ResourceTypeListener,
		Value: value,
	}
	err := dao.SaveMeta(meta)
	return err
}

func (ln *listener) Del(key string) error {
	key = genKeyFromServiceName(key)
	err := dao.DeleteMetaByKey(key)
	return err
}

func (ln *listener) Get(key string) (string, error) {
	key = genKeyFromServiceName(key)
	meta, err := dao.QueryMeta("key", key)
	if err != nil || len(*meta) == 0 {
		return "", err
	}
	return (*meta)[0], nil
}

func genKeyFromServiceName(key string) string {
	return 	fmt.Sprintf("%s/%s/%s", DefaultNamespace, constants.ResourceTypeListener, key)
}
