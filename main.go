package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/fsnotify.v1"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: golive <path-to-your-program>")
		os.Exit(1)
	}

	programPath, _ := filepath.Abs(os.Args[1])
	programArgs := os.Args[2:]

	restart(programPath, programArgs...)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(programPath)
	if err != nil {
		log.Fatal("Error adding file to watcher:", err)
	}

	go watchForChanges(watcher, func() {
		clearConsole()
		restart(programPath, programArgs...)
	})

	// Wait for termination signal
	select {}
}

var debouncer *time.Timer

func watchForChanges(watcher *fsnotify.Watcher, f func()) {
	for {
		select {
		case event := <-watcher.Events:
			// Ignore Chmod events
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}

			log.Printf("File %s changed. Restarting...\n", event.Name)

			if debouncer != nil {
				debouncer.Stop()
			}

			// Debounce restart
			debouncer = time.AfterFunc(500*time.Millisecond, f)
		case err := <-watcher.Errors:
			log.Println("Error watching file:", err)
		}
	}
}

var cmd *exec.Cmd

func restart(programPath string, args ...string) {
	running := cmd != nil && cmd.Process != nil

	// Stop the previous process if it's running
	if running {
		err := cmd.Process.Kill()
		if err != nil {
			log.Println("Error killing previous process:", err)
		}
	}

	var execArgs []string

	execArgs = append(execArgs, "run", programPath)
	execArgs = append(execArgs, args...)

	cmd = exec.Command("go", execArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Println("Error starting process:", err)
		return
	}

	if running {
		log.Println("Process restarted")
	}
}

func clearConsole() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
