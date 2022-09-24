package snapsdb

import (
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// 时间线元信息
type TimelineMateInfo struct {
	// Timestamp int64  // 时间戳
	TLFirst uint32 // 时间线第一条记录  TimelineLinkItem地址
	TLLast  uint32 // 时间线最后一条记录 TimelineLinkItem地址
}

/* 时间线连表元素 实际没用到 仅展示结构 */
type TimelineLinkItem struct {
	Timestamp int64  // 时间戳
	NextData  uint64 // 下一条记录
	DataLen   uint32 // 数据长度
	Data      []byte // 数据
}

type StoreData protoreflect.ProtoMessage

// func (s *StoreData) Length() int32 {

// }

type dbOptions struct {
	dataPath  string
	retention time.Duration
}
