package snapsdb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"go-admin/core/sdk/pkg"
	"io"
	"os"
	"reflect"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// 一天的时间线长度
	TimelineLengthOfDay = int64(86400)
	// 文件头的偏移量
	FileHeaderOffset = int64(16)
	// 单条时间线元数据的大小
	MateInfoSize = int64(8)
	// 一天的时间线元数据总大小
	MateTableSize = int64(691200)
	// 下一条数据记录的指针偏移位置（相对于数据记录的开始位置）
	NextDataOffset = int64(8)
	// 文件第一条数据的偏移位置
	FileDataOffset = uint32(MateTableSize + FileHeaderOffset)
)

type StoreFile interface {
	// 写入数据
	Write(timestamp time.Time, data ...StoreData) error
	// 查询某个时间线数据，并返回至列表
	Query(timestamp time.Time, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error
	// 查询某个时间区间数据
	QueryBetween(begin time.Time, end time.Time, map_object reflect.Value, key_type *reflect.Kind, slice_type *reflect.Type, element_type *reflect.Type) error
	// 关闭文件
	Close()
	// 读取文件时间线元信息
	ReadMateInfo(timeline int64) (*TimelineMateInfo, error)
}

type storeFile struct {
	TimelineBegin int64      // 存储文件的时间基线
	TimelineEnd   int64      // 存储文件的时间基线
	file          *os.File   //文件访问对象
	mutex         sync.Mutex // 文件操作互斥锁
}

// 加载时间片存储文件
// 写 autoCreated 为 true  时 文件不存在自动创建初始化文件
// 读 autoCreated 为 fakse 时 文件不存在返回error
func LoadStoreFile(filename string, timebaseline int64, autoCreated bool) (StoreFile, error) {
	filev := storeFile{TimelineBegin: timebaseline, TimelineEnd: timebaseline + TimelineLengthOfDay}
	var err error
	if !pkg.FileExist(filename) {
		if autoCreated {
			err = filev.init(filename)
		} else {
			return nil, fmt.Errorf("file “%s” not found", filename)
		}
	} else {
		err = filev.open(filename)
	}
	if err != nil {
		return nil, err
	}
	return &filev, nil
}

func (sf *storeFile) QueryBetween(begin time.Time, end time.Time, map_object reflect.Value, key_type *reflect.Kind, slice_type *reflect.Type, element_type *reflect.Type) error {
	sf.Lock()
	defer sf.Unlock()
	hitFile := end.Unix() > sf.TimelineBegin && begin.Unix() < sf.TimelineEnd
	if !hitFile {
		return errors.New("查询时间与文件不一致")
	}
	beginTimeline := time.Date(begin.Year(), begin.Month(), begin.Day(), begin.Hour(), begin.Minute(), begin.Second(), 0, time.Local).Unix()
	if sf.TimelineBegin > beginTimeline {
		beginTimeline = sf.TimelineBegin
	}
	endTimeline := time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), end.Second(), 0, time.Local).Unix()
	if sf.TimelineEnd < endTimeline {
		endTimeline = sf.TimelineEnd
	}
	length := int(endTimeline - beginTimeline)
	for i := 0; i <= length; i++ {
		timeline := beginTimeline + int64(i)
		// 创建切片对象
		slice := reflect.MakeSlice(*slice_type, 0, 16)
		// 创建 切片指针
		lpSlice := reflect.New(*slice_type)
		// 指针指向 切片对象
		lpSlice.Elem().Set(slice)
		// 获取切片指针
		slice_pointer := lpSlice.Elem()
		err := sf.queryByTimeline(timeline, &slice_pointer, &slice, element_type)
		if err != nil && err != io.EOF {
			return err
		}
		// 获取 out_map key 类型
		key := sf.GetReflectKey(time.Unix(timeline, 0), *key_type)
		// 添加至 Map内
		map_object.SetMapIndex(*key, slice_pointer)
	}
	return nil
}

func (sf *storeFile) GetReflectKey(timeline time.Time, key_type reflect.Kind) *reflect.Value {
	var value reflect.Value
	switch key_type {
	case reflect.String:
		value = reflect.ValueOf(timeline.Format("2006-01-02 15:04:05"))
	case reflect.Int64:
		value = reflect.ValueOf(timeline.Unix())
	case reflect.Uint64:
		value = reflect.ValueOf(uint64(timeline.Unix()))
	case reflect.Uint32:
		value = reflect.ValueOf(uint32(timeline.Unix()))
	case reflect.Int:
		value = reflect.ValueOf(int(timeline.Unix()))
	}
	return &value
}

// 查询某个时间线上的所有数据
func (sf *storeFile) Query(timestamp time.Time, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error {
	sf.Lock()
	defer sf.Unlock()
	// 获取时间戳的所属时间线
	timeline := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, time.Local).Unix()
	return sf.queryByTimeline(timeline, slice_pointer, origin_slice, element_type)
}

