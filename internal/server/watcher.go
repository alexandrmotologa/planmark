package server

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	onChange func(filePath string)
	mu       sync.Mutex
	timers   map[string]*time.Timer
	stopChan chan struct{}
}

func NewFileWatcher(onChange func(filePath string)) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:  w,
		onChange: onChange,
		timers:   make(map[string]*time.Timer),
		stopChan: make(chan struct{}),
	}

	go fw.loop()
	return fw, nil
}

func (fw *FileWatcher) Watch(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	return fw.watcher.Add(absPath)
}

func (fw *FileWatcher) Stop() error {
	close(fw.stopChan)
	return fw.watcher.Close()
}

func (fw *FileWatcher) loop() {
	for {
		select {
		case <-fw.stopChan:
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				fw.debounce(event.Name)
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("file watcher error: %v", err)
		}
	}
}

func (fw *FileWatcher) debounce(path string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if t, exists := fw.timers[path]; exists {
		t.Stop()
	}

	fw.timers[path] = time.AfterFunc(100*time.Millisecond, func() {
		fw.mu.Lock()
		delete(fw.timers, path)
		fw.mu.Unlock()

		if fw.onChange != nil {
			fw.onChange(path)
		}
	})
}
