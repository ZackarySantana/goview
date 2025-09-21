package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/a-h/templ"
	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
	"github.com/zackarysantana/goview/internal/watcher"

	"github.com/zackarysantana/goview/stats"
	"github.com/zackarysantana/goview/template"
)

func Main() {
	http.Handle("/", templ.Handler(template.Index()))

	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("assets"))))

	fmt.Println("Listening on :3000 (the proxy is on :7331)")
	http.ListenAndServe(":3000", nil)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	gitIgnore := []string{".git", "node_modules"}
	patterns := make([]gitignore.Pattern, len(gitIgnore))
	for i, p := range gitIgnore {
		patterns[i] = gitignore.ParsePattern(p, nil)
	}
	ignores := gitignore.NewMatcher(patterns)

	directory := "."

	fs := os.DirFS(directory)
	module, err := stats.ParseModule(fs, directory, ignores)
	if err != nil {
		panic("Error parsing module: " + err.Error())
	}

	fmt.Printf("Module: %+v\n", module.GoMod)

	watcher, err := watcher.NewWatcher(ctx, directory, ignores)
	if err != nil {
		panic("Error creating watcher: " + err.Error())
	}

	err = watcher.Watch(ctx, func(event fsnotify.Event) error {
		fmt.Printf("File changed: %s\n", event.Name)
		var rt stats.ReloadType
		switch {
		case event.Op&fsnotify.Write == fsnotify.Write:
			rt = stats.ReloadTypeUpdate
		case event.Op&fsnotify.Remove == fsnotify.Remove:
			rt = stats.ReloadTypeRemove
		case event.Op&fsnotify.Rename == fsnotify.Rename:
			rt = stats.ReloadTypeRename
		case event.Op&fsnotify.Create == fsnotify.Create:
			rt = stats.ReloadTypeCreate
		default:
			return fmt.Errorf("unhandled file event: %s", event.Op.String())
		}
		err := module.Reload(fs, event.Name, rt)
		if err != nil {
			return fmt.Errorf("reloading module: %w", err)
		}
		return nil
	})
	if err != nil {
		panic("Error watching files: " + err.Error())
	}
}
