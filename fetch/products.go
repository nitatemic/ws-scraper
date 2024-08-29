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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/Akenaide/biri"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

const PRODUCTS_URL = "https://ws-tcg.com/products/page/"

var BAN_PRODUCT = []string{
	"new_title_ws",
	"resale_news",
	"bp_renewal",
}

var TITLE_AND_WORK_NUMBER_REGEXP = regexp.MustCompile(".*/ .*：([\\w,]+)")

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
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		defer resp.Body.Close()
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

	matches := TITLE_AND_WORK_NUMBER_REGEXP.FindStringSubmatch(titleAndWorkNumber)
	if matches == nil {
		return ProductInfo{}, fmt.Errorf("String %q doesn't match expected format", titleAndWorkNumber)
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

func fetchProduct(page string) {
	productList := []ProductInfo{}
	doc := getDocument(PRODUCTS_URL + page)

	doc.Find(".product-list .show-detail a").Each(func(i int, s *goquery.Selection) {
		productDetail := s.AttrOr("href", "nope")
		for _, ban := range BAN_PRODUCT {
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

	res, errMarshal := json.Marshal(productList)
	if errMarshal != nil {
		log.Println("error marshal", errMarshal)
	}
	var buffer bytes.Buffer
	out, err := os.Create("product.json")
	if err != nil {
		log.Println("write error", err.Error())
	}
	json.Indent(&buffer, res, "", "\t")
	buffer.WriteTo(out)
	out.Close()
	log.Println("Finished")
}

// productsCmd represents the products command
var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Get products information",
	Long: `Get products information.
It will output the ReleaseDate, Title, Image, SetCode, LicenceCode in a 'product.json' file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("products called")
		biri.Config.PingServer = "https://ws-tcg.com/"
		biri.Config.TickMinuteDuration = 1
		biri.Config.Timeout = 25
		biri.ProxyStart()

		fetchProduct(cmd.Flag("page").Value.String())
	},
}

func init() {
	rootCmd.AddCommand(productsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// productsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	productsCmd.Flags().Int16P("page", "p", 1, "Give which page to parse")
}
