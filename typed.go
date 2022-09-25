package snapsdb

import (
	"reflect"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// 时间线元信息
type TimelineMateInfo struct {
	TLFirst uint32 // Timeline first record address , for data query
	TLLast  uint32 // Timeline last record address , for data write
}

type StoreData protoreflect.ProtoMessage

type dbOptions struct {
	dataPath  string
	retention time.Duration
}

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
	// 写入数据 至时间线 timestamp
	Write(timeline time.Time, data ...StoreData) error
	// 查询某个时间线数据，并返回至列表---
	// list类型应为 继承自 protoreflect.ProtoMessage 的数组
	QueryTimeline(timeline time.Time, lp_out_list interface{}) error
	// 查询某个时间区间数据，返回数据至 outmap,
	// StoreData 类型为 protobuf.proto 生成
	/*
		var out_map [string][]StoreData // keys is timeline.Format("2006-01-02 15:04:05")
		var out_map [uint32][]StoreData // keys is timeline.Unix()
		var out_map [int64][]StoreData  // keys is timeline.Unix()
		var out_map [uint64][]StoreData // keys is timeline.Unix()
	*/
	QueryBetween(begin time.Time, end time.Time, lp_out_map interface{}) error

	/* 删除时间线所属当天的存储文件*/
	DeleteStorageFile(timeline time.Time) error
}

type StoreFile interface {
	// 写入数据
	Write(timestamp time.Time, data ...StoreData) error
	// 查询某个时间线数据，并返回至列表
	QueryTimeline(timestamp time.Time, slice_pointer *reflect.Value, origin_slice *reflect.Value, element_type *reflect.Type) error
	// 查询某个时间区间数据
	QueryBetween(begin time.Time, end time.Time, map_object reflect.Value, key_type *reflect.Kind, slice_type *reflect.Type, element_type *reflect.Type) error
	// 关闭文件
	Close()
	// 读取文件时间线元信息
	ReadMateInfo(timeline int64) (*TimelineMateInfo, error)
}
