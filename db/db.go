package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
)

type Impl struct {
	DB *gorm.DB
}

func Init(dbtype string, user string, psword string, ipport string, dbname string) (bool, Impl) {
	i := Impl{}
	var err error
	para := user + ":" + psword + "@tcp" + "(" + ipport + ")" + "/" + dbname + "?charset=utf8&parseTime=True"
	i.DB, err = gorm.Open(dbtype, para)
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
		return false, i
	}

	return true, i

}

func InitSchema(i Impl, pitem interface{}) {
	i.DB.AutoMigrate(pitem)
}

func SaveItem(i Impl, item interface{}) {
	i.DB.Save(item)
}
