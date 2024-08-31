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

package fetch

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"strings"

	"github.com/Akenaide/biri"
	"github.com/PuerkitoBio/goquery"
)

const ProductsUrl = "https://ws-tcg.com/products/page/"

var banProduct = []string{
	"new_title_ws",
	"resale_news",
	"bp_renewal",
}

var titleAndWorkNumberRegexp = regexp.MustCompile(`.*/ .*：([\w,]+)`)

// ProductInfo represents the extracted information from the HTML
type ProductInfo struct {
	ReleaseDate string
	Title       string
	LicenceCode string
	Image       string
	SetCode     string
}

func getDocument(url string) *goquery.Document {
	var doc *goquery.Document

	for {
		var err error
		proxy := biri.GetClient()
		resp, err := proxy.Client.Get(url)
		if err != nil || resp.StatusCode != 200 {
			log.Println("Error on fetch page: ", err)
			proxy.Ban()
			continue
		}
		defer resp.Body.Close()
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println("Error on parse page: ", err)
			proxy.Ban()
			continue
		}
		proxy.Readd()
		break
	}

	return doc
}

func extractProductInfo(doc *goquery.Document) (ProductInfo, error) {
	var setCode string
	releaseDate := strings.Split(strings.TrimSpace(doc.Find(".release strong").Text()), "(")[0]
	titleAndWorkNumber := strings.TrimSpace(doc.Find(".release").Text())

	matches := titleAndWorkNumberRegexp.FindStringSubmatch(titleAndWorkNumber)
	if matches == nil {
		return ProductInfo{}, fmt.Errorf("string %q doesn't match expected format", titleAndWorkNumber)
	}
	licenceCode := matches[1]
	doc.Find(".entry-content img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		// Extract the filename from the path
		filename := path.Base(src)

		// Extract "W109" from the filename
		parts := strings.Split(filename, "_")
		if len(parts) >= 4 {
			setCode = parts[2]
		}
	})

	return ProductInfo{
		ReleaseDate: releaseDate,
		Title:       doc.Find(".entry-content > h3").Text(),
		LicenceCode: licenceCode,
		SetCode:     setCode,
		Image:       doc.Find(".product-detail .alignright img").AttrOr("src", "notfound"),
	}, nil
}

func Products(page string) []ProductInfo {
	biri.Config.PingServer = "https://ws-tcg.com/"
	biri.Config.TickMinuteDuration = 1
	biri.Config.Timeout = 25
	biri.ProxyStart()

	var productList []ProductInfo
	doc := getDocument(ProductsUrl + page)

	doc.Find(".product-list .show-detail a").Each(func(i int, s *goquery.Selection) {
		productDetail := s.AttrOr("href", "nope")
		for _, ban := range banProduct {
			if strings.Contains(productDetail, ban) {
				return
			}
		}
		log.Println("Extract :", productDetail)
		doc := getDocument(productDetail)
		if productInfo, err := extractProductInfo(doc); err != nil {
			log.Println("Error getting product info:", err)
		} else {
			productList = append(productList, productInfo)
		}
	})

	return productList
}
