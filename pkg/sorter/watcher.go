package sorter

import (
	"log"
	"sync"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"

	"github.com/fsnotify/fsnotify"
)

var timer *time.Timer

func Watcher(cfg config.Config, logger *log.Logger) {
	logger.Println("INFO: Watching for changes in", cfg.WatchDir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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
								logger.Println("INFO: File finished:", event.Name)
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
				logger.Println("ERROR:", err)
			}
		}
	}()

	err = watcher.Add(cfg.WatchDir)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{})
}
