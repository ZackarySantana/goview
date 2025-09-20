package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
)

type Watcher struct {
	w       *fsnotify.Watcher
	ignores gitignore.Matcher

	mu      sync.Mutex
	watched map[string]struct{}
}

// NewWatcher creates a recursive watcher rooted at rootDir.
func NewWatcher(ctx context.Context, rootDir string, ignores []string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	var ps []gitignore.Pattern
	for _, p := range ignores {
		ps = append(ps, gitignore.ParsePattern(p, nil))
	}
	wd := &Watcher{
		w:       w,
		watched: make(map[string]struct{}),
		ignores: gitignore.NewMatcher(ps),
	}

	// Add all existing dirs
	if err := wd.addDirRecursive(rootDir); err != nil {
		_ = w.Close()
		return nil, err
	}

	// Stop when ctx is done
	go func() {
		<-ctx.Done()
		_ = wd.w.Close()
	}()

	return wd, nil
}

func (wd *Watcher) addDirRecursive(root string) error {
	return filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if wd.ignores.Match(strings.Split(p, string(os.PathSeparator)), true) {
			return filepath.SkipDir
		}
		return wd.addWatch(p)
	})
}

func (wd *Watcher) addWatch(dir string) error {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	if _, ok := wd.watched[dir]; ok {
		return nil
	}
	if err := wd.w.Add(dir); err != nil {
		return err
	}
	wd.watched[dir] = struct{}{}
	return nil
}

func (wd *Watcher) removeWatch(dir string) {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	if _, ok := wd.watched[dir]; ok {
		_ = wd.w.Remove(dir)
		delete(wd.watched, dir)
	}
}

// Events runs the event loop and emits debounced paths.
func (wd *Watcher) Events(ctx context.Context) <-chan fsnotify.Event {
	out := make(chan fsnotify.Event, 64)

	// optional debounce
	const debounce = 50 * time.Millisecond
	type key struct {
		Name string
		Op   fsnotify.Op
	}
	pending := make(map[key]fsnotify.Event)
	var timer *time.Timer
	flush := func() {
		for _, ev := range pending {
			out <- ev
		}
		pending = make(map[key]fsnotify.Event)
		if timer != nil {
			timer.Stop()
			timer = nil
		}
	}

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				flush()
				return

			case err, ok := <-wd.w.Errors:
				if !ok {
					flush()
					return
				}
				log.Printf("fsnotify error: %v", err)

			case ev, ok := <-wd.w.Events:
				if !ok {
					flush()
					return
				}

				// Handle dynamic subdirs
				switch {
				case ev.Op&fsnotify.Create == fsnotify.Create:
					// If a directory was created, start watching it (and its children if already populated).
					info, err := os.Lstat(ev.Name)
					if err == nil && info.IsDir() {
						_ = wd.addDirRecursive(ev.Name)
					}
				case ev.Op&(fsnotify.Remove|fsnotify.Rename) != 0:
					// If a watched dir was removed/renamed, stop watching it.
					wd.removeWatch(ev.Name)
				}

				// Debounce coalesced bursts
				k := key{ev.Name, ev.Op}
				pending[k] = ev
				if timer == nil {
					timer = time.AfterFunc(debounce, flush)
				} else {
					timer.Reset(debounce)
				}
			}
		}
	}()

	return out
}

func Watch(ctx context.Context, root string, ignores []string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wd, err := NewWatcher(ctx, root, ignores)
	if err != nil {
		return err
	}

	fmt.Println("Watching:", root)
	for ev := range wd.Events(ctx) {
		fmt.Printf("event: %-10s %s\n", opString(ev.Op), ev.Name)
	}

	return nil
}

func opString(op fsnotify.Op) string {
	var s []string
	if op&fsnotify.Create != 0 {
		s = append(s, "CREATE")
	}
	if op&fsnotify.Write != 0 {
		s = append(s, "WRITE")
	}
	if op&fsnotify.Remove != 0 {
		s = append(s, "REMOVE")
	}
	if op&fsnotify.Rename != 0 {
		s = append(s, "RENAME")
	}
	if op&fsnotify.Chmod != 0 {
		s = append(s, "CHMOD")
	}
	if len(s) == 0 {
		return op.String()
	}
	return strings.Join(s, "|")
}
