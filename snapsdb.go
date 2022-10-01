package snapsdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"sync"
	"time"

	"github.com/vblegend/snapsdb/util"
)

// 初始化一个数据库
/*
	snapPath := filepath.Join(pkg.AssemblyDir(), "snapsdata/proc")
	snapDB, err := snapsdb.InitDB(snapsdb.WithDataPath(snapPath), snapsdb.WithDataRetention(time.Hour*24*14))
*/
func InitDB(opts ...Option) (SnapsDB, error) {
	options := &dbOptions{
		dataPath:      "./data",
		retention:     TimestampOf7Day,
		timekeyformat: "2006-01-02 15:04:05",
	}
	for _, opt := range opts {
		opt(options)
	}
	bpath, err := filepath.Abs(options.dataPath)
	if err != nil {
		return nil, err
	}
	err = util.MkDirIfNotExist(bpath)
	if err != nil {
		return nil, err
	}
	db := defaultDB{
		basePath:      bpath,
		retention:     options.retention,
		opendFiles:    make(map[int64]StoreFile),
		timeKeyFormat: options.timekeyformat,
	}
	return registerDB(&db)
}

type defaultDB struct {
	basePath      string
	opendFiles    map[int64]StoreFile
	retention     time.Duration
	mutex         sync.Mutex
	timeKeyFormat string
	isDisposed    bool
}

func (db *defaultDB) StorageDirectory() string {
	return db.basePath
}

func (db *defaultDB) QueryTimelineUnix(timeline int64, lp_out_slice interface{}) error {
	return db.QueryTimeline(time.Unix(timeline, 0), lp_out_slice)
}

func (db *defaultDB) QueryTimeline(timeline time.Time, out_list interface{}) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := util.GetUnixOfDay(timeline)
	storeFile, err := db.loadFile(timebaseline, false)
	if err != nil && err != ErrorDBFileNotHit {
		return err
	}
	if err == nil {
		slice_pointer, origin_slice, element_type, err := util.ParseSlicePointer(out_list, false)
		if err != nil {
			return err
		}
		return storeFile.QueryTimeline(timeline.Unix(), slice_pointer, origin_slice, element_type)
	}
	return nil
}

func (db *defaultDB) QueryBetweenUnix(begin int64, end int64, lp_out_map interface{}) error {
	return db.QueryBetween(time.Unix(begin, 0), time.Unix(end, 0), lp_out_map)
}

func (db *defaultDB) QueryBetween(begin time.Time, end time.Time, out_map interface{}) error {
	dis := end.Sub(begin)
	if dis < 0 {
		return errors.New("is not a valid time range")
	}
	map_pointer, map_type, key_type, slice_type, element_type, err := util.ParseMapPointer(out_map)
	if err != nil {
		return err
	}
	//
	map_object := reflect.MakeMap(*map_type)
	// 取 begin 当天
	timebasetime := util.GetTimeOfDay(begin)
	for {
		// 如果 时间基线大于 end 则退出
		if timebasetime.Sub(end) > 0 {
			break
		}
		timebaseline := timebasetime.Unix()
		storeFile, err := db.loadFile(timebaseline, false)
		if err != nil && err != ErrorDBFileNotHit {
			return err
		} else if err == nil {
			err = storeFile.QueryBetween(begin.Unix(), end.Unix(), map_object, key_type, slice_type, element_type)
			if err != nil {
				return err
			}
		}
		timebasetime = timebasetime.Add(TimestampOf1Day)
	}
	map_pointer.Elem().Set(map_object)
	return nil
}

func (db *defaultDB) WriteUnix(timeline int64, data ...StoreData) error {
	return db.Write(time.Unix(timeline, 0), data...)
}

func (db *defaultDB) Write(timeline time.Time, data ...StoreData) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := util.GetUnixOfDay(timeline)
	storeFile, err := db.loadFile(timebaseline, true)
	if err != nil {
		return err
	}
	return storeFile.Write(timeline.Unix(), data...)
}

func (db *defaultDB) loadFile(timebaseline int64, autoCreated bool) (StoreFile, error) {
	if db.isDisposed {
		return nil, errors.New("Database object has been destroyed")
	}
	db.mutex.Lock()
	defer db.mutex.Unlock()
	file := db.opendFiles[timebaseline]
	if file == nil {
		filepath := filepath.Join(db.basePath, fmt.Sprintf("%d.bin", timebaseline))
		stroe, err := loadStoreFile(filepath, timebaseline, db.timeKeyFormat, autoCreated)
		if err != nil {
			return nil, err
		}
		db.opendFiles[timebaseline] = stroe
	}
	return db.opendFiles[timebaseline], nil
}

func (db *defaultDB) freeFile(timebaseline int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	file := db.opendFiles[timebaseline]
	if file != nil {
		db.opendFiles[timebaseline].Close()
		delete(db.opendFiles, timebaseline)
	}
}

func (db *defaultDB) DeleteStorageFileUnix(timeline int64) error {
	return db.DeleteStorageFile(time.Unix(timeline, 0))
}

func (db *defaultDB) DeleteStorageFile(timeline time.Time) error {
	timebaseline := util.GetUnixOfDay(timeline)
	filepath := filepath.Join(db.basePath, fmt.Sprintf("%d.bin", timebaseline))
	if util.FileExist(filepath) {
		db.freeFile(timebaseline)
		return os.Remove(filepath)
	}
	return errors.New("file not found")
}

func (db *defaultDB) Dispose() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.isDisposed = true
	for k, _ := range db.opendFiles {
		db.freeFile(k)
	}
	return unRegisterDB(db)
}

func (db *defaultDB) IsExpired(timeline time.Time, now *time.Time) bool {
	if now == nil {
		n := time.Now()
		now = &n
	}
	return now.Sub(timeline) > db.retention
}
