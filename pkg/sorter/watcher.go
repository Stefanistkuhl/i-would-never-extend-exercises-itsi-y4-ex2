package sorter

import (
	"sync"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

func Watcher(cfg config.Config, logger logger.Logger) {
	logger.Info("Watching for changes", "directory", cfg.WatchDir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal("Failed to create file watcher", "error", err)
	}

	fileTimers := make(map[string]*time.Timer)
	mu := &sync.Mutex{}
	quietTimeout := 3 * time.Second

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					mu.Lock()
					if timer, exists := fileTimers[event.Name]; exists {
						timer.Stop()
					}

					fileTimers[event.Name] = time.AfterFunc(
						quietTimeout,
						func() {
							if cfg.LogLevel == "info" {
								logger.Info("File finished writing", "file", event.Name)
							}
							mu.Lock()
							delete(fileTimers, event.Name)
							mu.Unlock()

							go processFile(cfg, event.Name, logger)
						},
					)
					mu.Unlock()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("File watcher error", "error", err)
			}
		}
	}()

	err = watcher.Add(cfg.WatchDir)
	if err != nil {
		logger.Fatal("Failed to add watch directory", "path", cfg.WatchDir, "error", err)
	}

	<-make(chan struct{})
}
