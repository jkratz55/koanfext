package koanfext

type Option func(*KoanfWrapper)

func OnConfigChanged(fn func()) Option {
	return func(k *KoanfWrapper) {
		if k.onConfigChanged != nil {
			k.onConfigChanged = fn
		}
	}
}

func OnError(fn func(err error)) Option {
	return func(k *KoanfWrapper) {
		if k.onReloadError != nil {
			k.onReloadError = fn
		}
	}
}

func Sources(sources ...Source) Option {
	return func(k *KoanfWrapper) {
		k.sources = sources
	}
}
