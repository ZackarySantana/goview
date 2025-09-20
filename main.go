package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

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

	gitIgnore := []string{".git", "node_modules"}
	patterns := make([]gitignore.Pattern, len(gitIgnore))
	for i, p := range gitIgnore {
		patterns[i] = gitignore.ParsePattern(p, nil)
	}
	ignores := gitignore.NewMatcher(patterns)

	fs := os.DirFS(".")
	module, err := stats.ParseModule(fs, ".", ignores)
	if err != nil {
		fmt.Println("Error parsing module:", err)
		return
	}

	fmt.Printf("Module: %+v\n", module.GoMod)

	watcher, err := watcher.NewWatcher(context.Background(), ".", ignores)
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}

	err = watcher.Watch(context.Background(), ".", func(event fsnotify.Event) {
		fmt.Printf("File changed: %s\n", event.Name)
		module.Reload(fs, event.Name)
	})
	if err != nil {
		fmt.Println("Error watching files:", err)
		return
	}
}
