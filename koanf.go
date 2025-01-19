package koanfext

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

// Watchable is a type capable of watching for configuration changes and notifying
// changes through a callback.
//
// Implementations of Watchable MUST be nonblocking, or it will cause KoanfWrapper
// to either deadlock or react slowly to changes.
type Watchable interface {
	Watch(cb func(event interface{}, err error)) error
}

// Source represents the source of a configuration. The source contains the
// Provider to load read the configuration, and the Parser to decode it.
type Source struct {
	Provider koanf.Provider
	Parser   koanf.Parser
}

// KoanfWrapper is a wrapper around Koanf that abstracts away loading the
// configuration and handling changes/reloads when a Watchable Provider is
// changed. All Providers that implement the Watchable interface are watched
// automatically, and the configuration is reloaded when a change is detected.
//
// Note by default KoanfWrapper has no sources. KoanfWrapper will in nearly
// all cases we called by passing the Sources Option which provides the sources
// to be loaded in the order they are provided.
type KoanfWrapper struct {
	*koanf.Koanf
	sources         []Source
	mu              sync.Mutex
	onConfigChanged func()
	onReloadError   func(err error)
}

// NewKoanfWrapper initializes a new KoanfWrapper instance. The behavior and
// configuration of KoanfWrapper can be customized by passing various Option
// types.
func NewKoanfWrapper(opts ...Option) (*KoanfWrapper, error) {
	wrapper := &KoanfWrapper{
		Koanf:           koanf.New("."),
		sources:         make([]Source, 0),
		mu:              sync.Mutex{},
		onReloadError:   func(err error) {},
		onConfigChanged: func() {},
	}

	for _, opt := range opts {
		opt(wrapper)
	}

	if err := wrapper.load(); err != nil {
		return nil, err
	}

	if err := wrapper.setupWatchers(); err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (k *KoanfWrapper) load() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	conf := koanf.New(".")
	for _, source := range k.sources {
		if err := conf.Load(source.Provider, source.Parser); err != nil {
			return err
		}
	}

	k.Koanf = conf
	return nil
}

func (k *KoanfWrapper) setupWatchers() error {
	for _, source := range k.sources {
		if watchable, ok := source.Provider.(Watchable); ok {
			err := watchable.Watch(func(event interface{}, err error) {
				if err != nil {
					k.onReloadError(err)
					return
				}
				if err := k.load(); err != nil {
					k.onReloadError(err)
				}
				k.onConfigChanged()
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
