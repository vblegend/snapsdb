package test

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"reflect"

	"testing"
	"time"

	"github.com/vblegend/snapsdb"
	"github.com/vblegend/snapsdb/test/types"
	"github.com/vblegend/snapsdb/util"

	"google.golang.org/protobuf/proto"
)

func InitDB() snapsdb.SnapsDB {
	snapPath := filepath.Join(util.AssemblyDir(), "../snapdata/proc")
	db, err := snapsdb.InitDB(snapsdb.WithDataPath(snapPath), snapsdb.WithDataRetention(snapsdb.TimestampOf100Year))
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
		timestamp := time.Date(2020, 9, 22, 0, 0, i, 0, time.Local)
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
	timestamp := time.Date(2020, 01, 01, 01, 01, 01, 01, time.Local)
	fmt.Println(timestamp.Format("2006-01-02 15:04:05"))
	start := time.Now() // 获取当前时间
	db.Write(timestamp, v1, v2, v3, v4, v5)
	cost := time.Since(start)
	fmt.Printf("写入完毕，共计%d条数据，用时%v\n", 5, cost)
}

// 测试 snapshotDB 的时间线查询
func TestSnapshotDBQuery(t *testing.T) {
	db := InitDB()
	timestamp := time.Date(2020, 9, 22, 13, 27, 43, 0, time.Local)
	list := make([]types.ProcessInfo, 0)
	// var list []types.ProcessInfo
	list = append(list, types.ProcessInfo{Pid: 5, Name: "docker-compose - 1111", Cpu: 50.05, Mem: 95.67, Virt: 50000000000, Res: 550000000000000})
	start := time.Now()
	db.QueryTimeline(timestamp, &list)
	cost := time.Since(start)
	fmt.Printf("查询完毕，用时%v\n", cost)
	t.Log(list)
}

// 测试 snapshotDB 的时间段查询
func TestSnapshotDBQueryBetween(t *testing.T) {
	db := InitDB()
	beginTimestamp := time.Date(2020, 9, 22, 5, 0, 00, 0, time.Local)
	endTimestamp := time.Date(2020, 9, 22, 5, 2, 00, 0, time.Local)
	list := make(map[string][]types.ProcessInfo)
	start := time.Now()
	db.QueryBetween(beginTimestamp, endTimestamp, &list)
	cost := time.Since(start)
	fmt.Printf("查询完毕，用时%v\n", cost)
	t.Log(list)
}

func TestTimeline(t *testing.T) {
	tn := time.Now()
	year, month, day := tn.Date()
	timeline := time.Date(year, month, day, tn.Hour(), tn.Minute(), tn.Second(), 0, time.Local).Unix()
	fmt.Printf("%d-%d", timeline, tn.Unix())
}

func TestXXXASDASD(t *testing.T) {
	list := make(map[string][]types.ProcessInfo)
	list["xx"] = make([]types.ProcessInfo, 0)
	testReflect(&list)
	fmt.Println(list)
}

func testReflect(list interface{}) {
	// 获取slice的类型
	// read metainfo
	origin_slice := reflect.ValueOf(list)
	if origin_slice.Kind() != reflect.Ptr {
		// return nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	map_pointer := origin_slice.Elem()
	if !map_pointer.IsValid() {
		// return nil, nil, nil, errors.New("Invalid argument 'list interface{}'")
	}
	origin_slice = map_pointer
	// get list typed
	type_interface := reflect.TypeOf(list)
	// get list typed pointer typed
	type_map := type_interface.Elem()
	// get element typed
	type_slice := type_map.Elem()
	// get element typed
	type_key := type_map.Key()

	type_element := type_slice.Elem()
	// fmt.Print(type_interface, type_slice, element_type)
	// var type_interface = reflect.TypeOf(list)
	fmt.Println("slices's kind:", type_interface.Kind(), ", name:", type_interface.Name())
	// var type_slice = type_interface.Elem()
	fmt.Println("map's kind:", type_map.Kind(), ", name:", type_map.Name())

	// var type_element = type_slice.Elem()
	fmt.Println("map_key's kind:", type_key.Kind(), ", name:", type_key.Name())

	// var type_element = type_slice.Elem()
	fmt.Println("map_value's kind:", type_slice.Kind(), ", name:", type_slice.Name())
	// var type_element = type_slice.Elem()
	fmt.Println("element's kind:", type_element.Kind(), ", name:", type_element.Name())

	// slice := reflect.New(type_slice)
	// slice := ss.Interface()
	// fmt.Print(vx)

	// element_1 := reflect.New(type_element)
	// element_2 := element_1.Interface()

	newMap := reflect.MakeMap(type_map)
	slice := reflect.MakeSlice(type_slice, 0, 16)
	newMap.SetMapIndex(reflect.ValueOf("2022-01-01"), slice)
	map_pointer.Set(newMap)

}

func TestPipe(t *testing.T) {
	// reflectValue := reflect.Indirect(reflect.ValueOf(value))
	// reflectValue.Slice(i, ends).Interface()
	pipeReader, pipeWriter := io.Pipe()
	rr := bufio.NewReader(io.Reader(pipeReader))
	ww := bufio.NewWriter(io.Writer(pipeWriter))
	v := make([]byte, 5)
	v[0] = 8
	v[1] = 16
	v[2] = 32
	v[3] = 64
	v[4] = 128

	ww.Write(v)

	ww.WriteString("Hello,超级棒")

	// binary.Write(ww, binary.LittleEndian, uint32(0))

	size := rr.Size()
	buffer := make([]byte, size)
	rr.Read(buffer)

}
