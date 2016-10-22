// 收集股票的基本信息
package collect

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/axgle/mahonia"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm"
	"github.com/myself659/gostocks/csv"
	"github.com/myself659/gostocks/db"
	"golang.org/x/net/html"
	_ "io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	_ "os"
	"strconv"
	"strings"
	"time"
)

type qyitem struct {
	Code       string    `gorm:"type:varchar(64)"`  // 股票代号
	Name       string    `gorm:"type:varchar(128)"` // 股票企业名称
	MV         float64   // 以亿元为单位,市值
	Revenue    float64   // 以亿元为单位，收入
	RevenueG   float64   //营收同比增长
	Profits    float64   // 以亿元为单位，纯利润
	ProfitsG   float64   // 纯利润同比增长
	PE         float64   // 市盈率
	PB         float64   // 市净率
	GP         float64   // 毛利率
	NP         float64   // 净利率
	ROE        float64   // 净资产收益率
	BVPS       float64   // 每股资资产
	date       time.Time // 收集时间
	LEV        float64   // 负债率
	URL        string    // 对应url地址
	MarketTime string    //上市时间
}

var title = []string{
	"股票代码",
	"企业名字",
	"市值",
	"营收",
	"营收同比增长",
	"纯利润",
	"纯利润同比增长",
	"市盈率",
	"市净率",
	"毛利率",
	"净利率",
	"净资产收益率",
	"每股资资产",
	"负债率",
	"URL",
}

