package main

import (
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

type withGoroutineID struct {
	out io.Writer
}

func (w *withGoroutineID) Write(p []byte) (int, error) {
	firstline := []byte(strings.SplitN(string(debug.Stack()), "\n", 2)[0])
	return w.out.Write(append(firstline[:len(firstline)-10], p...))
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func checkArgs() bool {
	if len(os.Args) != 2 {
		log.Fatal("Invalid Arguments: ", len(os.Args))
		return false
	}

	if !(Exists(os.Args[1])) {
		log.Fatal("File Error: ", os.Args[1])
		return false
	}
	return true
}

func main() {
	log.SetOutput(&withGoroutineID{out: os.Stderr})
	if !checkArgs() {
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
