// SPDX-FileContributor: slowerloris <taylor@teukka.tech>
//
// SPDX-License-Identifier: AGPL-3.0-or-later
package files

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/model"
)

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// Debounce so we don't spam the index as write events can file multiple times before closing a file after editing
const debounceTime = 200 * time.Millisecond

// HasPathPrefix reports whether filePath equals dirPath or is contained within it,
// using the platform's path separator.
func HasPathPrefix(filePath, dirPath string) bool {
	if filePath == dirPath {
		return true
	}
	return strings.HasPrefix(filePath, dirPath+string(filepath.Separator))
}

// PathToFileURL converts an absolute filesystem path into a file:// URL.
// On Windows, paths like C:\foo\bar become file:///C:/foo/bar.
func PathToFileURL(absPath string) string {
	p := filepath.ToSlash(absPath)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

// FileURLToPath extracts an OS-native filesystem path from a file:// URL.
// On Windows, it strips the leading slash that precedes a drive letter
// (e.g. file:///C:/foo → C:\foo). A non-file URL is returned unchanged.
func FileURLToPath(fileURL string) string {
	p, ok := strings.CutPrefix(fileURL, "file://")
	if !ok {
		return fileURL
	}
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}

// FindMatchingDir returns the Directory config whose expanded path contains filePath, or nil.
func FindMatchingDir(dirs []*config.Directory, filePath string) *config.Directory {
	for i := range dirs {
		if dirs[i] == nil {
			continue
		}
		dirPath := filepath.Clean(ExpandHome(dirs[i].Path))
		if HasPathPrefix(filePath, dirPath) {
			return dirs[i]
		}
	}
	return nil
}

// DirectoryMatchesPath reports whether filePath would be discovered while
// walking dir with its current filters. In addition to the file name filters,
// it checks every parent directory below the configured root because excluded
// and hidden directories are pruned during directory walks.
func DirectoryMatchesPath(dir *config.Directory, filePath string) bool {
	if dir == nil {
		return false
	}
	root := filepath.Clean(ExpandHome(dir.Path))
	filePath = filepath.Clean(filePath)
	if !HasPathPrefix(filePath, root) || filePath == root {
		return false
	}
	rel, err := filepath.Rel(root, filePath)
	if err != nil {
		return false
	}
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		if shouldSkipDir(part, dir.Excludes, dir.IncludeHidden) {
			return false
		}
	}
	return dir.IsMatching(filePath)
}

// FindDirUser finds the directory config matching a file path and resolves its user to a user ID.
// Returns 0 for global directories (no user set). Returns an error if the username can't be resolved.
func FindDirUser(dirs []*config.Directory, filePath string) (uint, error) {
	dir := FindMatchingDir(dirs, filePath)
	if dir == nil {
		return 0, nil
	}
	if dir.User == "" {
		return 0, nil
	}
	u, err := model.GetUser(dir.User)
	if err != nil {
		return 0, fmt.Errorf("user %q not found", dir.User)
	}
	return u.ID, nil
}

// skipDirs lists directory names that are skipped by default during watching.
// These are well-known dependency/cache directories whose names are unambiguous
// and can contain tens of thousands of entries, easily exhausting OS watch limits.
// Hidden directories (starting with ".") are always skipped separately.
// Users can exclude additional directories via the per-directory excludes config.
var skipDirs = map[string]struct{}{
	"node_modules":     {},
	"bower_components": {},
	"jspm_packages":    {},
	"__pycache__":      {},
	"__pypackages__":   {},
}

