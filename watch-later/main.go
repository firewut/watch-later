package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"io/ioutil"
	"net/http"
	"os"
	"project/http_handlers"
	"project/models"
	"project/modules/config"
	"project/modules/log"
	"project/routers"
	"runtime"
	"time"
)

var (
	Log *log.Log
	err error
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	access_log := os.Stdout
	info_log := os.Stdout
	error_log := os.Stderr

	cfg, err_cfg := config.NewConfig()
	if err_cfg != nil {
		panic(err_cfg)
	}

	if len(cfg.Mode) == 0 {
		cfg.Mode = "prod"
	}

	http_handlers.Config = cfg
	models.Config = cfg
	http_handlers.SetHashes()

	s := routers.NewRouter(cfg)
	loggedRouter := handlers.LoggingHandler(os.Stdout, s)

	// init default logger - if something fails - this will save the day :D
	Log = log.InitLogggers(cfg, ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	if len(cfg.AccessLog) > 0 {
		access_log, err = os.OpenFile(cfg.AccessLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Log.Error(
				"main.AccessLog",
				err.Error(),
			)
			os.Exit(1)
		}
		loggedRouter = handlers.LoggingHandler(access_log, s)
		defer access_log.Close()
	}

	if len(cfg.ErrorLog) > 0 {
		error_log, err = os.OpenFile(cfg.ErrorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Log.Error(
				"main.ErrorLog",
				err.Error(),
			)
			os.Exit(1)
		}
		defer error_log.Close()
	}

	if len(cfg.InfoLog) > 0 {
		info_log, err = os.OpenFile(cfg.InfoLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Log.Error(
				"main.InfoLog",
				err.Error(),
			)
			os.Exit(1)
		}
		defer info_log.Close()
	}

	// init logger with correct pipes
	Log = log.InitLogggers(cfg, info_log, info_log, info_log, error_log)

	// Defer a shutdown message
	defer Log.Info(fmt.Sprintf("%s: completed shutdown.", cfg.HostName))

	http.Handle("/", loggedRouter)

	// Set loggers
	models.Log = Log
	http_handlers.Log = Log
	// --------------------------- //
	http_handlers.Router = s
	models.Router = s

	models.SyncSchemas()
	watcher := models.NewWatcher()
	watcher.Run()
	defer watcher.Quit()

	server_start_msg := fmt.Sprintf("HTTP Service Starting @%s", cfg.HttpHost)
	fmt.Fprintln(os.Stdout, server_start_msg)
	Log.Info(server_start_msg)

	srv_ptr := &http.Server{
		Addr:           cfg.HttpHost,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := srv_ptr.ListenAndServe(); err != nil {
		Log.StdOut.Error.Println(err)
		panic(err)
	}
}
