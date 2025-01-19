package kubernetes

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/knadh/koanf/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var _ koanf.Provider = (*ConfigMapFile)(nil)

// ConfigMapFile is an implementation of koanf.Provider that reads/loads a config
// file stored inside a configmap as a key. ConfigMapFile is capable of watching
// the ConfigMap in Kubernetes for changes and notifying via a callback.
type ConfigMapFile struct {
	client    *kubernetes.Clientset
	name      string
	namespace string
	key       string // key would be the filename in the ConfigMap
	watched   atomic.Uint32
	stopCh    chan struct{}
}

// ConfigMapFileProvider creates and returns a ConfigMapFile instance to read and
// watch a ConfigMap in Kubernetes.
func ConfigMapFileProvider(k8sClient *kubernetes.Clientset, cmName, cmNamespace, key string) *ConfigMapFile {
	if k8sClient == nil {
		panic("k8sClient cannot be nil")
	}
	return &ConfigMapFile{
		client:    k8sClient,
		name:      cmName,
		namespace: cmNamespace,
		key:       key,
		watched:   atomic.Uint32{},
		stopCh:    make(chan struct{}, 1),
	}
}

// ReadBytes reads and returns the contents of a configuration file stored in
// a Kubernetes ConfigMap.
func (c *ConfigMapFile) ReadBytes() ([]byte, error) {
	cm, err := c.client.CoreV1().ConfigMaps(c.namespace).Get(context.Background(), c.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data, ok := cm.Data[c.key]
	if !ok {
		return nil, fmt.Errorf("key %s not found in configmap %s/%s", c.key, c.namespace, c.name)
	}
	return []byte(data), nil
}

// Read is not supported by ConfigMapFile and will always return an error.
func (c *ConfigMapFile) Read() (map[string]interface{}, error) {
	return nil, fmt.Errorf("%T does not support Read()", c)
}

// Watch sets up a listener to monitor changes in the ConfigMap and invokes the
// callback upon add or update events. It ensures the method can only be invoked once
// and blocks until the cache syncs successfully. Returns an error if the watch
// activation fails or cache synchronization times out.
func (c *ConfigMapFile) Watch(cb func(event interface{}, err error)) error {
	activated := c.watched.CompareAndSwap(0, 1)
	if !activated {
		return fmt.Errorf("%T.Watch may only be invoked once", c)
	}

	listWatch := cache.NewListWatchFromClient(c.client.CoreV1().RESTClient(),
		"configmaps", c.namespace, fields.OneTermEqualSelector("metadata.name", c.name))

	informer := cache.NewSharedInformer(listWatch, &corev1.ConfigMap{}, 0)

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			cb(cm, nil)
		},
		UpdateFunc: func(old, new interface{}) {
			cm := new.(*corev1.ConfigMap)
			cb(cm, nil)
		},
	})
	if err != nil {
		return err
	}

	go informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	return nil
}

// Close gracefully closes a ConfigMap watch if Watch was called. Otherwise, it
// is a no-op.
func (c *ConfigMapFile) Close() {
	if c.watched.Load() == 1 {
		close(c.stopCh)
	}
}
