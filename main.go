package main

import (
	 "encoding/json"
	"fmt"
	 "os"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gocolly/colly"
	_ "github.com/go-sql-driver/mysql"
)

type news struct {
	News_headline string `json:"news_headline"`
	News_text     string `json:"news_text"`
	News_time     string `json:"news_time"`
	URL           string `json:"url"`
}

func setupRouter(router *mux.Router) {
	router.
		Methods("POST").
		Path("/endpoint").
		HandlerFunc(postFunction)
}

func postFunction(w http.ResponseWriter, r *http.Request) {
	log.Println("You called a thing!")
}

func main() {
	/////DB
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/parser_news")
	if err != nil {
		panic(err)
	}
	///
	result, err := db.Query("SELECT * FROM news")
	if err != nil {
        panic(err)
    }
	for result.Next() {
		var id int
		var headline string
		var text string
		var time string
		var url string


		err = result.Scan(&id, &headline, &text, &time, &url)
		if err != nil {
            panic(err)
        }
         
        log.Printf("id: %d\n headline: %s\n text: %s\n time: %s\n url: %s\n ", id, headline, text, time, url)
    }

	defer db.Close()

	router := mux.NewRouter().StrictSlash(true)

	setupRouter(router)

	log.Fatal(http.ListenAndServe(":8080", router))

	c := colly.NewCollector(
		colly.AllowedDomains("acc.cv.ua"),
	)
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Scraping:", r.URL)
	})

	var allNews []news

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Status:", r.StatusCode)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "nError:", err)
	})

	// c.OnHTML("a.AllNewsItemInfo__name", func(h *colly.HTMLElement) {
	// 	fmt.Println(h.Text)
	// })

	c.OnHTML("div.AllNewsItem", func(h *colly.HTMLElement) {
		news := news{
			News_headline: h.ChildText("a.AllNewsItemInfo__name"),
			News_text:     h.ChildText("a.AllNewsItemInfo__desc"),
			News_time:     h.ChildText("div.AllNewsItemService__date"),
			URL:           h.ChildAttr("a.AllNewsItemInfo__name", "href"),
		}
		//
		allNews = append(allNews, news)
		// fmt.Println(allNews)
		//fmt.Println(news.URL)
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `news` (`headline`, `text`, `time`, `url`) VALUES('%s', '%s', '%s', '%s')", news.News_headline, news.News_text, news.News_time, news.URL))
	if err != nil {
		panic(err)
	}
	
	defer insert.Close()
	})

	c.Visit("https://acc.cv.ua/news/chernivtsi/")
	content, err := json.Marshal(allNews)

	if err != nil {
		fmt.Println(err.Error())
	}
	
	os.WriteFile("News.json", content, 0644)
	fmt.Println(len(allNews))

	c.OnHTML("a", func(p *colly.HTMLElement) {
		nextPage := p.Request.AbsoluteURL(p.Attr("data-href"))
		c.Visit(nextPage)
	})

}