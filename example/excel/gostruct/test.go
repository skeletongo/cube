// Code generated converter.
// DO NOT EDIT!
package gostruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

var _ = errors.New
var _ = fmt.Println
var _ = strconv.Itoa
var _ = strings.Split

type Test struct {
	Boolean    bool      // 布尔
	Name       string    // 字符串
	Int32      int32     // 整数32位
	Int64      int64     // 整数64位
	Int        int       // 整数
	Float32    float32   // 浮点数32位
	Float64    float64   // 浮点数64位
	Double     float64   // 浮点数64位
	ArrBool    []bool    // 布尔数组
	ArrString  []string  // 字符串数组
	ArrInt32   []int32   // 数字数组32
	ArrInt64   []int     // 数字数组64
	ArrInt     []int     // 数字数组
	ArrFloat32 []float32 // 浮点数数组32
	ArrFloat64 []float64 // 浮点数数组64
	ArrDouble  []float64 // 浮点数数组
}

var TestSingle = new(TestFile)

type TestFile struct {
	Array []*Test
}

func (t *TestFile) LoadJson(p string) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	t.Array = t.Array[:0]
	return json.Unmarshal(b, t)
}

func (t *TestFile) LoadXlsx(p string) error {
	f, err := xlsx.OpenFile(p)
	if err != nil {
		return err
	}

	if len(f.Sheets) == 0 {
		return nil
	}

	rows := f.Sheets[0].Rows
	if len(rows) < 3 {
		return nil
	}

	t.Array = t.Array[:0]

	var line *xlsx.Row
	for _, line = range rows[2:] {
		_ = line
		l := len(line.Cells)
		_ = l
		r := new(Test)
		for {
			if l < 0+1 {
				break
			}

			r.Boolean = line.Cells[0].Bool()

			if l < 1+1 {
				break
			}

			r.Name = line.Cells[1].String()

			if l < 2+1 {
				break
			}

			n2, err := line.Cells[2].Int()
			if err != nil {
				return errors.New(fmt.Sprintf("Int32 error:%v", err))
			}
			r.Int32 = int32(n2)
			if l < 3+1 {
				break
			}

			n3, err := line.Cells[3].Int()
			if err != nil {
				return errors.New(fmt.Sprintf("Int64 error:%v", err))
			}
			r.Int64 = int64(n3)

			if l < 4+1 {
				break
			}

			if r.Int, err = line.Cells[4].Int(); err != nil {
				return errors.New(fmt.Sprintf("Int error:%v", err))
			}

			if l < 5+1 {
				break
			}

			n5, err := line.Cells[5].Float()
			if err != nil {
				return errors.New(fmt.Sprintf("Float32 error:%v", err))
			}
			r.Float32 = float32(n5)

			if l < 6+1 {
				break
			}

			if r.Float64, err = line.Cells[6].Float(); err != nil {
				return errors.New(fmt.Sprintf("Float64 error:%v", err))
			}

			if l < 7+1 {
				break
			}

			if r.Double, err = line.Cells[7].Float(); err != nil {
				return errors.New(fmt.Sprintf("Double error:%v", err))
			}

			if l < 8+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[8].String(), ",") {
				b, err := strconv.ParseBool(v)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrBool error:%v", err))
				}
				r.ArrBool = append(r.ArrBool, b)
			}

			if l < 9+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[9].String(), ",") {
				r.ArrString = append(r.ArrString, v)
			}

			if l < 10+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[10].String(), ",") {
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrInt32 error:%v", err))
				}
				r.ArrInt32 = append(r.ArrInt32, int32(i))
			}

			if l < 11+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[11].String(), ",") {
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrInt64 error:%v", err))
				}
				r.ArrInt64 = append(r.ArrInt64, int(i))
			}

			if l < 12+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[12].String(), ",") {
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrInt error:%v", err))
				}
				r.ArrInt = append(r.ArrInt, int(i))
			}

			if l < 13+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[13].String(), ",") {
				i, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrFloat32 error:%v", err))
				}
				r.ArrFloat32 = append(r.ArrFloat32, float32(i))
			}

			if l < 14+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[14].String(), ",") {
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrFloat64 error:%v", err))
				}
				r.ArrFloat64 = append(r.ArrFloat64, i)
			}

			if l < 15+1 {
				break
			}

			for _, v := range strings.Split(line.Cells[15].String(), ",") {
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return errors.New(fmt.Sprintf("ArrDouble error:%v", err))
				}
				r.ArrDouble = append(r.ArrDouble, i)
			}

			break
		}
		t.Array = append(t.Array, r)
	}
	return nil
}
