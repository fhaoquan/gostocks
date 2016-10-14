// 收集股票的基本信息
package collect

import (
	"bytes"
	"crypto/tls"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm"
	"github.com/myself659/gostocks/db"
	"golang.org/x/net/html"
	_ "io"
	"io/ioutil"
	"log"
	"net/http"
	_ "os"
	"strconv"
	"strings"
	"time"
)

type qyitem struct {
	Num      string    `gorm:"type:varchar(64)"`  // 股票代号
	Name     string    `gorm:"type:varchar(128)"` // 股票企业名称
	MV       float64   // 以亿元为单位,市值
	Revenue  float64   // 以亿元为单位，收入
	RevenueG float64   //营收同比增长
	Profits  float64   // 以亿元为单位，纯利润
	ProfitsG float64   // 纯利润同比增长
	PE       float64   // 市盈率
	PB       float64   // 市净率
	GP       float64   // 毛利率
	NP       float64   // 净利率
	ROE      float64   // 净资产收益率
	BVPS     float64   // 每股资资产
	date     time.Time // 收集时间
	LEV      float64   // 负债率
}

const (
	PETag       = "PE(动)："
	BVPSTag     = "净资产"
	RevenueTag  = "营收"
	ProfitsTag  = "净利润"
	ProfitsGTag = "同比："
	GPTag       = "毛利率"
	NPTag       = "净利率"
	LEVTag      = "负债率"
	MVTag       = "总值"
	RevenueGTag = "同比"
	PBTag       = "市净率："
)

// 获取主板，中小板，创业板的股票的基本情况
const preUrl = "http://quote.eastmoney.com/"

func isStartMb10(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "div" {
		for _, attr := range token.Attr {
			if attr.Key == "class" && attr.Val == "box-x1 mb10" {
				return true
			}
		}
	}

	return false
}

// 找到更合理的定位方法,直接解析表格
func isTable(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "table" {
		for _, attr := range token.Attr {
			if attr.Key == "class" && attr.Val == "line23 w100p text-indent3 bt txtUL" {
				return true
			}
		}
	}

	return false
}

func getNextTable(page *html.Tokenizer) html.Token {
	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTable(token) {
			return token
		}
	}
}

func isTableTr(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "tr" {
		return true
	}
	return false
}

func getNextTableTr(page *html.Tokenizer) html.Token {

	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTableTr(token) {
			return token
		}
	}

}

func isTableTd(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "td" {
		return true
	}
	return false

}

func getNextTableTd(page *html.Tokenizer) html.Token {

	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTableTd(token) {
			return token
		}
	}

}

func isEndDiv(token html.Token) bool {
	if token.Type == html.EndTagToken && token.Data == "div" {
		return true
	}
	return false
}

func getPE(page *html.Tokenizer) (bool, float64) {

	for i := 0; i < 2; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false, 0.0
		}
	}
	token := page.Token()
	fpe, err := strconv.ParseFloat(strings.TrimSpace(token.Data), 32)
	if err != nil {
		fmt.Println(err)
		return false, 0.0
	}
	return true, fpe
}

func getPB(page *html.Tokenizer) (bool, float64) {

	for i := 0; i < 2; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false, 0.0
		}
	}
	token := page.Token()
	fpe, err := strconv.ParseFloat(strings.TrimSpace(token.Data), 32)
	if err != nil {
		fmt.Println(err)
		return false, 0.0
	}
	return true, fpe
}

func getBVPS(page *html.Tokenizer) (bool, float64) {
	for i := 0; i < 2; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false, 0.0
		}
	}
	token := page.Token()
	fpe, err := strconv.ParseFloat(strings.TrimSpace(token.Data), 32)
	if err != nil {
		fmt.Println(err)
		return false, 0.0
	}
	return true, fpe

}

func getQYItem(url string, impl db.Impl) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(url, err)
		return
	}

	breader := bytes.NewReader(b)

	defer resp.Body.Close()

	page := html.NewTokenizer(breader)
	var num_mb10 int = 0
	var item qyitem
	var ntr int = 0
	var ntd int = 0
	fmt.Println(PETag)
	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return
		}
		token := page.Token()

		if isTable(token) {

			if isTableTr(token) {
				ntr++
				ntd = 0
			}
			if ifTableTd(token) {

				// 进行处理
				ntd++

			}

		}
		/*
			fmt.Println("start -------")
			fmt.Println("Type:", token.Type)
			fmt.Println("DataAtom:", token.DataAtom)
			fmt.Println("Data:", token.Data)
			for _, attr := range token.Attr {
				fmt.Println("key:", attr.Key, "value:", attr.Val)
			}
			fmt.Println("end -------")
		*/
		/*
			if isStartMb10(token) {
				num_mb10++
				fmt.Println(num_mb10)
			}
			//定位数据
			if num_mb10 == 2 {
				// 更好地解析

				if token.Type == html.TextToken {
					fmt.Println(token.Data)
				}
				data := strings.TrimSpace(token.Data)
				if PETag == data {
					fmt.Println("match")
					ok, fpe := getPE(page)
					if !ok {
						return
					}

					item.PE = fpe
					fmt.Println(fpe)
				}
				if strings.Contains(data, PETag) {
					//获得pe
					ok, fpe := getPE(page)
					if !ok {
						return
					}

					item.PE = fpe
					fmt.Println(fpe)
					continue
				}

				if strings.Contains(data, PBTag) {
					//获得PB
					ok, fpb := getPB(page)
					if !ok {
						return
					}

					item.PE = fpb
					fmt.Println(fpb)
					continue
				}

				if strings.Contains(data, BVPSTag) {
					ok, fbvps := getBVPS(page)
					if !ok {
						return
					}
					item.BVPS = fbvps
				}

			}
		*/
	}
}

