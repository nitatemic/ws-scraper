// Copyright © 2024
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fetch retrieves desired information from the [en.]ws-tcg.com websites.
package fetch

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/publicsuffix"

	"github.com/Akenaide/biri"
	"github.com/PuerkitoBio/goquery"
)

// The maximum number of workers at each stage that can do tasks locally
// (that don't have to interact with the websites).
const maxLocalWorker int = 10
// The maximum number of workers at each stage that have to interact with the websites.
const maxScrapeWorker int = 5

type SiteLanguage string

const (
	En SiteLanguage = "EN"
	Jp SiteLanguage = "JP"
)

type siteConfig struct {
	baseURL                    string
	cardListURL                string
	cardSearchURL              string
	languageCode               string
	lastPageFunc               func(doc *goquery.Document) int
	pageScanParseFunc          func(task *scrapeTask, wgCardSel *sync.WaitGroup, cardSelCh chan<- *goquery.Selection, resp *http.Response) (pageDone bool)
	recentReleaseDistinguisher string
	recentRelaseExpansionFunc  func(page *goquery.Selection) *url.Values
	supportTitleNumber         bool
}

var siteConfigs = map[SiteLanguage]siteConfig{
	En: {
		baseURL:       "https://en.ws-tcg.com/",
		cardListURL:   "https://en.ws-tcg.com/cardlist/list/",
		cardSearchURL: "https://en.ws-tcg.com/cardlist/searchresults/",
		languageCode:  "EN",
		lastPageFunc: func(doc *goquery.Document) int {
			numCardsS := doc.Find(".c-search__results-item span").First().Text()
			numCards, err := strconv.Atoi(numCardsS)
			if err != nil {
				log.Fatalf("Couldn't get num cards: %v\n", err)
				return 1
			}
			// As of 2024-9-3, there are 15 cards per "page".
			// TODO: figure out a better way to get the total number of pages
			return (numCards-1)/15 + 1
		},
		pageScanParseFunc: func(task *scrapeTask, wgCardSel *sync.WaitGroup, cardSelCh chan<- *goquery.Selection, resp *http.Response) (pageDone bool) {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				task.pageURLCh <- resp.Request.URL.String()
				log.Println("goquery error: ", err, "for page: ", resp.Request.URL)
				return false
			}
			resultList := doc.Find(".p_cards__results-box ul li")

			if resultList.Length() == 0 && resp.StatusCode == http.StatusOK {
				log.Println("No cards on response page")
			} else {
				log.Println("Found cards !!", resp.Request.URL)
				resultList.Each(func(i int, s *goquery.Selection) {
					subPath, exists := s.Find("a").First().Attr("href")
					if !exists {
						log.Printf("Error getting sub path: %v\n", err)
						return
					}
					fp, err := joinPath(task.siteConfig.baseURL, subPath)
					if err != nil {
						log.Printf("Error getting full path: %v\n", err)
						return
					}
					fullPath := fp.String()

					proxy := biri.GetClient()
					proxy.Client.Jar = task.cookieJar
					detailedPageResp, err := proxy.Client.Get(fullPath)
					if err != nil || detailedPageResp.StatusCode != http.StatusOK {
						var sc string
						if detailedPageResp != nil {
							sc = fmt.Sprintf(" (statusCode=%d)", detailedPageResp.StatusCode)
						}
						log.Printf("Failed to get detailed page from %q%s: %v\n", fullPath, sc, err)
					} else {
						proxy.Readd()
						doc, err := goquery.NewDocumentFromReader(detailedPageResp.Body)
						if err != nil {
							task.pageURLCh <- detailedPageResp.Request.URL.String()
							log.Println("goquery error: ", err, "for page: ", detailedPageResp.Request.URL)
							return
						}
						cardDetails := doc.Find(".p-cards__detail-wrapper")
						wgCardSel.Add(1)
						cardSelCh <- cardDetails
					}
				})
			}

			return true
		},
		recentReleaseDistinguisher: "div.p-cards__latest-products ul.c-product__list a",
		recentRelaseExpansionFunc: func(sel *goquery.Selection) *url.Values {
			if hrefAttr, exists := sel.Attr("href"); exists {
				re := regexp.MustCompile(`expansion=(\d+)`)
				if m := re.FindStringSubmatch(hrefAttr); m != nil {
					return &url.Values{
						"cmd":             {"search"},
						"show_page_count": {"100"},
						"show_small":      {"0"},
						"parallel":        {"0"},
						"expansion":       {m[1]},
					}
				}
			}
			return nil
		},
		supportTitleNumber: true,
	},
	Jp: {
		baseURL:       "https://ws-tcg.com/",
		cardListURL:   "https://ws-tcg.com/cardlist/",
		cardSearchURL: "https://ws-tcg.com/cardlist/search",
		languageCode:  "JP",
		lastPageFunc: func(doc *goquery.Document) int {
			all := doc.Find(".pager .next")

			last, _ := strconv.Atoi(all.Prev().First().Text())
			// default is 1, there no .pager .next if it's the only page
			if last == 0 {
				last = 1
			}
			return last
		},
		pageScanParseFunc: func(task *scrapeTask, wgCardSel *sync.WaitGroup, cardSelCh chan<- *goquery.Selection, resp *http.Response) (pageDone bool) {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				task.pageURLCh <- resp.Request.URL.String()
				log.Println("goquery error: ", err, "for page: ", resp.Request.URL)
				return false
			}
			resultTable := doc.Find(".search-result-table tr")

			if resultTable.Length() == 0 && resp.StatusCode == http.StatusOK {
				log.Println("No cards on response page")
			} else {
				log.Println("Found cards !!", resp.Request.URL)
				resultTable.Each(func(i int, s *goquery.Selection) {
					wgCardSel.Add(1)
					cardSelCh <- s
				})
			}

			return true
		},
		recentReleaseDistinguisher: "div.system > ul.expansion-list a[onclick]",
		recentRelaseExpansionFunc: func(sel *goquery.Selection) *url.Values {
			onclickAttr, exists := sel.Attr("onclick")
			if exists {
				// Extract the integer value from the onclick attribute
				parts := strings.Split(onclickAttr, "('")
				if len(parts) >= 2 {
					value := strings.TrimSuffix(parts[1], "')")
					return &url.Values{
						"cmd":             {"search"},
						"show_page_count": {"100"},
						"show_small":      {"0"},
						"parallel":        {"0"},
						"expansion":       {value},
					}
				}
			}
			return nil
		},
		supportTitleNumber: false,
	},
}

