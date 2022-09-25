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

type SnapsDB interface {
	// 写入数据 至时间线 timestamp
	Write(timeline time.Time, data ...StoreData) error
	// 查询某个时间线数据，并返回至列表---
	// list类型应为 继承自 protoreflect.ProtoMessage 的数组
	Query(timeline time.Time, out_list interface{}) error
	// 查询某个时间区间数据，返回数据至 outmap,
	// StoreData 类型为 protobuf.proto 生成
	/*
		var out_map [string][]StoreData
		var out_map [uint32][]StoreData
		var out_map [int64][]StoreData
		var out_map [uint64][]StoreData
	*/
	QueryBetween(begin time.Time, end time.Time, out_map interface{}) error

	/* 删除时间线所属当天的存储文件*/
	DeleteFile(timeline time.Time) error
}

// 初始化一个数据库
/*
	snapPath := filepath.Join(pkg.AssemblyDir(), "snapsdata/proc")
	snapDB, err := snapsdb.InitDB(snapsdb.WithDataPath(snapPath), snapsdb.WithDataRetention(time.Hour*24*14))
*/
func InitDB(opts ...Option) (SnapsDB, error) {
	options := &dbOptions{
		dataPath:  "./snapdb",
		retention: time.Hour * 24 * 7,
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
func (db *defaultDB) parseMapInterface(str_key_map interface{}) (*reflect.Value, *reflect.Type, *reflect.Kind, *reflect.Type, *reflect.Type, error) {
	// 获取slice的类型
	// read metainfo
	origin_map := reflect.ValueOf(str_key_map)
	if origin_map.Kind() != reflect.Ptr {
		return nil, nil, nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	map_pointer := origin_map.Elem()
	if !map_pointer.IsValid() {
		return nil, nil, nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	origin_map = map_pointer
	// get list typed
	type_interface := reflect.TypeOf(str_key_map)
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

func (db *defaultDB) Query(timestamp time.Time, out_list interface{}) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, time.Local).Unix()
	storeFile, err := db.loadFile(timebaseline, false)
	if err == nil {
		slice_pointer, origin_slice, element_type, err := db.parseSliceInterface(out_list, false)
		if err != nil {
			return err
		}
		return storeFile.Query(timestamp, slice_pointer, origin_slice, element_type)
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
	timebasetime := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, time.Local)
	for {
		// 如果 时间基线大于 end 则退出
		if timebasetime.Sub(end) > 0 {
			break
		}
		timebaseline := timebasetime.Unix()
		fmt.Printf("命中文件%d\n", timebaseline)
		storeFile, err := db.loadFile(timebaseline, false)
		if err == nil {
			err = storeFile.QueryBetween(begin, end, map_object, key_type, slice_type, element_type)
			if err != nil {
				return err
			}
		}
		timebasetime = timebasetime.Add(time.Hour * 24)
	}
	map_pointer.Set(map_object)
	return nil
}

func (db *defaultDB) Write(timestamp time.Time, data ...StoreData) error {
	// 获取时间戳的时间基线，当天的0点时间戳，文件名
	timebaseline := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, time.Local).Unix()
	storeFile, err := db.loadFile(timebaseline, true)
	if err != nil {
		return err
	}
	return storeFile.Write(timestamp, data...)
}

func (db *defaultDB) loadFile(timebaseline int64, autoCreated bool) (StoreFile, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	file := db.opendFiles[timebaseline]
	if file == nil {
		filepath := filepath.Join(db.basePath, fmt.Sprintf("%d.bin", timebaseline))
		stroe, err := LoadStoreFile(filepath, timebaseline, autoCreated)
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

func (db *defaultDB) DeleteFile(timeline time.Time) error {
	timebaseline := time.Date(timeline.Year(), timeline.Month(), timeline.Day(), 0, 0, 0, 0, time.Local).Unix()
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
				if info.IsDir() || !strings.HasSuffix(path, ".bin") {
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
