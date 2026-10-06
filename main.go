// Main file, starts the web interface and the main go routine that manages backups
package main

import (
	"console_backupper/engine"
	"console_backupper/model"
	"console_backupper/utils"
	"console_backupper/web"
	"console_backupper/notification"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {

	backupInterval, err := strconv.Atoi(utils.Getenv("CONSOLE_BACKUPPER_BACKUP_INTERVAL", "1"))
	if err != nil {
		log.Fatalf("Invalid backup interval: %s %v",backupInterval, err)
	}

	resetInterval, err := strconv.Atoi(utils.Getenv("CONSOLE_BACKUPPER_RESET_INTERVAL", "10"))
	if err != nil {
		log.Fatalf("Invalid backup interval: %s %v",backupInterval, err)
	}

	pruneInterval, err := strconv.Atoi(utils.Getenv("CONSOLE_BACKUPPER_PRUNE_INTERVAL", "1"))
	if err != nil {
		log.Fatalf("Invalid prune interval: %s %v",pruneInterval, err)
	}

	dataDir := utils.Getenv("CONSOLE_BACKUPPER_DATA_DIR","/var/lib/console_backupper")
	cacheDir := utils.Getenv("CONSOLE_BACKUPPER_CACHE_DIR","/var/cache/console_backupper")
	configFilePath := utils.Getenv("CONSOLE_BACKUPPER_CONFIG_FILE", "/etc/console_backupper/config.yml")
	backupsToKeep, err := strconv.Atoi(utils.Getenv("CONSOLE_BACKUPPER_BACKUPS_TO_KEEP", "5"))

	backupNotificationChannel := make(chan *model.ConsoleConfig)
	pruneNotificationChannel := make(chan *model.ConsoleConfig)
	messageNotificationChannel := make(chan notification.NotificationChannel)
	quitChannel := make(chan bool, 1)
	startBackupsTimer := time.NewTicker(time.Duration(backupInterval) * time.Minute)
	resetBackupsTimer := time.NewTicker(time.Duration(resetInterval) * time.Minute)

	go engine.Engine(
		configFilePath,
		backupNotificationChannel,
		startBackupsTimer,
		resetBackupsTimer,
		pruneNotificationChannel,
		messageNotificationChannel,
		quitChannel,
		dataDir,
		cacheDir,
		backupsToKeep,
	)


	http.HandleFunc("/", web.HomeHandler)
	http.HandleFunc("/consoles/{console}", web.ConsoleHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir(utils.Getenv("CONSOLE_BACKUPPER_DATA_DIR", "/var/lib/console_backupper")))))

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	http.Handle(utils.Getenv("CONSOLE_BACKUPPER_METRICS_PATH", "/metrics"), promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
