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
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Card info to export
type Card struct {
	Set           string   `json:"set"`
	SetName       string   `json:"setName"`
	Side          string   `json:"side"`
	Release       string   `json:"release"`
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Language      string   `json:"language"`
	CardType      string   `json:"cardType"`
	Colour        string   `json:"colour"`
	Level         string   `json:"level"`
	Cost          string   `json:"cost"`
	Power         string   `json:"power"`
	Soul          string   `json:"soul"`
	Rarity        string   `json:"rarity"`
	FlavourText   string   `json:"flavourText"`
	Trigger       []string `json:"trigger"`
	Ability       []string `json:"ability"`
	SpecialAttrib []string `json:"specialAttrib"`
	Version       string   `json:"version"`
	Cardcode      string   `json:"cardcode"`
	ImageURL      string   `json:"imageURL"`
	Tags          []string `json:"tags"`
}

// CardModelVersion : Card format version
const CardModelVersion = "4"

var imgRE = regexp.MustCompile(`<img .*>`)

var suffix = []string{
	"SP",
	"S",
	"R",
}

var baseRarity = []string{
	"C",
	"CC",
	"CR",
	"FR",
	"MR",
	"PR",
	"PS",
	"R",
	"RE",
	"RR",
	"RR+",
	"TD",
	"U",
	"AR",
}

var triggersMap = map[string]string{
	"soul":     "SOUL",
	"salvage":  "COMEBACK",
	"draw":     "DRAW",
	"stock":    "POOL",
	"treasure": "TREASURE",
	"shot":     "SHOT",
	"bounce":   "RETURN",
	"gate":     "GATE",
	"standby":  "STANDBY",
	"choice":   "CHOICE",
}

func processInt(st string) string {
	if strings.Contains(st, "-") {
		st = "0"
	}
	return st
}

