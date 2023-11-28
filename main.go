package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/fsnotify.v1"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path-to-your-program>")
		os.Exit(1)
	}

	programPath, _ := filepath.Abs(os.Args[1])

	// Run the initial version of the program
	restart(programPath)

	// Watch for file changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(programPath)
	if err != nil {
		log.Fatal("Error adding file to watcher:", err)
	}

	// Watch for changes in a separate goroutine
	go watchForChanges(watcher, programPath)

	// Wait for termination signal
	select {}
}

func watchForChanges(watcher *fsnotify.Watcher, programPath string) {
	for {
		select {
		case event := <-watcher.Events:
			// Ignore Chmod events
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}

			fmt.Printf("File %s changed. Restarting...\n", event.Name)

			// Delay the restart to avoid issues with editor saving
			time.Sleep(500 * time.Millisecond)

			// Restart the program on file change
			restart(programPath)

		case err := <-watcher.Errors:
			log.Println("Error watching file:", err)
		}
	}
}

func restart(programPath string) {
	// Stop the previous process if it's running
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.Println("Error killing previous process:", err)
		}
	}

	// Start a new process
	cmd = exec.Command("go", "run", programPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Println("Error starting process:", err)
		return
	}

	fmt.Println("Process restarted")
}

var cmd *exec.Cmd