// 查询某个时间线上所有的数据
func (sf *storeFile) queryByTimeline(timeline int64, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error {
	// read metainfo
	meta, err := sf.ReadMateInfo(timeline)
	if err != nil {
		return err
	}
	nextRecord := meta.TLFirst
	for nextRecord != 0 {
		buffer := make([]byte, 16)
		dataPos := int64(nextRecord + 16)
		readsize, err := sf.file.ReadAt(buffer, int64(nextRecord))
		if err != nil || readsize != 16 {
			return err
		}
		_timeline := binary.LittleEndian.Uint32(buffer[:8])
		if int64(_timeline) != timeline {
			break
		}
		_nextdata := binary.LittleEndian.Uint32(buffer[8:12])
		_datalen := binary.LittleEndian.Uint32(buffer[12:])
		buffer = make([]byte, _datalen)
		sf.file.ReadAt(buffer, dataPos)
		//
		refObject := reflect.New(*element_type)
		object := refObject.Interface()
		switch typed := object.(type) {
		case protoreflect.ProtoMessage:
			err := proto.Unmarshal(buffer, typed)
			if err == nil {
				*origin_slice = reflect.Append(*origin_slice, reflect.ValueOf(typed).Elem())
			}
		}
		nextRecord = _nextdata
	}
	kls := origin_slice.Interface()
	kil := slice_pointer.Interface()
	fmt.Println(kil, kls)
	slice_pointer.Set(*origin_slice)
	return nil
}

func (sf *storeFile) Write(timestamp time.Time, data ...StoreData) error {
	sf.Lock()
	defer sf.Unlock()
	// 获取时间戳的所属时间线
	timeline := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, time.Local).Unix()
	// read metainfo
	meta, err := sf.ReadMateInfo(timeline)
	if err != nil {
		return err
	}
	for _, item := range data {
		position, err := sf.file.Seek(0, 2)
		if err != nil {
			return err
		}
		outdata, err := proto.Marshal(item)
		if err != nil {
			return err
		}
		if meta.TLLast != 0 { // > FileDataOffset
			// 修改 last记录的 netdata记录
			nextRecordPosition := make([]byte, 4)
			binary.LittleEndian.PutUint32(nextRecordPosition, uint32(position))
			sf.file.WriteAt(nextRecordPosition, int64(meta.TLLast)+NextDataOffset)
		}
		meta.TLLast = uint32(position)
		if meta.TLFirst == 0 {
			meta.TLFirst = uint32(position)
		}
		binary.Write(sf.file, binary.LittleEndian, timeline) // timeline
		position, err = sf.file.Seek(0, 2)
		binary.Write(sf.file, binary.LittleEndian, uint32(0))            // nextdata
		binary.Write(sf.file, binary.LittleEndian, uint32(len(outdata))) // datalen
		sf.file.Write(outdata)
	}
	sf.writeMateInfo(timeline, meta)
	return nil
}

func (sf *storeFile) ReadMateInfo(timeline int64) (*TimelineMateInfo, error) {
	if timeline < sf.TimelineBegin || timeline > sf.TimelineEnd {
		return nil, errors.New("超出范围。")
	}
	buffer := make([]byte, 8)
	disc := timeline - sf.TimelineBegin
	offset := MateInfoSize*disc + FileHeaderOffset
	sf.file.ReadAt(buffer, offset)
	first := binary.LittleEndian.Uint32(buffer[:4])
	last := binary.LittleEndian.Uint32(buffer[4:])
	return &TimelineMateInfo{TLFirst: first, TLLast: last}, nil
}

func (sf *storeFile) writeMateInfo(timeline int64, info *TimelineMateInfo) error {
	if timeline < sf.TimelineBegin || timeline > sf.TimelineEnd {
		return errors.New("超出范围。")
	}
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint32(buffer[:4], info.TLFirst)
	binary.LittleEndian.PutUint32(buffer[4:], info.TLLast)
	disc := timeline - sf.TimelineBegin
	offset := MateInfoSize*disc + FileHeaderOffset
	sf.file.WriteAt(buffer, offset)
	return nil
}

func (sf *storeFile) open(filepath string) error {
	var err error = nil
	sf.file, err = os.OpenFile(filepath, os.O_RDWR, 0777)
	return err
}

func (sf *storeFile) Lock() {
	sf.mutex.Lock()
}

func (sf *storeFile) Unlock() {
	sf.mutex.Unlock()
}

func (sf *storeFile) Close() {
	sf.Lock()
	defer sf.Unlock()
	sf.file.Close()
	sf.file = nil
}
func (sf *storeFile) init(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	// file.ReadAt(00, 00)
	binary.Write(file, binary.LittleEndian, uint64(7089841687217925715)) // file flags
	binary.Write(file, binary.LittleEndian, sf.TimelineBegin)            // timebaseline
	for i := int64(0); i < TimelineLengthOfDay; i++ {
		binary.Write(file, binary.LittleEndian, uint32(0)) // TLFirst
		binary.Write(file, binary.LittleEndian, uint32(0)) // TLLast
	}
	sf.file = file
	return err
}
