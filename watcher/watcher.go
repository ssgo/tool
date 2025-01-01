package watcher

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ssgo/u"
)

const (
	Create = "create"
	Change = "change"
	Remove = "remove"
	Rename = "rename"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	isRunning bool
	fileTypes []string
	ignores   []string
	callback  func(string, string)
	stopChan  chan bool
	hash      map[string]string
	hashLock  sync.RWMutex
}

func (w *Watcher) Stop() {
	if !w.isRunning {
		return
	}
	w.stopChan = make(chan bool, 1)
	w.isRunning = false
	if w.watcher != nil {
		_ = w.watcher.Close()
	}
	<-w.stopChan
	w.watcher = nil
}

func (w *Watcher) inType(filename string) bool {
	if len(w.fileTypes) == 0 {
		return true
	}
	for _, fileType := range w.fileTypes {
		if strings.HasSuffix(filename, fileType) {
			return true
		}
	}
	return false
}

func (w *Watcher) Add(path string) error {
	return w.add(path, false)
}

func getHash(filename string) string {
	if fp, err := os.Open(filename); err == nil {
		defer fp.Close()
		hash := md5.New()
		if _, err = io.Copy(hash, fp); err == nil {
			return hex.EncodeToString(hash.Sum(nil))
		}
	}
	return ""
}

func (w *Watcher) isIgnore(filename string) bool {
	for _, ignore := range w.ignores {
		if strings.HasPrefix(ignore, "*/") {
			if strings.Contains(filename, strings.ReplaceAll(ignore[1:], "/", string(os.PathSeparator))) {
				return true
			}
		} else if strings.HasPrefix(filename, ignore) {
			return true
		}
	}
	return false
}

func (w *Watcher) add(path string, checkFile bool) error {
	if !w.isRunning {
		return nil
	}
	if absPath, err := filepath.Abs(path); err == nil {
		path = absPath
	}
	if !u.FileExists(path) {
		_ = os.MkdirAll(path, 0755)
	}
	if err := w.watcher.Add(path); err != nil {
		return err
	} else {
		var outErr error
		for _, f := range u.ReadDirN(path) {
			if !w.isRunning {
				break
			}
			if w.isIgnore(f.FullName) {
				continue
			}
			if f.IsDir {
				if err := w.Add(f.FullName); err != nil {
					outErr = err
				}
			} else {
				// fmt.Println("add file", u.BMagenta(f.FullName))
				w.hashLock.Lock()
				w.hash[f.FullName] = getHash(f.FullName)
				w.hashLock.Unlock()
				if checkFile && w.inType(f.FullName) {
					w.callback(f.FullName, Create)
				}
			}
		}
		return outErr
	}
}

func (w *Watcher) Remove(path string) {
	if !w.isRunning {
		return
	}
	eventFileDir := path + string(os.PathSeparator)
	for _, item := range w.watcher.WatchList() {
		if item == path || strings.HasPrefix(item, eventFileDir) {
			_ = w.watcher.Remove(item)
		}
	}
}

func (w *Watcher) SetFileTypes(fileTypes []string) {
	if !w.isRunning {
		return
	}
	if fileTypes == nil {
		fileTypes = make([]string, 0)
	}
	for i, fileType := range fileTypes {
		if !strings.HasPrefix(fileType, ".") {
			fileTypes[i] = "." + fileType
		}
	}
	w.fileTypes = fileTypes
}

func (w *Watcher) WatchList() []string {
	if !w.isRunning {
		return nil
	}
	return w.watcher.WatchList()
}

func Start(paths, fileTypes, ignores []string, callback func(filename string, event string)) (*Watcher, error) {
	if watcher, err := fsnotify.NewWatcher(); err == nil {
		if paths == nil {
			paths = make([]string, 0)
		}
		if ignores == nil {
			ignores = make([]string, 0)
		}
		// */ 开头的匹配任意路径的开头，否则匹配绝对路径
		for i, ignorePath := range ignores {
			if strings.HasPrefix(ignorePath, "*/") {
				continue
			}
			if absPath, err := filepath.Abs(ignorePath); err == nil {
				ignores[i] = absPath
			}
		}

		w := &Watcher{
			watcher:   watcher,
			callback:  callback,
			isRunning: true,
			ignores:   ignores,
			hash:      map[string]string{},
		}
		w.SetFileTypes(fileTypes)
		for _, path := range paths {
			_ = w.add(path, false)
		}

		go func() {
			for w.isRunning {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						w.isRunning = false
						break
					}

					eventFilename := event.Name
					if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						w.Remove(eventFilename)
						if w.inType(eventFilename) && !w.isIgnore(eventFilename) {
							if event.Has(fsnotify.Remove) {
								callback(eventFilename, Remove)
							} else {
								callback(eventFilename, Rename)
							}
						}
					} else if event.Has(fsnotify.Write) {
						if w.inType(eventFilename) && !w.isIgnore(eventFilename) {
							callback(eventFilename, Change)
						}
					} else if event.Has(fsnotify.Create) {
						fileInfo := u.GetFileInfo(event.Name)
						if fileInfo.IsDir {
							_ = w.add(eventFilename, true)
						} else {
							if w.inType(eventFilename) && !w.isIgnore(eventFilename) {
								callback(eventFilename, Create)
							}
						}
					}
				case _, ok := <-w.watcher.Errors:
					if !ok {
						w.isRunning = false
						break
					}
				}
			}
			w.stopChan <- true
		}()

		return w, nil
	} else {
		return nil, err
	}
}
