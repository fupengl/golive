package main

import (
	"fmt"
	"go/build"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"gopkg.in/fsnotify.v1"
)

var cwd, _ = os.Getwd()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: golive <path-to-your-program>")
		os.Exit(1)
	}

	programPath, _ := filepath.Abs(os.Args[1])
	programArgs := os.Args[2:]

	dirsToWatch, err := findDependencyDirs(programPath)
	if err != nil {
		log.Fatal("Error finding dependency directories:", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error creating watcher:", err)
	}
	defer watcher.Close()

	// Watch all directory
	for _, path := range dirsToWatch {
		err := watcher.Add(path)
		if err != nil {
			log.Println("Error adding path to watcher:", err)
		}
		err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return watcher.Add(path)
		})
		if err != nil {
			log.Println("Error adding path to watcher:", err)
		}
		relPath, _ := filepath.Rel(cwd, path)
		if cwd == path {
			relPath = filepath.Base(cwd)
		}
		log.Printf("Watching directory \033[36m%s\u001B[0m...\n", relPath)
	}

	go watchForChanges(watcher, func() {
		clearConsole()

		restart(programPath, programArgs...)
	})

	restart(programPath, programArgs...)

	// Wait for termination signal
	select {}
}

func findDependencyDirs(programPath string) ([]string, error) {
	goModPath, err := findGoModPath(programPath)
	if err != nil {
		return nil, err
	}
	goModDir := filepath.Dir(goModPath)

	// Read the go.mod file to extract dependencies
	goModData, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	modFile, err := modfile.Parse("go.mod", goModData, nil)
	if err != nil {
		return nil, err
	}

	// Extract module dependencies
	var dirsToWatch []string

	// The mod directory needs to be added
	dirsToWatch = append(dirsToWatch, goModDir)

	for _, rep := range modFile.Replace {

		// Convert the replacement path to the corresponding directory path
		depPath := filepath.Join(goModDir, rep.New.Path)
		dirsToWatch = append(dirsToWatch, depPath)
	}

	return dirsToWatch, nil
}

func findGoModPath(programPath string) (string, error) {
	// Use go/packages to find the go.mod file
	cfg := &packages.Config{Mode: packages.NeedName}
	pkgs, err := packages.Load(cfg, programPath)
	if err != nil {
		return "", err
	}

	if len(pkgs) == 0 {
		return "", fmt.Errorf("no packages found in %s", programPath)
	}

	// Check if the package has a go.mod file
	goModPath := filepath.Join(pkgs[0].PkgPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return goModPath, nil
	}

	// If not, check parent directories
	dir := filepath.Dir(programPath)
	for dir != "" && dir != "." && dir != string(filepath.Separator) {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath, nil
		}

		// Move up to the parent directory
		dir = filepath.Dir(dir)
	}

	// Fallback to GOPATH if not found
	if gopath := build.Default.GOPATH; gopath != "" {
		// Split GOPATH into multiple paths
		gopathPaths := filepath.SplitList(gopath)
		for _, gopathPath := range gopathPaths {
			goModPath := filepath.Join(gopathPath, "src", pkgs[0].PkgPath, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				return goModPath, nil
			}
		}
	}

	return "", fmt.Errorf("go.mod not found in %s or its parent directories", programPath)
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

			if debouncer != nil {
				debouncer.Stop()
			}

			// Debounce restart
			debouncer = time.AfterFunc(500*time.Millisecond, func() {
				relPath, _ := filepath.Rel(cwd, event.Name)
				log.Printf("File \033[36m%s\033[0m changed. Restarting...\n", relPath)

				f()
			})
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

	// Output exec cmd
	if !running {
		log.Println("Process starting...")
		log.Printf("\033[33m > %s\033[0m", cmd.String())
	}

	err := cmd.Start()
	if err != nil {
		log.Println("Error starting process:", err)
		return
	}

	if running {
		log.Println("Process restarted.")
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