type Booster struct {
	SetCode string
	Cards   []Card
}

type scrapeTask struct {
	pageURLCh  chan string
	pageRespCh chan *http.Response
	siteConfig siteConfig
	urlValues  url.Values
	cookieJar  http.CookieJar
	lastPage   int
	wgPageScan *sync.WaitGroup
}

func (s *scrapeTask) getLastPage() int {
	log.Println(s.siteConfig.cardSearchURL, s.urlValues)
	resp, err := http.PostForm(fmt.Sprintf("%v?page=%d", s.siteConfig.cardSearchURL, 1), s.urlValues)
	if err != nil {
		log.Fatalf("Error on getting last page: %v\n", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error on getting last page parse: %v\n", err)
	}

	last := s.siteConfig.lastPageFunc(doc)

	log.Println("Last page is", last, " for :", s.urlValues)
	s.lastPage = last
	return last
}

func getTasksForRecentReleases(siteCfg siteConfig, doc *goquery.Document) []scrapeTask {
	var tasks []scrapeTask
	// Find all <a> elements with onclick attributes within the <ul> element
	doc.Find(siteCfg.recentReleaseDistinguisher).Each(func(i int, sel *goquery.Selection) {
		if v := siteCfg.recentRelaseExpansionFunc(sel); v != nil {

			tasks = append(tasks, scrapeTask{urlValues: *v})
		}
	})
	return tasks
}

func joinPath(baseURL, subPath string) (*url.URL, error) {
	b, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL: %v", err)
	}
	sp, err := url.Parse(subPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse sub path: %v", err)
	}
	return b.ResolveReference(sp), nil
}