var entitle = []string{
	"Code",
	"Name",
	"MV",
	"Revenue",
	"RevenueG",
	"Profits",
	"ProfitsG",
	"PE",
	"PB",
	"GP",
	"NP",
	"ROE",
	"BVPS",
	"LEV",
	"URL",
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

type Pos struct {
	x int
	y int
}

var PEPos = Pos{x: 0, y: 1}
var BVPSPos = Pos{x: 1, y: 0}
var PBPos = Pos{x: 1, y: 1}
var RevenuePos = Pos{x: 2, y: 0}
var RevenueGPos = Pos{x: 2, y: 1}
var ProfitsPos = Pos{x: 3, y: 0}
var ProfitsGPos = Pos{x: 3, y: 1}
var GPPos = Pos{x: 4, y: 0}
var NPPos = Pos{x: 4, y: 1}
var ROEPos = Pos{x: 5, y: 0}
var LEVPos = Pos{x: 5, y: 1}
var MVPos = Pos{x: 6, y: 1}

func fix(fv float64, a float64) float64 {
	fr := math.Floor((fv / a)) * a
	return fr
}

func saveName(page *html.Tokenizer, ptoken *html.Token, item *qyitem) bool {

	if ptoken.Type == html.StartTagToken && ptoken.Data == "h2" {
		for _, attr := range ptoken.Attr {
			if attr.Key == "class" && attr.Val == "header-title-h2 fl" {
				tokentype := page.Next()
				tokentype = tokentype
				token := page.Token()

				item.Name = token.Data
				return true
			}
		}
	}

	return false
}

func saveCode(page *html.Tokenizer, ptoken *html.Token, item *qyitem) bool {

	if ptoken.Type == html.StartTagToken && ptoken.Data == "b" {
		for _, attr := range ptoken.Attr {
			if attr.Key == "class" && attr.Val == "header-title-c fl" {
				tokentype := page.Next()
				tokentype = tokentype
				token := page.Token()
				enc := mahonia.NewEncoder("gbk")
				s := strings.TrimSpace(token.Data)
				s = enc.ConvertString(s)
				item.Code = s
				return true
			}
		}
	}

	return false
}

func savePE(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 3; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	fpe, err := strconv.ParseFloat(strings.TrimSpace(token.Data), 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("savePE:", fpe)
	item.PE = fix(fpe, 0.001)

	return true

}
func saveBVPS(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 4; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	enc := mahonia.NewEncoder("gbk")
	s := strings.TrimSpace(token.Data)
	cs := strings.TrimPrefix(s, enc.ConvertString("："))
	f, err := strconv.ParseFloat(cs, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveBVPS:", f)
	item.BVPS = fix(f, 0.001)

	return true
}

func savePB(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 3; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	f, err := strconv.ParseFloat(strings.TrimSpace(token.Data), 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("savePB:", f)
	item.PB = fix(f, 0.001)

	return true

}

func saveRevenue(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 1; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	var factor float64 = 1.0
	var fs string
	// 去掉中文

	if strings.HasSuffix(as[1], enc.ConvertString("亿")) {
		fs = strings.TrimSuffix(as[1], enc.ConvertString("亿"))

	} else if strings.HasSuffix(as[1], enc.ConvertString("万")) {
		factor = 0.0001
		fs = strings.TrimSuffix(as[1], enc.ConvertString("万"))
	} else {
		factor = 0.00000001
		fs = as[1]
	}
	f, err := strconv.ParseFloat(fs, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveRevenue:", f)
	item.Revenue = fix(f*factor, 0.001)

	return true
}

func saveRevenueG(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 4; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	enc := mahonia.NewEncoder("gbk")
	fs := strings.TrimPrefix(s, enc.ConvertString("："))
	ffs := strings.TrimSuffix(fs, "%")
	f, err := strconv.ParseFloat(ffs, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveRevenueG:", f)
	item.RevenueG = fix(f, 0.001)

	return true

}

func saveProfits(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 1; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	var factor float64 = 1.0
	var fs string

	if strings.HasSuffix(as[1], enc.ConvertString("亿")) {
		fs = strings.TrimSuffix(as[1], enc.ConvertString("亿"))

	} else if strings.HasSuffix(as[1], enc.ConvertString("万")) {
		factor = 0.0001
		fs = strings.TrimSuffix(as[1], enc.ConvertString("万"))
	} else {
		factor = 0.00000001
		fs = as[1]
	}

	f, err := strconv.ParseFloat(fs, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveProfits:", f)
	item.Profits = fix(f*factor, 0.001)

	return true
}

func saveProfitsG(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 1; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	enc := mahonia.NewEncoder("gbk")
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	sf := strings.TrimSuffix(as[1], "%")
	f, err := strconv.ParseFloat(sf, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveProfitsG:", f)
	item.ProfitsG = fix(f, 0.001)

	return true

}

func saveGP(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 4; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	token := page.Token()
	enc := mahonia.NewEncoder("gbk")
	s := strings.TrimSpace(token.Data)
	s = strings.TrimPrefix(s, enc.ConvertString("："))
	s = strings.TrimSuffix(s, "%")
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveGP:", f)
	item.GP = fix(f, 0.001)
	return true
}

func saveNP(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 1; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	sf := strings.TrimSuffix(as[1], "%")
	f, err := strconv.ParseFloat(sf, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveNP:", f)
	item.NP = fix(f, 0.001)

	return true

}
func saveROE(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 6; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	sf := strings.TrimSuffix(as[1], "%")
	f, err := strconv.ParseFloat(sf, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveROE:", f)
	item.ROE = fix(f, 0.001)
	return true
}

func saveLEV(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 1; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)
	as := strings.Split(s, enc.ConvertString("："))
	if len(as) != 2 {
		return false
	}
	sf := strings.TrimSuffix(as[1], "%")
	f, err := strconv.ParseFloat(sf, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveLEV:", f)
	item.LEV = fix(f, 0.001)
	return true
}

func saveMV(page *html.Tokenizer, item *qyitem) bool {
	for i := 0; i < 3; i++ {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return false
		}
	}
	enc := mahonia.NewEncoder("gbk")
	token := page.Token()
	s := strings.TrimSpace(token.Data)

	var factor float64 = 1.0
	var fs string
	// 去掉中文
	if strings.HasSuffix(s, enc.ConvertString("亿")) {
		fs = strings.TrimSuffix(s, enc.ConvertString("亿"))

	} else if strings.HasSuffix(s, enc.ConvertString("万")) {
		factor = 0.0001
		fs = strings.TrimSuffix(s, enc.ConvertString("万"))
	} else {
		factor = 0.00000001
		fs = s
	}

	f, err := strconv.ParseFloat(fs, 32)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("saveMV:", f)
	item.MV = fix(f*factor, 0.001)

	return true
}

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

func getNextTable(page *html.Tokenizer) *html.Token {
	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTable(token) {
			return &token
		}
	}
}

func isTableTr(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "tr" {
		return true
	}
	return false
}

func getNextTableTr(page *html.Tokenizer) *html.Token {

	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTableTr(token) {
			return &token
		}
	}

}

func isTableTd(token html.Token) bool {
	if token.Type == html.StartTagToken && token.Data == "td" {
		return true
	}
	return false

}

func getNextTableTd(page *html.Tokenizer) *html.Token {

	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return nil
		}
		token := page.Token()

		if isTableTd(token) {
			return &token
		}
	}

}

func isEndDiv(token html.Token) bool {
	if token.Type == html.EndTagToken && token.Data == "div" {
		return true
	}
	return false
}

func getQYItem(url string, impl db.Impl, csvi csv.Impl) {
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
	var item qyitem
	var ntr int = 0
	var temppos Pos
	var hasSaveName bool = false
	var hasSaveCode bool = false
	for {
		tokentype := page.Next()
		if tokentype == html.ErrorToken {
			return
		}

		token := page.Token()
		if isTable(token) {
			for {
				token := getNextTableTr(page)
				if token == nil {
					return
				}

				for i := 0; i < 2; i++ {
					token := getNextTableTd(page)
					if token == nil {
						return
					}
					temppos.x = ntr
					temppos.y = i
					switch temppos {
					case PEPos:
						{
							savePE(page, &item)
						}
					case PBPos:
						{
							savePB(page, &item)
						}
					case BVPSPos:
						{
							saveBVPS(page, &item)
						}
					case RevenuePos:
						{
							saveRevenue(page, &item)
						}
					case RevenueGPos:
						{
							saveRevenueG(page, &item)
						}
					case ProfitsPos:
						{
							saveProfits(page, &item)
						}
					case ProfitsGPos:
						{
							saveProfitsG(page, &item)
						}
					case GPPos:
						{
							saveGP(page, &item)
						}
					case NPPos:
						{
							saveNP(page, &item)
						}
					case ROEPos:
						{
							saveROE(page, &item)
						}
					case LEVPos:
						{
							saveLEV(page, &item)
						}
					case MVPos:
						{
							saveMV(page, &item)
							item.URL = url
							impl.DB.Save(item)
							csvi.Write(item)
							return
						}
					}

				}
				ntr++

			}

		} else {
			if hasSaveName == false {
				hasSaveCode = saveName(page, &token, &item)
			}
			if hasSaveCode == false {
				hasSaveCode = saveCode(page, &token, &item)
			}
		}
		/*
			fmt.Println("start #####")
			fmt.Println("Type:", token.Type)
			fmt.Println("DataAtom:", token.DataAtom)
			fmt.Println("Data:", token.Data)
			for _, attr := range token.Attr {
				fmt.Println("key:", attr.Key, "value:", attr.Val)
			}
			fmt.Println("end -------")
		*/

	}
}

func Run() {
	// 也不用高级，一个个来就行了
	ok, impl := db.Init("mysql", "root", "passwd", "127.0.0.1:3306", "test3")
	if ok == false {
		log.Fatal("db init failed")
		return
	}
	db.InitSchema(impl, &qyitem{})

	csvi := csv.NewCsv("stock.csv")
	csvi.Init(entitle)
	// 主板
	for i := 600000; i < 609000; i++ {
		url := preUrl + "sh" + strconv.Itoa(i) + ".html"
		getQYItem(url, impl, csvi)
	}

	// 中小板
	for i := 0; i < 10000; i++ {
		url := preUrl + "sz" + fmt.Sprintf("%06d", i) + ".html"
		getQYItem(url, impl, csvi)
	}

	//创业板
	for i := 300000; i < 309000; i++ {
		url := preUrl + "sz" + strconv.Itoa(i) + ".html"
		getQYItem(url, impl, csvi)
	}
	csvi.Close()

}
