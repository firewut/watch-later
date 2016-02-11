package models

import (
	"fmt"
	"golang.org/x/oauth2"
	"project/modules/messages"
	"runtime"
	"strings"
	"sync"
	"time"
)

type task struct {
	Name           string        `json:"name"`
	Tick           time.Duration `json:"tick"`
	force_tick     chan bool
	quit           chan bool
	quit_completed chan bool
	function       func(*task) error
	watcher        *Watcher

	wg *sync.WaitGroup
}

// Trigger a task's function
func (t *task) trigger() {
	if err := t.function(t); err != nil {
		Log.Error(
			fmt.Sprintf("watcher task %s", t.Name),
			err.Error(),
		)
	}
}

// Set up a task
func (t *task) up() {
	// If no tick discovered
	if t.Tick <= 0 {
		// Tick one per five years :D
		t.Tick = time.Hour * 43800
		// Trigger task forced :D
		go func(task_passed *task) {
			task_passed.trigger()
		}(t)
	}
	timer := time.NewTicker(t.Tick)
	for {
		select {
		case <-timer.C:
			t.trigger()
		case <-t.force_tick:
			t.trigger()
		case <-t.quit:
			Log.Info(fmt.Sprintf("stopped task: %s", t.Name))
			t.quit_completed <- true

			return
		}
	}
}

// list of all tasks
var tasks = map[string]*task{
	"watch_later": &task{
		Tick:           CHECK_WATCH_LATER_PERIOD,
		quit:           make(chan bool, 1),
		quit_completed: make(chan bool, 1),
		force_tick:     make(chan bool, 1),
		function: func(task_passed *task) (err error) {
			var wg sync.WaitGroup
			concurrency := make(chan struct{}, runtime.NumCPU())
			defer close(concurrency)

			// Get config
			cfg := GetYoutubeConfig()

			// Get all tokens from `deform.io`
			page := int64(1)
			for {
				tokens_from_page, next_page, _ := GetActiveTokens(page)
				for _, token := range tokens_from_page {
					wg.Add(1)
					go func(
						wg_passed *sync.WaitGroup,
						concurrency_channel chan struct{},
						token_passed *Token,
					) {
						defer wg_passed.Done()
						concurrency_channel <- struct{}{}

						start_time := time.Now().UTC()

						// Call process - will cause auto-refresh if necessary
						client := cfg.Client(oauth2.NoContext, token_passed.OAuth2Token)

						old_token := *token_passed

						// If token was auto-refreshed - delete old token and save new
						if token_passed.OAuth2Token.AccessToken != old_token.OAuth2Token.AccessToken ||
							token_passed.OAuth2Token.RefreshToken != old_token.OAuth2Token.RefreshToken {
							Log.Warning(
								fmt.Sprintf(
									"Token refreshed %v => %v. %v => %v",
									old_token.OAuth2Token.AccessToken,
									token_passed.OAuth2Token.AccessToken,
									old_token.OAuth2Token.RefreshToken,
									token_passed.OAuth2Token.RefreshToken,
								),
							)
							token_passed.Save()
						}

						_, err_processing := ProcessWatchLater(client, token_passed)
						if err_processing != nil {
							err_message := err_processing.Error()
							if strings.Contains(err_message, "Token has been revoked.") ||
								strings.Contains(err_message, "Invalid Credentials") {

								token_passed.Delete()
								Log.Warning("watch_later task: removing revoked token ", token_passed.Profile.Id)
							} else {
								Log.Error("watch_later task: ", token_passed.Profile.Id, err_processing)
							}
						}

						Log.Info(fmt.Sprintf("processing a %s took %v", token_passed.Profile.Id, time.Since(start_time)))
						go token_passed.Patch(map[string]interface{}{
							"latest_operation": time.Now().UTC(),
						})

					}(&wg, concurrency, token)
				}
				if next_page == 0 {
					break
				} else {
					page = next_page
				}
			}

			// Read from a channel
			go func() {
				for _ = range concurrency {
				}
			}()

			wg.Wait()

			return nil
		},
	},
}

// This will periodically run tasks
type Watcher struct {
	tasks map[string]*task
}

func NewWatcher() (watcher *Watcher) {
	watcher = &Watcher{}

	watcher.tasks = make(map[string]*task)
	for name, task_ptr := range tasks {
		task_ptr.Name = name
		task_ptr.watcher = watcher
		watcher.tasks[name] = task_ptr
	}

	return
}

// Run task - wg will be Done() when a task will finish
func (w *Watcher) RunTask(name string, wg *sync.WaitGroup) (err error) {
	err = messages.ErrTaskNotFound
	for _, task := range w.tasks {
		if task.Name == name {
			err = nil
			task.wg = wg
			task.force_tick <- true
		}
	}
	return
}

// Quit watcher. All tasks will be notified.
func (w *Watcher) Quit() {
	var wg sync.WaitGroup
	wg.Add(len(w.tasks))
	for _, task_instance := range w.tasks {
		go func(wg_passed *sync.WaitGroup, task_passed *task) {
			task_passed.quit <- true
			<-task_passed.quit_completed
			wg_passed.Done()
		}(&wg, task_instance)
	}
	wg.Wait()
}

// Run all tasks
func (w *Watcher) Run() {
	for _, task := range w.tasks {
		go task.up()
	}
}
