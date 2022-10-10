package snapsdb

import (
	"bytes"
	"encoding/binary"
	"errors"

	"io"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/vblegend/snapsdb/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// file format
// =============================
// 1.file header
// offset 0 byte
// size 16 byte
//
// magic code  offset +0
// timestamp   offset +8
// =============================
// 2.index table
// offset 16 byte
// size 691200 bytes (86400 * 8)
//
// First Record Address  offset (16 + index * 8) byte
// Last  Record Address  offset (16 + index * 8 + 4) byte
// =============================
// 3.data block
// offset (691200 + 16) byte
//
// Timestamp			 size 8 byte     offset RecordAddress + 0
// Next Record Address   size 4 byte     offset RecordAddress + 8
// data length  		 size 4 byte     offset RecordAddress + 12
// binary data   		 size ... byte   offset RecordAddress + 16

type storeFile struct {
	TimelineBegin int64      // storage file time base line of begin
	TimelineEnd   int64      // storage file time base line of end
	file          *os.File   // storage file access object
	mutex         sync.Mutex // access lock
	timeKeyFormat string
}

// load file object from timebaseline
// autoCreated = true  automatically created and initialized when file does not exist
// autoCreated = fakse return error if file does not exist
func loadStoreFile(filename string, timebaseline int64, timeKeyFormat string, autoCreated bool) (StoreFile, error) {
	filev := storeFile{TimelineBegin: timebaseline, TimelineEnd: timebaseline + TimelineLengthOfDay, timeKeyFormat: timeKeyFormat}
	var err error
	if !util.FileExist(filename) {
		if autoCreated {
			err = filev.init(filename)
		} else {
			return nil, ErrorDBFileNotHit
		}
	} else {
		err = filev.open(filename)
	}
	if err != nil {
		return nil, err
	}
	return &filev, nil
}

func (sf *storeFile) QueryBetween(begin int64, end int64, map_object reflect.Value, key_type *reflect.Kind, slice_type *reflect.Type, element_type *reflect.Type) error {
	sf.Lock()
	defer sf.Unlock()
	hitFile := end > sf.TimelineBegin && begin < sf.TimelineEnd
	if !hitFile {
		return errors.New("beyond the scope of the query")
	}
	beginTimeline := begin
	if sf.TimelineBegin > beginTimeline {
		beginTimeline = sf.TimelineBegin
	}
	endTimeline := end
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
		err := sf.queryByTimeline(timeline, &lpSlice, &slice, element_type)
		if err != nil && err != io.EOF {
			return err
		}
		// 获取 out_map key 类型
		key := sf.GetReflectKey(time.Unix(timeline, 0), *key_type)
		// 添加至 Map内
		map_object.SetMapIndex(*key, lpSlice.Elem())
	}
	return nil
}

func (sf *storeFile) GetReflectKey(timeline time.Time, key_type reflect.Kind) *reflect.Value {
	var value reflect.Value
	switch key_type {
	case reflect.String:
		value = reflect.ValueOf(timeline.Format(sf.timeKeyFormat))
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
func (sf *storeFile) QueryTimeline(timeline int64, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error {
	sf.Lock()
	defer sf.Unlock()
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
	slice_pointer.Elem().Set(*origin_slice)
	return nil
}

func (sf *storeFile) Write(timeline int64, data ...StoreData) error {
	lenObject := len(data)
	if lenObject == 0 {
		return nil
	}
	sf.Lock()
	defer sf.Unlock()
	// read metainfo
	meta, err := sf.ReadMateInfo(timeline)
	if err != nil {
		return err
	}
	// get file eof position
	writePos, err := sf.file.Seek(0, 2)
	if err != nil {
		return err
	}
	// create batch buf
	writeBuf := bytes.NewBuffer(make([]byte, 0))
	var linkedOfLast int64 = 0
	if meta.TLLast != 0 {
		linkedOfLast = int64(meta.TLLast) + NextDataOffset
	}
	if meta.TLFirst == 0 {
		meta.TLFirst = uint32(writePos)
	}
	for i, item := range data {
		position := writePos + int64(writeBuf.Len())
		outdata, err := proto.Marshal(item)
		if err != nil {
			return err
		}
		meta.TLLast = uint32(position)
		var nextDataAddr uint32 = 0
		if i < lenObject-1 {
			nextDataAddr = uint32(position) + DataHeaderLen + uint32(len(outdata))
		}
		binary.Write(writeBuf, binary.LittleEndian, timeline)             // timeline   8byte
		binary.Write(writeBuf, binary.LittleEndian, nextDataAddr)         // nextdata	4byte
		binary.Write(writeBuf, binary.LittleEndian, uint32(len(outdata))) // datalen    4byte
		writeBuf.Write(outdata)                                           // data
	}
	// linked last record
	if linkedOfLast > 0 {
		nextRecordPosition := make([]byte, 4)
		binary.LittleEndian.PutUint32(nextRecordPosition, uint32(writePos))
		sf.file.WriteAt(nextRecordPosition, linkedOfLast)
	}
	// flush buf
	sf.file.Seek(0, 2)
	writeBuf.WriteTo(sf.file)
	// update metadata
	sf.writeMateInfo(timeline, meta)
	return nil
}

func (sf *storeFile) ReadMateInfo(timeline int64) (*timelineMateInfo, error) {
	if timeline < sf.TimelineBegin || timeline > sf.TimelineEnd {
		return nil, errors.New("beyond the scope of the query.")
	}
	buffer := make([]byte, 8)
	disc := timeline - sf.TimelineBegin
	offset := MateInfoSize*disc + FileHeaderOffset
	sf.file.ReadAt(buffer, offset)
	first := binary.LittleEndian.Uint32(buffer[:4])
	last := binary.LittleEndian.Uint32(buffer[4:])
	return &timelineMateInfo{TLFirst: first, TLLast: last}, nil
}

func (sf *storeFile) writeMateInfo(timeline int64, info *timelineMateInfo) error {
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
	binary.Write(file, binary.LittleEndian, uint64(7089841687217925715)) // file flags    offset + 0
	binary.Write(file, binary.LittleEndian, sf.TimelineBegin)            // timebaseline  offset + 8
	for i := int64(0); i < TimelineLengthOfDay; i++ {                    //  			  offset + (16 + timeline * UnitSize) UnitSize = 8
		binary.Write(file, binary.LittleEndian, uint32(0)) // TLFirst	  				  offset + 0
		binary.Write(file, binary.LittleEndian, uint32(0)) // TLLast	  				  offset + 4
	}
	sf.file = file
	return err
}
func (sf *storeFile) TimeBaseline() int64 {
	return sf.TimelineBegin
}
