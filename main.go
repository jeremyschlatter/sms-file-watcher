package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfreiberg/gotwilio"
	"gopkg.in/fsnotify.v1"
)

var (
	accountSID, authToken, from, to string
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("Must pass exactly one argument: The directory to watch.")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	watch(watcher.Events, watcher.Errors)
}

func watch(events chan fsnotify.Event, errors chan error) {
	var (
		timer  <-chan time.Time
		count  int
		twilio = gotwilio.NewTwilioClient(accountSID, authToken)
	)
	for {
		select {
		case err := <-errors:
			log.Println(err)
		case evt := <-events:
			name := filepath.Base(evt.Name)
			if evt.Op == fsnotify.Create &&
				strings.HasPrefix(name, "scan") &&
				strings.HasSuffix(name, ".jpg") {
				timer = time.After(5 * time.Second)
				count++
			}
		// I don't want to send an update for every file in a batch update,
		// so the timer is present to make sure we wait five seconds after
		// the last file creation before sending the SMS.
		case <-timer:
			timer = nil
			pages := "pages"
			if count == 1 {
				pages = "page"
			}
			_, exc, err := twilio.SendSMS(from, to, fmt.Sprintf("Scanned %d %s", count, pages), "", "")
			if exc != nil {
				log.Println(exc)
			}
			if err != nil {
				log.Println(err)
			}
			count = 0
		}
	}
}