func pageFetchWorker(id int, task *scrapeTask) {
	for link := range task.pageURLCh {
		log.Println("ID :", id, "Fetch page : ", link, "with params : ", task.urlValues)
		proxy := biri.GetClient()
		log.Println("Got proxy")
		proxy.Client.Jar = task.cookieJar

		resp, err := proxy.Client.PostForm(link, task.urlValues)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("Ban proxy:", err)
			proxy.Ban()
			// Retry again later
			task.pageURLCh <- link
		} else {
			proxy.Readd()
			task.pageRespCh <- resp
		}
	}
	log.Println("Page fetch worker", id, "done")
}

func pageScanWorker(
	id int,
	task *scrapeTask,
	wgCardSel *sync.WaitGroup,
	cardSelCh chan<- *goquery.Selection,
) {
	for resp := range task.pageRespCh {
		log.Printf("Start page: %v", resp.Request.URL)
		if task.siteConfig.pageScanParseFunc(task, wgCardSel, cardSelCh, resp) {
			task.wgPageScan.Done()
		}
		log.Printf("Finish page: %v", resp.Request.URL)
	}
	log.Println("Page scan worker", id, "done")
}

func extractWorker(siteCfg siteConfig, wgCardSel *sync.WaitGroup, cardSelChan <-chan *goquery.Selection, cardCh chan<- Card) {
	for s := range cardSelChan {
		c := extractData(siteCfg, s)
		cardCh <- c
		wgCardSel.Done()
	}
}

type reducer interface {
	reduce(config reducerConfig)
}

type reducerConfig struct {
	wg     *sync.WaitGroup
	cardCh chan Card
}

type cardListReducer struct {
	cards []Card
}

func (clr *cardListReducer) reduce(rc reducerConfig) {
	for c := range rc.cardCh {
		clr.cards = append(clr.cards, c)
	}
	rc.wg.Done()
}

type boosterReducer struct {
	boosterMap map[string]Booster
}

func (br *boosterReducer) reduce(rc reducerConfig) {
	for c := range rc.cardCh {
		boosterCode := c.Release
		boosterObj := br.boosterMap[boosterCode]

		boosterObj.Cards = append(boosterObj.Cards, c)
		br.boosterMap[boosterCode] = boosterObj
	}
	rc.wg.Done()
}

func prepareBiri(cfg siteConfig) {
	biri.Config.PingServer = cfg.baseURL
	biri.Config.TickMinuteDuration = 1
	biri.Config.Timeout = 25
}

type Config struct {
	// The website's internal code for each expansion. The value is language-specific.
	// For example,
	//   159 is "BanG Dream! Girls Band Party Premium Booster" in EN
	//   159 is "Monogatari Series: Second Season"
	ExpansionNumber int
	GetAllRarities  bool
	GetRecent       bool
	Language        SiteLanguage
	PageStart       int
	Reverse         bool
	SetCode         []string
	// The website's internal code for each set. The value is language-specific.
	// For example
	//   159 is "Tokyo Revengers" in EN
	//   159 isn't supported in JP
	TitleNumber int
}

