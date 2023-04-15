package filewatcher

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"gopkg.in/fsnotify.v1"
	"sync"
)

type FileWatcher interface {
	GetContent() interface{}
	Run(ctx context.Context) error // Blocking function
}

type Parser func(fileName string) (interface{}, error)

var _ FileWatcher = &fileWatcher{}

func New(fileName string, parser Parser, logger logr.Logger) (FileWatcher, error) {
	fw := &fileWatcher{
		fileName: fileName,
		logger:   logger,
		parser:   parser,
	}
	var err error
	fw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	// Make a first data upload. This to have a coherent dataset even before starting as daemon and to check initial user fila coherency
	fw.content, err = fw.parser(fw.fileName)
	if err != nil {
		return nil, err
	}
	return fw, nil
}

type fileWatcher struct {
	sync.Mutex
	logger logr.Logger

	fileName string
	parser   Parser
	watcher  *fsnotify.Watcher
	content  interface{}
}

func (fw *fileWatcher) GetContentSync() (interface{}, error) {
	fw.Lock()
	defer fw.Unlock()
	return fw.parser(fw.fileName)
}

func (fw *fileWatcher) GetContent() interface{} {
	fw.Lock()
	defer fw.Unlock()
	return fw.content
}

func (fw *fileWatcher) Run(ctx context.Context) error {
	// Initial Reading and parsing file
	content, err := fw.parser(fw.fileName)
	if err != nil {
		return err
	}
	fw.content = content
	err = fw.watcher.Add(fw.fileName)
	if err != nil {
		return err
	}
	go fw.watch()

	fw.logger.Info("Starting fileWatcher")

	// Block until the stop channel is closed.
	<-ctx.Done()

	return fw.watcher.Close()
}

// Watch reads events from the watcher's channel and reacts to changes.
func (fw *fileWatcher) watch() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				// Channel is closed
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				// Channel is closed
				return
			}
			fw.logger.Error(err, fmt.Sprintf("FileWatcher(%s) error", fw.fileName))
		}
	}

}

func (fw *fileWatcher) handleEvent(event fsnotify.Event) {
	// Only care about events which may modify the contents of the file.
	if !(isWrite(event) || isRemove(event) || isCreate(event)) {
		return
	}

	fw.logger.V(1).Info("certificate event", "event", event)

	// If the file was removed, re-add the watch.
	if isRemove(event) {
		if err := fw.watcher.Add(event.Name); err != nil {
			fw.logger.Error(err, "error re-watching file")
		}
	}

	if err := fw.readFile(); err != nil {
		fw.logger.Error(err, "error re-reading file")
	}

}

func (fw *fileWatcher) readFile() error {
	fw.logger.Info("Reload file", "file", fw.fileName)
	content, err := fw.parser(fw.fileName)
	if err != nil {
		return err
	}
	fw.Lock()
	defer fw.Unlock()
	fw.content = content
	return nil
}

func isWrite(event fsnotify.Event) bool {
	return event.Op&fsnotify.Write == fsnotify.Write
}

func isCreate(event fsnotify.Event) bool {
	return event.Op&fsnotify.Create == fsnotify.Create
}

func isRemove(event fsnotify.Event) bool {
	return event.Op&fsnotify.Remove == fsnotify.Remove
}
