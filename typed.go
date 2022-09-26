package snapsdb

import (
	"reflect"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// 时间线元信息
type timelineMateInfo struct {
	TLFirst uint32 // Timeline first record address , for data query
	TLLast  uint32 // Timeline last record address , for data write
}

type dbOptions struct {
	dataPath      string
	retention     time.Duration
	timekeyformat string
}

type StoreData protoreflect.ProtoMessage

/* time */
const (
	// Timestamp length in 1 day
	TimestampOf1Day = time.Hour * 24
	// Timestamp length in 7 day
	TimestampOf7Day = TimestampOf1Day * 7
	// Timestamp length in 14 day
	TimestampOf14Day = TimestampOf1Day * 14
	// Timestamp length in 30 day
	TimestampOf30Day = TimestampOf1Day * 30
	// Timestamp length in 1 year
	TimestampOf1Year = TimestampOf1Day * 365
	// Timestamp length in 100 year
	TimestampOf100Year = TimestampOf1Year * 100
)

/* stroage file */
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

type SnapsDB interface {
	// write one or more pieces of data to the timeline.
	Write(timeline time.Time, data ...StoreData) error
	// Query a certain timeline data, and return to the slice
	// the slice type should be inherited from protoreflect.ProtoMessage
	/*
		@example
		timestamp := time.Date(2020, 9, 22, 13, 27, 43, 0, time.Local)
		list := make([]types.ProcessInfo, 0)
		db.QueryTimeline(timestamp, &list)
	*/
	QueryTimeline(timeline time.Time, lp_out_slice interface{}) error
	// 查询某个时间区间数据，返回数据至 lp_out_map,
	// 类型为 protobuf.proto 生成
	/*
		var out_map [string][]StoreData // keys is timeline.Format(timekeyformat)
		var out_map [uint32][]StoreData // keys is timeline.Unix()
		var out_map [int64][]StoreData  // keys is timeline.Unix()
		var out_map [uint64][]StoreData // keys is timeline.Unix()

		@example

		db := InitDB()
		beginTimestamp := time.Date(2020, 9, 22, 5, 0, 00, 0, time.Local)
		endTimestamp := time.Date(2020, 9, 22, 5, 2, 00, 0, time.Local)
		map := make(map[string][]types.ProcessInfo)
		db.QueryBetween(beginTimestamp, endTimestamp, &map)
	*/
	QueryBetween(begin time.Time, end time.Time, lp_out_map interface{}) error

	/* Delete the stored file for the current day of the timeline */
	DeleteStorageFile(timeline time.Time) error

	/* Get data file storage directory */
	StorageDirectory() string

	/* Determine whether the specified date has expired in the database */
	IsExpired(timeline time.Time, now *time.Time) bool

	/* Destroy database objects */
	Dispose() error
}

type StoreFile interface {
	// 写入数据
	Write(timestamp time.Time, data ...StoreData) error
	// query a timeline for data and return to a list
	QueryTimeline(timestamp time.Time, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error
	// Query the data of a certain time interval and fill it with map[][]typed
	QueryBetween(begin time.Time, end time.Time, map_object reflect.Value, key_type *reflect.Kind, slice_type *reflect.Type, element_type *reflect.Type) error
	// close file
	Close()
	// read file timeline meta information
	ReadMateInfo(timeline int64) (*timelineMateInfo, error)
	/* get file  time base line*/
	TimeBaseline() int64
}
