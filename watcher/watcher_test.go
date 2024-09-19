package watcher_test

import (
	"github.com/ssgo/tool/watcher"
	"github.com/ssgo/u"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("aaa")
		_ = os.RemoveAll("bbb")
	}()
	_ = os.MkdirAll("aaa/dir1", 0755)
	lastFile := ""
	lastEvent := ""
	w, err := watcher.Start([]string{"aaa", "bbb"}, nil, func(filename, event string) {
		//fmt.Println(u.Magenta(filename), u.BMagenta(event))
		if event != watcher.Rename {
			lastFile = filename
			lastEvent = event
		}
	})
	defer func() {
		w.Stop()
	}()
	if err != nil {
		t.Fatal(err.Error(), w.WatchList())
	}

	_ = u.WriteFile(filepath.Join("aaa", "1.txt"), "something")
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "1.txt") || lastEvent != watcher.Create {
		t.Fatal("check 1.txt", lastFile, lastEvent, w.WatchList())
	}

	_ = u.WriteFile(filepath.Join("aaa", "dir1", "2.txt"), "something")
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "2.txt") || lastEvent != watcher.Create {
		t.Fatal("check 2.txt", lastFile, lastEvent, w.WatchList())
	}

	_ = u.WriteFile(filepath.Join("bbb", "dir2", "3.txt"), "something")
	time.Sleep(100 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "3.txt") || lastEvent != watcher.Create {
		t.Fatal("check 3.txt", lastFile, lastEvent, w.WatchList())
	}

	_ = u.WriteFile(filepath.Join("bbb", "dir2", "3.txt"), "something!")
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "3.txt") || lastEvent != watcher.Change {
		t.Fatal("check 3.txt", lastFile, lastEvent, w.WatchList())
	}

	_ = os.Rename(filepath.Join("bbb", "dir2", "3.txt"), filepath.Join("bbb", "dir2", "4.txt"))
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "4.txt") || lastEvent != watcher.Create {
		t.Fatal("check 4.txt", lastFile, lastEvent, w.WatchList())
	}

	w.SetFileTypes([]string{"txt"})

	_ = os.Remove(filepath.Join("aaa", "1.txt"))
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "1.txt") || lastEvent != watcher.Remove {
		t.Fatal("check 1.txt", lastFile, lastEvent, w.WatchList())
	}

	_ = os.RemoveAll("aaa")
	time.Sleep(10 * time.Millisecond)
	if !strings.HasSuffix(lastFile, "2.txt") || lastEvent != watcher.Remove {
		t.Fatal("check 2.txt", lastFile, lastEvent, w.WatchList())
	}

}
