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
	"log/slog"
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
	baseURLValues              func() url.Values
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
		baseURL: "https://en.ws-tcg.com/",
		baseURLValues: func() url.Values {
			return url.Values{
				"view": {"text"},
			}
		},
		cardListURL:   "https://en.ws-tcg.com/cardlist/",
		cardSearchURL: "https://en.ws-tcg.com/cardlist/searchresults/",
		languageCode:  "EN",
		lastPageFunc: func(doc *goquery.Document) int {
			numCardsS := doc.Find(".c-search__results-item span").First().Text()
			numCards, err := strconv.Atoi(numCardsS)
			if err != nil {
				slog.Error(fmt.Sprintf("Couldn't get num cards: %v", err))
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
				slog.With("url", resp.Request.URL).Error(fmt.Sprintf("goquery error: %v", err))
				return false
			}
			resultList := doc.Find(".p_cards__results-box ul li")

			if resultList.Length() == 0 && resp.StatusCode == http.StatusOK {
				slog.With("url", resp.Request.URL).Warn("No cards on response page")
			} else {
				slog.With("url", resp.Request.URL).Debug("Found cards!")
				resultList.Each(func(i int, s *goquery.Selection) {
					subPath, exists := s.Find("a").First().Attr("href")
					if !exists {
						slog.With("url", resp.Request.URL).Error(fmt.Sprintf("Error getting sub path: %v", err))
						return
					}
					fp, err := joinPath(task.siteConfig.baseURL, subPath)
					if err != nil {
						slog.With("url", resp.Request.URL).Error(fmt.Sprintf("Error getting full path: %v", err))
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
						slog.With("url", fullPath).Error(fmt.Sprintf("Failed to get detailed page %s", sc), "error", err)
					} else {
						proxy.Readd()
						doc, err := goquery.NewDocumentFromReader(detailedPageResp.Body)
						if err != nil {
							task.pageURLCh <- detailedPageResp.Request.URL.String()
							slog.With("url", detailedPageResp.Request.URL).Error(fmt.Sprintf("goquery error: %v", err))
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
						"view":      {"text"},
						"expansion": {m[1]},
					}
				}
			}
			return nil
		},
		supportTitleNumber: true,
	},
	Jp: {
		baseURL: "https://ws-tcg.com/",
		baseURLValues: func() url.Values {
			return url.Values{
				"cmd":             {"search"},
				"show_page_count": {"100"},
				"show_small":      {"0"},
			}
		},
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
				slog.With("url", resp.Request.URL).Error(fmt.Sprintf("goquery error: %v", err))
				return false
			}
			resultTable := doc.Find(".search-result-table tr")

			if resultTable.Length() == 0 && resp.StatusCode == http.StatusOK {
				slog.With("url", resp.Request.URL).Warn("No cards on response page")
			} else {
				slog.With("url", resp.Request.URL).Debug("Found cards!")
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
	// ReleaseCode is the first set of characters following the / in the card
	// number. See Card.Release for more information.
	ReleaseCode string
	Cards       []Card
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

func (s *scrapeTask) getLastPage() (int, error) {
	slog.Info(fmt.Sprintf("Getting last page of %q with %v", s.siteConfig.cardSearchURL, s.urlValues))
	resp, err := http.PostForm(fmt.Sprintf("%v?page=%d", s.siteConfig.cardSearchURL, 1), s.urlValues)
	if err != nil {
		return 0, fmt.Errorf("error getting last page: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error parsing last page: %v", err)
	}

	last := s.siteConfig.lastPageFunc(doc)

	slog.With("url", resp.Request.URL).Info(fmt.Sprintf("Last page is %d for %v", last, s.urlValues))
	s.lastPage = last
	return last, nil
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
		slog.Debug(fmt.Sprintf("ID %d: fetching page %q with params %v", id, link, task.urlValues))
		proxy := biri.GetClient()
		slog.Debug("Got proxy")
		proxy.Client.Jar = task.cookieJar

		resp, err := proxy.Client.PostForm(link, task.urlValues)
		if err != nil || resp.StatusCode != http.StatusOK {
			slog.With("url", link).Info("Ban proxy", "error", err)
			proxy.Ban()
			// Retry again later
			task.pageURLCh <- link
		} else {
			proxy.Readd()
			task.pageRespCh <- resp
		}
	}
	slog.Info(fmt.Sprintf("Page fetch worker %d done", id))
}

func pageScanWorker(
	id int,
	task *scrapeTask,
	wgCardSel *sync.WaitGroup,
	cardSelCh chan<- *goquery.Selection,
) {
	for resp := range task.pageRespCh {
		slog.Debug(fmt.Sprintf("Start scanning page: %v", resp.Request.URL))
		if task.siteConfig.pageScanParseFunc(task, wgCardSel, cardSelCh, resp) {
			task.wgPageScan.Done()
		}
		slog.Debug(fmt.Sprintf("Finish scanning page: %v", resp.Request.URL))
	}
	slog.Info(fmt.Sprintf("Page scan worker %d done", id))
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
		boosterObj.ReleaseCode = boosterCode

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
		return fmt.Errorf("unsupported language: %q", cfg.Language)
	} else {
		siteCfg = c
		slog.Info(fmt.Sprintf("Fetching %s cards", cfg.Language))
	}

	slog.Info("Streaming cards", "config", cfg)

	prepareBiri(siteCfg)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return fmt.Errorf("failed to get new cookiejar: %v", err)
	}

	biri.ProxyStart()

	urlValues := siteCfg.baseURLValues()
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
			return fmt.Errorf("can't use title number on %s site", cfg.Language)
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
			return fmt.Errorf("error getting recent: %v", err)
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing recent: %v", err)
		}
		for _, recent := range getTasksForRecentReleases(siteCfg, doc) {
			copyTask := defaultScrapeTask
			copyTask.urlValues = recent.urlValues
			slog.Debug(fmt.Sprintf("default scrape task=%v, recent=%v", defaultScrapeTask, recent))
			scrapeTasks = append(scrapeTasks, &copyTask)
		}
	} else {
		scrapeTasks = append(scrapeTasks, &defaultScrapeTask)
	}

	loopNum := 0
	for _, st := range scrapeTasks {
		lastPage, err := st.getLastPage()
		if err != nil {
			return err
		}
		loopNum += lastPage
		st.pageURLCh = make(chan string, lastPage)
		st.pageRespCh = make(chan *http.Response, maxScrapeWorker)
		st.wgPageScan = &sync.WaitGroup{}
		st.wgPageScan.Add(lastPage)
	}

	slog.Debug(fmt.Sprintf("Number of loop %v", loopNum))

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

