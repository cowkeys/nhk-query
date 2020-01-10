package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get("http://metalsucks.net")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find(".sidebar-reviews article .content-block").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		band := s.Find("a").Text()
		title := s.Find("i").Text()
		fmt.Printf("Review %d: %s - %s\n", i, band, title)
	})
}

var category map[string]string

func init() {
	category = make(map[string]string, 0)
	category["母婴专区"] = "http://www.wellcome.co.nz/gallery-267--6--%v--grid.html"
	category["营养保健"] = "http://www.wellcome.co.nz/gallery-270--6--%v--grid.html"
	category["美妆个护"] = "http://www.wellcome.co.nz/gallery-271--6--%v--grid.html"
	category["美食特产"] = "http://www.wellcome.co.nz/gallery-272--6--%v--grid.html"
	/*category["母婴"] = "http://www.wellcome.co.nz/gallery-270--6--%v--grid.html"
	category["母婴"] = "http://www.wellcome.co.nz/gallery-267--6--1--grid.html"
	category["母婴"] = "http://www.wellcome.co.nz/gallery-267--6--1--grid.html"*/
}

func Do(url string, cate string) ([]*Item, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("fetch url :%v status code error: %d %s \n", url, res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Do load html error %v", err))
	}

	items := make([]*Item, 0)

	// Find the review items
	doc.Find(".GoodsSearchWrap .items-gallery").Each(func(i int, s *goquery.Selection) {
		item := &Item{
			Category: cate,
		}
		// For each item found, get the band and title
		s.Find(".entry-content tr").Each(func(j int, ss *goquery.Selection) {
			//fmt.Println(j, ss.Find("a").Text())
			if j == 0 {
				item.Name = ss.Find("a").Text()
				item.IndexHtml, _ = ss.Find("a").Attr("href")
			}
			if j == 1 {
				item.Price = ss.Find(".price1").Text()
				item.Sale = ss.Find("font").Text()
			}
			items = append(items, item)
		})
	})
	return items, nil
}

type Item struct {
	Category    string
	IndexHtml   string
	SubCategory string
	Name        string
	Price       string
	Sale        string
}

func QueryNHK(cate string, num int) []*Item {
	wg := &sync.WaitGroup{}
	items := make([]*Item, 0)
	for i := 1; i <= num; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			path := fmt.Sprintf(category[cate], index)
			list, err := Do(path, cate)
			if err != nil {
				fmt.Printf("fetch %v err: %v \n", path, err)
				return
			}
			if list == nil || len(list) == 0 {
				return
			}
			items = append(items, list...)
		}(i)
	}

	wg.Wait()
	return items
}

func FetchData() {
	list1 := QueryNHK("母婴专区", 30)

	list2 := QueryNHK("营养保健", 30)
	list3 := QueryNHK("美妆个护", 30)
	list4 := QueryNHK("美食特产", 30)

	list := append(list1, list2...)
	list = append(list, list3...)
	list = append(list, list4...)

	b, _ := json.Marshal(list)
	fmt.Println(string(b))
}

func main() {
	// FetchData()
	Load("./datafilter.json")

}

func Load(filename string) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	var list []*Item
	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, &list)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	parsePrice := func(p string) string {
		a := strings.Split(p, "NZ$")
		v1, _ := strconv.ParseFloat(a[1], 32)
		return fmt.Sprintf("%.2f", v1*4.58)
	}
	//fmt.Println("lll", len(list))
	result := make(map[string]*Item, 0)
	index := 1
	for _, v := range list {
		_, ok := result[v.Name]
		if ok {
			continue
		}
		result[v.Name] = v
		fmt.Printf("%v|%v|%v|%v|%v\n", index, v.Category, v.Name, parsePrice(v.Price), v.Sale)
		index++
	}
}
