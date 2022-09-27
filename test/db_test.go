package test

import (
	"fmt"
	"path/filepath"

	"testing"
	"time"

	"github.com/vblegend/snapsdb"
	"github.com/vblegend/snapsdb/test/types"
	"github.com/vblegend/snapsdb/util"

	"google.golang.org/protobuf/proto"
)

func InitDB() snapsdb.SnapsDB {
	snapPath := filepath.Join(util.AssemblyDir(), "../snapdata/proc")
	db, err := snapsdb.InitDB(
		snapsdb.WithDataPath(snapPath),
		snapsdb.WithDataRetention(snapsdb.TimestampOf14Day),
		snapsdb.WithTimeKeyFormat("2006-01-02 15:04:05"),
	)
	if err != nil {
		panic(err)
	}
	return db
}

// 测试 snapshotDB 批量插入
func TestSnapshotDBWrite(_ *testing.T) {
	fmt.Println("开始测试")
	db := InitDB()
	v1 := &types.ProcessInfo{Pid: 1, Name: "docker-compose - 1", Cpu: 10.01, Mem: 91.23, Virt: 10000000000, Res: 110000000000000}
	v2 := &types.ProcessInfo{Pid: 2, Name: "docker-compose - 2", Cpu: 20.02, Mem: 92.34, Virt: 20000000000, Res: 220000000000000}
	v3 := &types.ProcessInfo{Pid: 3, Name: "docker-compose - 3", Cpu: 30.03, Mem: 93.45, Virt: 30000000000, Res: 330000000000000}
	v4 := &types.ProcessInfo{Pid: 4, Name: "docker-compose - 4", Cpu: 40.04, Mem: 94.56, Virt: 40000000000, Res: 440000000000000}
	v5 := &types.ProcessInfo{Pid: 5, Name: "docker-compose - 5", Cpu: 50.05, Mem: 95.67, Virt: 50000000000, Res: 550000000000000}

	data1, _ := proto.Marshal(v1)
	count := 172800
	array := []snapsdb.StoreData{v1, v2, v3, v4, v5}

	start := time.Now() // 获取当前时间
	for i := 0; i < count; i++ {
		timestamp := time.Date(2022, 9, 22, 0, 0, i, 0, time.Local)
		db.Write(timestamp, array...)
	}
	cost := time.Since(start)
	total := count * len(array)
	bySec := float64(total) / cost.Seconds()
	fmt.Printf("写入完毕...\n共计写入%s条数据\n每条数据长度%s字节\n单位数据量%s条\n共用时%s\n每秒写入量%s条\n",
		util.Green(fmt.Sprintf("%d", total)),
		util.Green(fmt.Sprintf("%d", len(data1))),
		util.Green(fmt.Sprintf("%d", len(array))),
		util.Green(fmt.Sprintf("%v", cost)),
		util.Green(fmt.Sprintf("%.0f", bySec)))
}

// 测试 snapshotDB 单条插入
func TestSnapshotDBWriteOnce(t *testing.T) {
	fmt.Println("开始测试")
	db := InitDB()
	v1 := &types.ProcessInfo{Pid: 1, Name: "docker-compose - 1", Cpu: 10.01, Mem: 91.23, Virt: 10000000000, Res: 110000000000000}
	v2 := &types.ProcessInfo{Pid: 2, Name: "docker-compose - 2", Cpu: 20.02, Mem: 92.34, Virt: 20000000000, Res: 220000000000000}
	v3 := &types.ProcessInfo{Pid: 3, Name: "docker-compose - 3", Cpu: 30.03, Mem: 93.45, Virt: 30000000000, Res: 330000000000000}
	v4 := &types.ProcessInfo{Pid: 4, Name: "docker-compose - 4", Cpu: 40.04, Mem: 94.56, Virt: 40000000000, Res: 440000000000000}
	v5 := &types.ProcessInfo{Pid: 5, Name: "docker-compose - 5", Cpu: 50.05, Mem: 95.67, Virt: 50000000000, Res: 550000000000000}
	timestamp := time.Date(2022, 01, 01, 01, 01, 01, 01, time.Local)
	fmt.Println(timestamp.Format("2006-01-02 15:04:05"))
	start := time.Now() // 获取当前时间
	db.Write(timestamp, v1, v2, v3, v4, v5)
	cost := time.Since(start)
	fmt.Printf("写入完毕，共计%d条数据，用时%v\n", 5, cost)
}

// 测试 snapshotDB 的时间线查询
func TestSnapshotDBQuery(t *testing.T) {
	db := InitDB()
	timestamp := time.Date(2022, 9, 22, 13, 27, 43, 0, time.Local)
	list := make([]types.ProcessInfo, 0)
	list = append(list, types.ProcessInfo{Pid: 5, Name: "docker-compose - 1111", Cpu: 50.05, Mem: 95.67, Virt: 50000000000, Res: 550000000000000})
	start := time.Now()
	err := db.QueryTimeline(timestamp, &list)
	cost := time.Since(start)
	if err != nil {
		fmt.Println(util.Red(err.Error()))
	}
	fmt.Println(list)
	fmt.Printf("查询完毕，用时%v\n", cost)
}

// 测试 snapshotDB 的时间段查询
func TestSnapshotDBQueryBetween(t *testing.T) {
	db := InitDB()
	beginTimestamp := time.Date(2022, 9, 22, 5, 0, 00, 0, time.Local)
	endTimestamp := time.Date(2022, 9, 22, 5, 2, 00, 0, time.Local)
	outmap := make(map[string][]types.ProcessInfo)
	start := time.Now()
	err := db.QueryBetween(beginTimestamp, endTimestamp, &outmap)
	cost := time.Since(start)
	if err != nil && err != snapsdb.ErrorDBFileNotHit {
		fmt.Println(util.Red(err.Error()))
	}
	fmt.Println(outmap)
	fmt.Printf("查询完毕，用时%v\n", cost)
}