// ExpansionList returns a map of expansion numbers to their titles for the
// specified language in the Config.
func ExpansionList(cfg Config) (map[int]string, error) {
	var siteCfg siteConfig
	if c, ok := siteConfigs[cfg.Language]; !ok {
		return nil, fmt.Errorf("unsupported language: %q", cfg.Language)
	} else {
		siteCfg = c
		slog.Info(fmt.Sprintf("Fetching %s expansion list", cfg.Language))
	}

	prepareBiri(siteCfg)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		err = fmt.Errorf("failed to get new cookiejar: %v", err)
		slog.Error(err.Error())
		return nil, err
	}

	biri.ProxyStart()

	proxy := biri.GetClient()
	slog.Debug("Got proxy")
	proxy.Client.Jar = jar

	resp, err := proxy.Client.PostForm(siteCfg.cardListURL, url.Values{})
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("couldn't read page: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery error for page %q: %v", resp.Request.URL, err)
	}

	expansionList := doc.Find("select#expansion option")
	if expansionList.Length() == 0 && resp.StatusCode == http.StatusOK {
		return nil, fmt.Errorf("couldn't find expansion list")
	}

	eMap := make(map[int]string)
	expansionList.Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("value")
		val = strings.TrimSpace(val)
		if !exists || val == "" {
			// This is probably the "All" option
			slog.Warn(fmt.Sprintf("Option %q had no value", s.Text()))
			return
		}
		if v, err := strconv.Atoi(val); err != nil {
			slog.Error(fmt.Sprintf("Error parsing expansion value: %v", err))
		} else {
			eMap[v] = s.Text()
		}

	})

	return eMap, nil
}