// shouldSkipDir reports whether a directory should be excluded from watching.
// It skips hidden directories, well-known dependency/cache directories, and
// directories matching any exclude pattern from the config.
func shouldSkipDir(name string, excludes []string, includeHidden bool) bool {
	if !includeHidden {
		if strings.HasPrefix(name, ".") {
			return true
		}
		if _, ok := skipDirs[name]; ok {
			return true
		}
	}
	for _, pattern := range excludes {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

// ShouldSkipDir is the exported form of shouldSkipDir for use by the indexer.
func ShouldSkipDir(name string, excludes []string, includeHidden bool) bool {
	return shouldSkipDir(name, excludes, includeHidden)
}

// walkAndWatch registers all subdirectories of each configured directory with
// the fsnotify watcher, skipping hidden dirs and user-configured excludes.
func walkAndWatch(watcher *fsnotify.Watcher, dirs []*config.Directory) {
	for _, dir := range dirs {
		expanded := ExpandHome(dir.Path)
		if err := watcher.Add(expanded); err != nil {
			log.Error().Err(err).Str("path", expanded).Msg("Failed to add path to file watcher")
		}
		excludes := dir.Excludes
		_ = filepath.WalkDir(expanded, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("Error walking directory")
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			if path != expanded && shouldSkipDir(d.Name(), excludes, dir.IncludeHidden) {
				return filepath.SkipDir
			}
			if err := watcher.Add(path); err != nil {
				log.Warn().Err(err).Str("path", path).Msg("Failed to watch subdirectory")
			}
			return nil
		})
	}
}

// handleWrite debounces a file-write event and invokes the callback after the
// debounce period.
func handleWrite(event fsnotify.Event, dirs []*config.Directory, mu *sync.Mutex, debounced map[string]*time.Timer, callback func(string)) {
	dir := FindMatchingDir(dirs, event.Name)
	if !DirectoryMatchesPath(dir, event.Name) {
		return
	}
	name := event.Name
	mu.Lock()
	if t, ok := debounced[name]; ok {
		t.Reset(debounceTime)
	} else {
		debounced[name] = time.AfterFunc(debounceTime, func() {
			mu.Lock()
			delete(debounced, name)
			mu.Unlock()
			callback(name)
		})
	}
	mu.Unlock()
}

// handleRemove processes a file removal or rename event. If the file belongs
// to a directory configured with delete_on_remove, the onRemove callback is
// invoked with the file path. Directories and files that do not match the
// configured filters are silently ignored.
func handleRemove(event fsnotify.Event, dirs []*config.Directory, onRemove func(string)) {
	if onRemove == nil {
		return
	}
	dir := FindMatchingDir(dirs, event.Name)
	if dir == nil || !dir.DeleteOnRemove || !DirectoryMatchesPath(dir, event.Name) {
		return
	}
	onRemove(event.Name)
}

// handleCreate processes a file or directory creation event: new directories
// are added to the watcher, new files matching filters are passed to the callback.
func handleCreate(event fsnotify.Event, dirs []*config.Directory, watcher *fsnotify.Watcher, callback func(string)) {
	st, err := os.Stat(event.Name)
	if err != nil {
		return
	}
	if st.IsDir() {
		dir := FindMatchingDir(dirs, event.Name)
		if dir == nil || shouldSkipDir(filepath.Base(event.Name), dir.Excludes, dir.IncludeHidden) {
			return
		}
		if !slices.Contains(watcher.WatchList(), event.Name) {
			if err := watcher.Add(event.Name); err != nil {
				log.Warn().Err(err).Str("path", event.Name).Msg("Failed to watch new directory")
			}
		}
		return
	}
	dir := FindMatchingDir(dirs, event.Name)
	if !DirectoryMatchesPath(dir, event.Name) {
		return
	}
	callback(event.Name)
}

func WatchDirectories(ctx context.Context, dirs []*config.Directory, callback func(string), onRemove func(string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close file watcher")
		}
	}()

	var mu sync.Mutex
	debounced := make(map[string]*time.Timer)

	log.Debug().Msg("Starting file watcher")
	walkAndWatch(watcher, dirs)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			switch {
			case event.Has(fsnotify.Write):
				handleWrite(event, dirs, &mu, debounced, callback)
			case event.Has(fsnotify.Create):
				handleCreate(event, dirs, watcher, callback)
			case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
				handleRemove(event, dirs, onRemove)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Error().Err(err).Msg("Watcher failed to process event")
		}
	}
}
