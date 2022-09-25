package snapsdb

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"

	"strconv"
	"strings"
	"sync"
	"syscall"
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
		dataPath:  "./snapdb",
		retention: TimestampOf7Day,
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
		basePath:   bpath,
		retention:  options.retention,
		opendFiles: make(map[int64]StoreFile),
	}
	go db.monitorRetention()
	return &db, nil
}

type defaultDB struct {
	basePath   string
	opendFiles map[int64]StoreFile
	retention  time.Duration
	mutex      sync.Mutex
}

// parse map interface typed
// returm [map_pointer,map_type,map_keytype,slice_type,element_type,error]
func (db *defaultDB) parseMapInterface(key_map interface{}) (*reflect.Value, *reflect.Type, *reflect.Kind, *reflect.Type, *reflect.Type, error) {
	// 获取slice的类型
	// read metainfo
	origin_map := reflect.ValueOf(key_map)
	if origin_map.Kind() != reflect.Ptr {
		return nil, nil, nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	map_pointer := origin_map.Elem()
	if !map_pointer.IsValid() {
		return nil, nil, nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	origin_map = map_pointer
	// get list typed
	type_interface := reflect.TypeOf(key_map)
	// get list typed pointer typed
	type_map := type_interface.Elem()
	// get element typed
	type_slice := type_map.Elem()
	// get element typed
	type_keys := type_map.Key()
	if type_keys.Kind() != reflect.String && type_keys.Kind() != reflect.Int64 {
		return nil, nil, nil, nil, nil, errors.New("map key must be of type string or int64'")
	}
	type_element := type_slice.Elem()
	type_key := type_keys.Kind()
	return &map_pointer, &type_map, &type_key, &type_slice, &type_element, nil
}

// parse slice interface typed
// returm [slice_pointer,origin_slice,element_type,error]
func (db *defaultDB) parseSliceInterface(list interface{}, clearList bool) (*reflect.Value, *reflect.Value, *reflect.Type, error) {
	// read metainfo
	typed := reflect.ValueOf(list)
	if typed.Kind() != reflect.Ptr {
		return nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	slice_pointer := typed.Elem()
	if !slice_pointer.IsValid() {
		return nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	origin_slice := slice_pointer
	// get list typed
	type_interface := reflect.TypeOf(list)
	// get list typed pointer typed
	type_slice := type_interface.Elem()
	// get element typed
	element_type := type_slice.Elem()
	if clearList {
		origin_slice = reflect.Zero(origin_slice.Type())
	}
	return &slice_pointer, &origin_slice, &element_type, nil
}

func (db *defaultDB) QueryTimeline(timeline time.Time, out_list interface{}) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := util.GetUnixOfDay(timeline)
	storeFile, err := db.loadFile(timebaseline, false)
	if err == nil {
		slice_pointer, origin_slice, element_type, err := db.parseSliceInterface(out_list, false)
		if err != nil {
			return err
		}
		return storeFile.QueryTimeline(timeline, slice_pointer, origin_slice, element_type)
	}
	return nil
}

func (db *defaultDB) QueryBetween(begin time.Time, end time.Time, out_map interface{}) error {
	dis := end.Sub(begin)
	if dis < 0 {
		return errors.New("is not a valid time range")
	}
	map_pointer, map_type, key_type, slice_type, element_type, err := db.parseMapInterface(out_map)
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
		// fmt.Printf("命中文件%d\n", timebaseline)
		storeFile, err := db.loadFile(timebaseline, false)
		if err == nil {
			err = storeFile.QueryBetween(begin, end, map_object, key_type, slice_type, element_type)
			if err != nil {
				return err
			}
		}
		timebasetime = timebasetime.Add(TimestampOf1Day)
	}
	map_pointer.Set(map_object)
	return nil
}

func (db *defaultDB) Write(timeline time.Time, data ...StoreData) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := util.GetUnixOfDay(timeline)
	storeFile, err := db.loadFile(timebaseline, true)
	if err != nil {
		return err
	}
	return storeFile.Write(timeline, data...)
}

func (db *defaultDB) loadFile(timebaseline int64, autoCreated bool) (StoreFile, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	file := db.opendFiles[timebaseline]
	if file == nil {
		filepath := filepath.Join(db.basePath, fmt.Sprintf("%d.bin", timebaseline))
		stroe, err := loadStoreFile(filepath, timebaseline, autoCreated)
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

func (db *defaultDB) DeleteStorageFile(timeline time.Time) error {
	timebaseline := util.GetUnixOfDay(timeline)
	filepath := filepath.Join(db.basePath, fmt.Sprintf("%d.bin", timebaseline))
	if util.FileExist(filepath) {
		db.freeFile(timebaseline)
		os.Remove(filepath)
		return nil
	}
	return errors.New("file not found")
}

func (db *defaultDB) monitorRetention() {
	ticker := time.NewTicker(time.Second)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			now := time.Now()
			filepath.Walk(db.basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".bin") {
					return nil
				}
				filetimestamp, _, ok := strings.Cut(info.Name(), ".")
				if ok {
					if num, err := strconv.ParseInt(filetimestamp, 0, 8); err == nil {
						timestamp := time.Unix(num, 0)
						if now.Sub(timestamp) > db.retention {
							// 释放文件
							db.freeFile(num)
							// 删除文件
							os.Remove(path)
						}
					}
				}
				return nil
			})

		}
	}
}
