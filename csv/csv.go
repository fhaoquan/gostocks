package csv

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	_ "strconv"
	"strings"
)

type Impl struct {
	writer *csv.Writer
	file   *os.File
}

func (i *Impl) Close() {
	i.file.Close()
}

//  只支持简单的struct类型，不支持内部成员有map等struct类型
func (i *Impl) Write(item interface{}) {

	//rtype := reflect.TypeOf(item)
	//rk := rtype.Kind()
	rv := reflect.ValueOf(item)
	n := rv.NumField()
	ss := make([]string, n)
	pos := 0
	for i := 0; i < n; i++ {
		rfield := rv.Field(i)
		rft := rfield.Type()
		switch rft.Kind() {
		case reflect.Float64:
			{
				fv := rfield.Float()
				//sv := strconv.FormatFloat(fv, 'E', -1, 64)
				sv := fmt.Sprintf("%.2f", fv)
				ss[pos] = sv
				pos++
			}
		case reflect.String:
			{
				sv := rfield.String()
				sv = strings.TrimSpace(sv)
				ss[pos] = sv
				pos++
			}
			// 其他数据类型类似
		}
	}

	err := i.writer.Write(ss)
	checkError("csv write error:", err)

	i.writer.Flush()

}

func (i *Impl) Init(title []string) {
	err := i.writer.Write(title)
	checkError("csv write error:", err)

	i.writer.Flush()
}

func NewCsv(name string) Impl {
	file, err := os.Create(name)
	checkError("csv create file error:", err)

	writer := csv.NewWriter(file)

	return Impl{writer: writer, file: file}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