func CardsStream(cfg Config, cardCh chan<- Card) error {
	var siteCfg siteConfig
	if c, ok := siteConfigs[cfg.Language]; !ok {
		log.Fatalf("Unsupported language: %q\n", cfg.Language)
	} else {
		siteCfg = c
		log.Printf("Fetching %s cards\n", cfg.Language)
	}

	log.Println("Streaming cards with config:", cfg)

	prepareBiri(siteCfg)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		err = fmt.Errorf("failed to get new cookiejar: %v", err)
		log.Fatal(err)
		return err
	}

	biri.ProxyStart()

	urlValues := url.Values{
		"cmd":             {"search"},
		"show_page_count": {"100"},
		"show_small":      {"0"},
	}
	if cfg.ExpansionNumber != 0 {
		switch cfg.Language {
		case En:
			// "expansion" also works, but the website uses "expansion_name", so use "expansion" to
			// stay in line with the website
			urlValues.Add("expansion_name", strconv.Itoa(cfg.ExpansionNumber))
		case Jp:
			urlValues.Add("expansion", strconv.Itoa(cfg.ExpansionNumber))
		}
	}
	if cfg.TitleNumber != 0 {
		if !siteCfg.supportTitleNumber {
			err := fmt.Errorf("can't use title number on %s site", cfg.Language)
			log.Fatalln(err)
			return err
		}
		urlValues.Add("title", strconv.Itoa(cfg.TitleNumber))
	}
	if cfg.GetAllRarities {
		urlValues.Add("parallel", "0")
	} else {
		urlValues.Add("parallel", "1")
	}
	if len(cfg.SetCode) > 0 {
		switch cfg.Language {
		case En:
			urlValues.Add("keyword_or", strings.Join(cfg.SetCode, " "))
			urlValues.Add("keyword_type[]", "no")
		case Jp:
			urlValues.Add("title_number", fmt.Sprintf("##%s##", strings.Join(cfg.SetCode, "##")))
		}
	}

	var scrapeTasks []*scrapeTask
	defaultScrapeTask := scrapeTask{
		cookieJar:  jar,
		siteConfig: siteCfg,
		urlValues:  urlValues,
	}
	if cfg.GetRecent {
		resp, err := http.Get(siteCfg.cardListURL)
		if err != nil {
			log.Fatal("Error on get recent")
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal("Error on parse recent")
		}
		for _, recent := range getTasksForRecentReleases(siteCfg, doc) {
			copyTask := defaultScrapeTask
			copyTask.urlValues = recent.urlValues
			log.Println(defaultScrapeTask, recent)
			scrapeTasks = append(scrapeTasks, &copyTask)
		}
	} else {
		scrapeTasks = append(scrapeTasks, &defaultScrapeTask)
	}

	loopNum := 0
	for _, st := range scrapeTasks {
		lastPage := st.getLastPage()
		loopNum += lastPage
		st.pageURLCh = make(chan string, lastPage)
		st.pageRespCh = make(chan *http.Response, maxScrapeWorker)
		st.wgPageScan = &sync.WaitGroup{}
		st.wgPageScan.Add(lastPage)
	}

	log.Printf("Number of loop %v\n", loopNum)

	var wgScanner, wgCardSel sync.WaitGroup
	cardSelCh := make(chan *goquery.Selection, maxLocalWorker)
	for i := 0; i < maxLocalWorker; i++ {
		go extractWorker(siteCfg, &wgCardSel, cardSelCh, cardCh)
	}
	for _, st := range scrapeTasks {
		wgScanner.Add(1)
		go func(s *scrapeTask) {
			// Wait for page scanning to finish instead of the fetch workers because
			// sometimes the scanners put work back in the fetch channel.
			s.wgPageScan.Wait()
			close(s.pageURLCh)
			close(s.pageRespCh)
			wgScanner.Done()
		}(st)
		for i := 0; i < maxScrapeWorker; i++ {
			go pageFetchWorker(i, st)
			go pageScanWorker(i, st, &wgCardSel, cardSelCh)
		}
		for i := 1; i <= st.lastPage; i++ {
			if i < cfg.PageStart {
				// Skip everything before this page. Mark as done so the routines aren't waiting for it.
				st.wgPageScan.Done()
				continue
			}

			id := i
			if cfg.Reverse {
				id = st.lastPage - i + 1
			}
			st.pageURLCh <- fmt.Sprintf("%v?page=%d", siteCfg.cardSearchURL, id)
		}
	}

	wgScanner.Wait()
	wgCardSel.Wait()
	close(cardSelCh)
	close(cardCh)
	biri.Done()

	return nil
}

func aggregate(cfg Config, r reducer) error {
	cardCh := make(chan Card, maxScrapeWorker)

	var wg sync.WaitGroup
	wg.Add(1)

	reducerCfg := reducerConfig{
		wg:     &wg,
		cardCh: cardCh,
	}

	go r.reduce(reducerCfg)

	err := CardsStream(cfg, cardCh)

	wg.Wait()

	return err
}

func Cards(cfg Config) ([]Card, error) {
	var reducer cardListReducer
	err := aggregate(cfg, &reducer)

	return reducer.cards, err
}

func Boosters(cfg Config) (map[string]Booster, error) {
	var reducer boosterReducer
	err := aggregate(cfg, &reducer)

	return reducer.boosterMap, err
}