// extractData extract data to card
func extractData(config siteConfig, mainHTML *goquery.Selection) Card {
	titleSpan := mainHTML.Find("h4 span").Last().Text()
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic for %v. Error=%v", titleSpan, err)
		}
	}()
	log.Println("Start card:", titleSpan)
	var set string
	var setInfo []string
	if strings.Contains(titleSpan, "/") {
		set = strings.Split(titleSpan, "/")[0]
		setInfo = strings.Split(strings.Split(titleSpan, "/")[1], "-")
	} else {
		// TODO: deal with "BSF2024-03 PR" and similar cards
		log.Println("Can't get set info from:", titleSpan)
	}
	setName := strings.TrimSpace(strings.Split(mainHTML.Find("h4").Text(), ") -")[1])
	imageCardURL, _ := mainHTML.Find("a img").Attr("src")

	ability, err := extractAbilities(mainHTML.Find("span").Last())
	if err != nil {
		log.Printf("Failed to get ability node: %v\n", err)
	}

	infos := make(map[string]string)
	mainHTML.Find(".unit").Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		switch {
		// Color
		case strings.HasPrefix(txt, "色：") || strings.HasPrefix(txt, "[Color]:"):
			_, colorName := path.Split(s.Children().AttrOr("src", "yay"))
			infos["color"] = strings.ToUpper(strings.Split(colorName, ".")[0])
			// Card type
		case strings.HasPrefix(txt, "種類：") || strings.HasPrefix(txt, "[Card Type]:"):
			cType := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "種類："), "[Card Type]:"))

			switch cType {
			case "イベント", "Event":
				infos["type"] = "EV"
			case "キャラ", "Character":
				infos["type"] = "CH"
			case "クライマックス", "Climax":
				infos["type"] = "CX"
			}
			// Cost
		case strings.HasPrefix(txt, "コスト：") || strings.HasPrefix(txt, "[Cost]:"):
			cost := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "コスト："), "[Cost]:"))
			infos["cost"] = cost
			// Flavor text
		case strings.HasPrefix(txt, "フレーバー：") || strings.HasPrefix(txt, "[Flavor Text]:"):
			flvr := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "フレーバー："), "[Flavor Text]:"))
			infos["flavourText"] = flvr
			// Level
		case strings.HasPrefix(txt, "レベル：") || strings.HasPrefix(txt, "[Level]:"):
			lvl := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "レベル："), "[Level]:"))
			infos["level"] = lvl
			// Power
		case strings.HasPrefix(txt, "パワー：") || strings.HasPrefix(txt, "[Power]:"):
			pwr := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "パワー："), "[Power]:"))
			infos["power"] = pwr
			// Rarity
		case strings.HasPrefix(txt, "レアリティ：") || strings.HasPrefix(txt, "[Rarity]:"):
			rarity := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(txt, "レアリティ："), "[Rarity]:"))
			infos["rarity"] = rarity
			// Side
		case strings.HasPrefix(txt, "サイド：") || strings.HasPrefix(txt, "[Side]:"):
			_, side := path.Split(s.Children().AttrOr("src", "yay"))
			infos["side"] = strings.ToUpper(strings.Split(side, ".")[0])
			// Soul
		case strings.HasPrefix(txt, "ソウル：") || strings.HasPrefix(txt, "[Soul]:"):
			infos["soul"] = strconv.Itoa(s.Children().Length())
			// Trigger
		case strings.HasPrefix(txt, "トリガー：") || strings.HasPrefix(txt, "[Trigger]:"):
			var res bytes.Buffer
			s.Children().Each(func(i int, ss *goquery.Selection) {
				if i != 0 {
					res.WriteString(" ")
				}
				_, trigger := path.Split(ss.AttrOr("src", "yay"))
				res.WriteString(triggersMap[strings.Split(trigger, ".")[0]])
			})
			infos["trigger"] = strings.ToUpper(strings.TrimSpace(res.String()))
			// Trait
		case strings.HasPrefix(txt, "特徴：") || strings.HasPrefix(txt, "[Special Attribute]:"):
			var res bytes.Buffer
			s.Children().Each(func(i int, ss *goquery.Selection) {
				res.WriteString(strings.TrimSpace(ss.Text()))
			})
			if strings.Contains(res.String(), "-") {
				infos["specialAttribute"] = ""
			} else {
				infos["specialAttribute"] = strings.TrimSpace(res.String())
			}
		default:
			log.Println("Unknown:", txt)
		}
	})

	card := Card{
		Name:        strings.TrimSpace(mainHTML.Find("h4 span").First().Text()),
		Set:         set,
		SetName:     setName,
		Side:        infos["side"],
		CardType:    infos["type"],
		Level:       processInt(infos["level"]),
		FlavourText: infos["flavourText"],
		Colour:      infos["color"],
		Power:       processInt(infos["power"]),
		Soul:        infos["soul"],
		Cost:        processInt(infos["cost"]),
		Rarity:      infos["rarity"],
		Ability:     ability,
		Version:     CardModelVersion,
		Cardcode:    titleSpan,
	}
	if fullURL, err := url.JoinPath(config.baseURL, imageCardURL); err == nil {
		card.ImageURL = fullURL
	} else {
		log.Printf("Couldn't form full image URL: %v\n", err)
		card.ImageURL = imageCardURL
	}
	if infos["specialAttribute"] != "" {
		card.SpecialAttrib = strings.Split(infos["specialAttribute"], "・")
	}
	if infos["trigger"] != "" {
		card.Trigger = strings.Split(infos["trigger"], " ")
	}
	if len(setInfo) > 1 {
		card.Release = setInfo[0]
		card.ID = setInfo[1]
	}
	return card
}

func extractAbilities(abilityNode *goquery.Selection) ([]string, error) {
	var ability []string
	abilityNode.Find("img").Each(func(i int, s *goquery.Selection) {
		url, has := s.Attr("src")
		if has {
			_, _imgPlaceHolder := path.Split(url)
			_imgPlaceHolder = strings.Split(_imgPlaceHolder, ".")[0]
			t := fmt.Sprintf("[%v]", triggersMap[_imgPlaceHolder])
			s.ReplaceWithHtml(t)
		}
	})
	abilityNodeHtml, err := abilityNode.Html()
	if err != nil {
		err = fmt.Errorf("failed to get ability node: %v", err)
	}
	for _, line := range strings.Split(abilityNodeHtml, "<br/>") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ability = append(ability, line)
	}
	return ability, err
}

// IsbaseRarity check if a card is a C / U / R / RR
func IsbaseRarity(card Card) bool {
	for _, rarity := range baseRarity {
		if rarity == card.Rarity && isTrullyNotFoil(card) {
			return true
		}
	}
	return false
}

func isTrullyNotFoil(card Card) bool {
	for _, _suffix := range suffix {
		if strings.HasSuffix(card.ID, _suffix) {
			return false
		}
	}
	return true
}
