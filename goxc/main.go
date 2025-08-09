// goxc/main.go
package main

import (
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
)

type CLI struct {
	Path  string `kong:"arg,optional,default='.',help='Path to compile or watch (file or directory).'"`
	Watch bool   `kong:"help='Enable watch mode to automatically recompile .gox files.'"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

// Run is the main entry point for the application logic.
func (cli *CLI) Run() error {
	if cli.Watch {
		log.Printf("Starting watcher for path: %s", cli.Path)
		return runWatcher(cli.Path)
	}

	log.Printf("Running single compilation pass for path: %s", cli.Path)
	return runSinglePass(cli.Path)
}

// runSinglePass compiles all .gox files in a directory/file and then exits.
func runSinglePass(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if info.IsDir() {
		// It's a directory, so walk it.
		return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(p, ".gox") {
				if err := processFile(p); err != nil {
					log.Printf("ERROR: Failed to process %s: %v", p, err)
				}
			}
			return nil
		})
	}

	// It's a single file.
	if strings.HasSuffix(path, ".gox") {
		return processFile(path)
	}
	return nil
}

// runWatcher starts the file watcher and recompiles files on change.
func runWatcher(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.HasSuffix(event.Name, ".gox") {
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
						if err := processFile(event.Name); err != nil {
							log.Printf("ERROR: Failed to process %s: %v", event.Name, err)
						}
					}
					if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						deleteGeneratedFile(event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("WATCHER ERROR: %v", err)
			}
		}
	}()

	log.Println("Performing initial compilation pass...")
	if err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := watcher.Add(p); err != nil {
				log.Printf("WARN: Failed to add directory %s to watcher: %v", p, err)
			}
		} else if strings.HasSuffix(p, ".gox") {
			if err := processFile(p); err != nil {
				log.Printf("ERROR: Failed to process %s: %v", p, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error during initial scan: %w", err)
	}

	log.Println("Watcher is running. Press Ctrl+C to exit.")
	<-make(chan struct{}) // Block forever

	return nil
}

// processFile reads, transforms, and writes a single file.
func processFile(path string) error {
	log.Printf("Processing %s...", path)
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	parser := NewParser(source)
	transformedSource, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse gox source: %w", err)
	}

	formattedSource, err := format.Source(transformedSource)
	if err != nil {
		return fmt.Errorf("could not format generated go code: %w", err)
	}

	outputPath := strings.TrimSuffix(path, ".gox") + "__gox.go"
	if err := os.WriteFile(outputPath, formattedSource, 0644); err != nil {
		return fmt.Errorf("could not write output file: %w", err)
	}

	log.Printf("Successfully generated %s", outputPath)
	return nil
}

// deleteGeneratedFile removes the corresponding .gox.go file.
func deleteGeneratedFile(path string) {
	outputPath := strings.TrimSuffix(path, ".gox") + ".gox.go"
	if _, err := os.Stat(outputPath); err == nil {
		log.Printf("Removing stale file: %s", outputPath)
		os.Remove(outputPath)
	}
}