func Run() {
	// 也不用高级，一个个来就行了
	ok, i := db.Init("mysql", "root", "dbstar", "127.0.0.1:3306", "test3")
	if ok == false {
		log.Fatal("db init failed")
		return
	}
	db.InitSchema(i, &qyitem{})

	url := preUrl + "sz300017.html"

	getQYItem(url, i)

}

/*
数据格式
*/

/*
start -------
Type: StartTag
DataAtom: div
Data: div
key: class value: box-x1 mb10
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: div
Data: div
key: class value: pad5
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: table
Data: table
key: class value: line23 w100p text-indent3 bt txtUL
key: id value: rtp2
key: cellspacing value: 0
key: cellpadding value: 0
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tbody
Data: tbody
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: StartTag
DataAtom: a
Data: a
key: href value: http://data.eastmoney.com/bbsj/300017.html
key: target value: _blank
end -------
start -------
Type: Text
DataAtom:
Data: 收益
end -------
start -------
Type: EndTag
DataAtom: a
Data: a
end -------
start -------
Type: Text
DataAtom:
Data: (
end -------
start -------
Type: StartTag
DataAtom: span
Data: span
key: title value: 第二季度
end -------
start -------
Type: Text
DataAtom:
Data: 二
end -------
start -------
Type: EndTag
DataAtom: span
Data: span
end -------
start -------
Type: Text
DataAtom:
Data: )：0.733
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------

start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: PE(动)： # 向右移动两个
end -------

start -------
Type: StartTag
DataAtom: span
Data: span
key: id value: gt6_2
end -------

start -------
Type: Text
DataAtom:
Data: 47.97
end -------


start -------
Type: EndTag
DataAtom: span
Data: span
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------

start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------

start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: StartTag
DataAtom: a
Data: a
key: href value: http://data.eastmoney.com/bbsj/300017.html
key: target value: _blank
end -------
start -------
Type: Text
DataAtom:
Data: 净资产  #向后移动两个token
end -------

start -------
Type: EndTag
DataAtom: a
Data: a
end -------

start -------
Type: Text
DataAtom:
Data: ：8.212
end -------

start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------

start -------
Type: Text
DataAtom:
Data: 市净率：  #向后移动两个token
end -------

start -------
Type: StartTag
DataAtom: span
Data: span
key: id value: gt13_2
end -------

start -------
Type: Text
DataAtom:
Data: 8.56
end -------


start -------
Type: EndTag
DataAtom: span
Data: span
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 营收：20.56亿
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: StartTag
DataAtom: a
Data: a
key: href value: http://data.eastmoney.com/bbsj/300017.html
key: target value: _blank
end -------
start -------
Type: Text
DataAtom:
Data: 同比
end -------
start -------
Type: EndTag
DataAtom: a
Data: a
end -------
start -------
Type: Text
DataAtom:
Data: ：66.24%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 净利润：5.86亿
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 同比：81.78%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: StartTag
DataAtom: a
Data: a
key: href value: http://data.eastmoney.com/bbsj/300017.html
key: target value: _blank
end -------
start -------
Type: Text
DataAtom:
Data: 毛利率
end -------
start -------
Type: EndTag
DataAtom: a
Data: a
end -------
start -------
Type: Text
DataAtom:
Data: ：43.61%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 净利率：28.45%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: StartTag
DataAtom: a
Data: a
key: href value: http://data.eastmoney.com/bbsj/300017.html
key: target value: _blank
end -------
start -------
Type: Text
DataAtom:
Data: ROE
end -------
start -------
Type: StartTag
DataAtom: b
Data: b
key: title value: 净资产收益率
key: class value: hxsjccsyl
end -------
start -------
Type: EndTag
DataAtom: b
Data: b
end -------
start -------
Type: EndTag
DataAtom: a
Data: a
end -------
start -------
Type: Text
DataAtom:
Data: ：12.81%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 负债率：12.80%
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
key: title value: 7.99亿
end -------
start -------
Type: Text
DataAtom:
Data: 总股本：7.99亿
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 总值：
end -------
start -------
Type: StartTag
DataAtom: span
Data: span
key: id value: gt7_2
end -------
start -------
Type: Text
DataAtom:
Data: 562.0亿
end -------
start -------
Type: EndTag
DataAtom: span
Data: span
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
key: title value: 5.12亿
end -------
start -------
Type: Text
DataAtom:
Data: 流通股：5.12亿
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data: 流值：
end -------
start -------
Type: StartTag
DataAtom: span
Data: span
key: id value: gt14_2
end -------
start -------
Type: Text
DataAtom:
Data: 360.1亿
end -------
start -------
Type: EndTag
DataAtom: span
Data: span
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
key: colspan value: 2
end -------
start -------
Type: Text
DataAtom:
Data: 每股未分配利润：2.387元
end -------
start -------
Type: EndTag
DataAtom: td
Data: td
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: EndTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: tr
Data: tr
end -------
start -------
Type: Text
DataAtom:
Data:

end -------
start -------
Type: StartTag
DataAtom: td
Data: td
key: colspan value: 2
key: class value: pb3
end -------
start -------
Type: Text
DataAtom:
Data: 上市时间：2009-10-30
end -------
*/
