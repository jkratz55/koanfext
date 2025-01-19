package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/v2"
)

var _ koanf.Provider = (*File)(nil)

// File is an implementation of koanf.Provider that reads/loads configuration from
// a file. File is also capable of watching for changes to the configured file
// and invoking a callback when changes are detected.
type File struct {
	path    string
	watcher *fsnotify.Watcher
	watched atomic.Uint32
}

// Provider initializes a new File.
func Provider(path string) *File {
	return &File{
		path:    path,
		watcher: nil,
		watched: atomic.Uint32{},
	}
}

// ReadBytes reads the file and returns the raw bytes.
func (f *File) ReadBytes() ([]byte, error) {
	return os.ReadFile(f.path)
}

// Read is not implemented for File and will always return an error
func (f *File) Read() (map[string]interface{}, error) {
	return nil, fmt.Errorf("%T does not support Read()", f)
}

// Watch monitors the file for changes and invokes the provided callback when
// changes are detected.
//
// Watch may only be invoked once per instance of File and providing a nil callback
// will result in a panic.
func (f *File) Watch(cb func(event interface{}, err error)) error {
	activated := f.watched.CompareAndSwap(0, 1)
	if !activated {
		return fmt.Errorf("%T.Watch may only be invoked once", f)
	}

	configFile := filepath.Clean(f.path)
	configDir, _ := filepath.Split(configFile)
	realConfigFile, err := filepath.EvalSymlinks(f.path)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					cb(nil, fmt.Errorf("fsnotify watcher closed"))
					return
				}

				currentConfigFile, err := filepath.EvalSymlinks(f.path)
				if err != nil {
					cb(nil, err)
					continue
				}

				// If the filename matches the file being monitored and the file
				// was either created than notify the file has changed so the caller
				// can decided if they want to refresh the configuration.
				if (filepath.Clean(event.Name) == configFile &&
					(event.Has(fsnotify.Write) || event.Has(fsnotify.Create))) ||
					(currentConfigFile != "" && currentConfigFile != realConfigFile) {
					cb(event, nil)
				} else if filepath.Clean(event.Name) == configFile && event.Has(fsnotify.Remove) {
					cb(nil, fmt.Errorf("file %s was removed", event.Name))
					return
				}
			case err, ok := <-watcher.Errors:
				if ok {
					cb(nil, err)
				}
			}
		}
	}()

	f.watcher = watcher
	return watcher.Add(configDir)
}

// Close gracefully closes and releases any resources File is using.
func (f *File) Close() error {
	if f.watcher != nil {
		return f.watcher.Close()
	}
	// If Watch was never called this is a no-op
	return nil
}
