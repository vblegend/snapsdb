package snapsdb

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// db instances
var _db_instances map[string]SnapsDB = make(map[string]SnapsDB)

// access lock
var _checkmutex sync.Mutex

var _initcheck bool

func registerDB(db SnapsDB) (SnapsDB, error) {
	_checkmutex.Lock()
	defer _checkmutex.Unlock()
	if _db_instances[db.StorageDirectory()] != nil {
		return nil, fmt.Errorf("the data directory XX has already been initialized, and the same data directory cannot be initialized more than once.")
	}
	_db_instances[db.StorageDirectory()] = db
	checkDBStorage(db, time.Now())
	if !_initcheck {
		go monitorRetention()
		_initcheck = true
	}
	return db, nil
}

func unRegisterDB(db SnapsDB) error {
	_checkmutex.Lock()
	defer _checkmutex.Unlock()
	baseDir := db.StorageDirectory()
	cache := _db_instances[baseDir]
	if cache != nil {
		delete(_db_instances, baseDir)
		return nil
	}
	return errors.New("db not open")
}

func monitorRetention() {
	ticker := time.NewTicker(time.Second * 60 * 5)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			systemcheckcall()
		}
	}
}

func systemcheckcall() {
	_checkmutex.Lock()
	defer _checkmutex.Unlock()
	now := time.Now()
	for _, db := range _db_instances {
		checkDBStorage(db, now)
	}
}

func checkDBStorage(db SnapsDB, now time.Time) {
	filepath.Walk(db.StorageDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".bin") {
			return nil
		}
		filetimestamp, _, ok := strings.Cut(info.Name(), ".")
		if ok {
			num, err := strconv.ParseInt(filetimestamp, 0, 64)
			if err == nil {
				timestamp := time.Unix(num, 0)
				if db.IsExpired(timestamp, &now) {
					db.DeleteStorageFile(timestamp)
				}
			}
		}
		return nil
	})
}
