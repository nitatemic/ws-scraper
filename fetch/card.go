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
	"html"
	"image"
	"log/slog"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Card info to export
type Card struct {
	// CardNumber is the full card number/code used to identify each card.
	// It typically consists of the SetID, Side, Release, ReleasePackID, and ID,
	// though the format is different in some situations.
	CardNumber string `json:"cardNumber"`
	// SetID is the alphanumeric string found at the beginning of card numbers,
	// before the "/"".
	SetID string `json:"setId"`
	// SetName is the official name of the set.
	SetName       string `json:"setName"`
	ExpansionName string `json:"expansionName"`
	// Side is either "W" for Weiss, or "S" for Schwarz.
	Side string `json:"side"`
	// Release typically consists of the card's side, followed by a number
	// (the release pack ID) indicating which consecutive release for the relative
	// side the release is.
	// For example, "W64" would mean the 64th set of the Weiss side.
	// There are certain situations that don't follow the aforementioned format,
	// such as with promo cards (eg. BSF2024) or special sets (eg. EN-W03).
	Release string `json:"release"`
	// ReleasePackID indicates which consecutive release for the relative
	// side the release is.
	// For example, "W64" would mean the 64th set of the Weiss side.
	// For cards with non-standard release codes, a best-effort/most sensible
	// ID is chosen (eg. 2021 from BSL2021). This may be empty if there's
	// no sensible ID to choose (eg. from TCPR-P01).
	ReleasePackID string `json:"releasePackId"`
	// ID of the card within the set+release. This is usually the last part
	// of the card number (after the -).
	ID string `json:"id"`
	// Language the card is printed in.
	Language string `json:"language"`

	// Type can be either "CH" for character, "EV" for event, or "CX" for climax.
	Type string `json:"type"`

	// Name of the card.
	Name string `json:"name"`
	// Color of the card. Should be either "BLUE", "GREEN", "RED", or "YELLOW".
	// ...Except for the two purple cards (むらさきパプリス(PY/S38-125) and むらさきぷよ(PY/S38-120)).
	Color string `json:"color"`
	// Cost to play the card
	Cost string `json:"cost"`
	// Level required in order to play the card.
	Level string `json:"level"`
	// Power indicates the card's battle strength. Only valid for Character cards.
	Power string `json:"power"`
	// Soul is an integer indicating how many soul points the card has. Only valid for Character cards.
	Soul string `json:"soul"`
	// Text describing the card's abilities.
	Text []string `json:"text"`
	// Traits indicating the attributes the card has. These are often referenced in card text.
	Traits []string `json:"traits"`
	// Triggers that the card has and are activated during trigger checks.
	Triggers []string `json:"triggers"`

	FlavorText string      `json:"flavorText"`
	ImageURL   string      `json:"imageURL"`
	Image      image.Image `json:"-"`
	Rarity     string      `json:"rarity"`

	Version string `json:"version"`
}

// CardModelVersion : Card format version
const CardModelVersion = "1"

var (
	standardCardSuffixRE = regexp.MustCompile(`(?P<setID>[a-zA-Z0-9]+)/(?P<release>[a-zA-Z0-9-]+)-(?P<id>[a-zA-Z0-9]+\+?)$`)

	standardReleaseRE = regexp.MustCompile(`(?P<code>[a-zA-Z-]+)(?P<packID>[0-9]+)`)
)

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

func filterDash(st string) string {
	if strings.Contains(st, "-") {
		return ""
	}
	return st
}

// extractData extract data to card
func extractData(config siteConfig, mainHTML *goquery.Selection) Card {
	switch config.languageCode {
	case "EN":
		return extractDataEn(config, mainHTML)
	case "JP":
		return extractDataJp(config, mainHTML)
	default:
		slog.Error(fmt.Sprintf("Unsupported site: %q", config.languageCode))
		return Card{}
	}
}

func extractDataEn(config siteConfig, mainHTML *goquery.Selection) Card {
	txtArea := mainHTML.Find(".p-cards__detail-textarea").Last()
	cardNumber := txtArea.Find(".number").First().Last().Text()
	defer func() {
		if err := recover(); err != nil {
			slog.With("cardnumber", cardNumber).Error(fmt.Sprintf("Panic during card extraction=%v", err))
		}
	}()

	cardNumber = sanitizeCardNumber(cardNumber)
	slog.Debug(fmt.Sprintf("Start card: %s", cardNumber))

	setID, release, releasePackID, cardID := parseCardNumber(cardNumber)

	cardName := mainHTML.Find(".ttl").Last().Text()
	imageCardURL, _ := mainHTML.Find("div.image img").Attr("src")

	info := make(map[string]string)
	mainHTML.Find("dl").Each(func(i int, s *goquery.Selection) {
		dt := strings.TrimSpace(s.Find("dt").First().Text())
		dd := s.Find("dd").First()
		ddText := strings.TrimSpace(dd.Text())
		switch dt {
		case "Card Type":
			switch ddText {
			case "Event":
				info["type"] = "EV"
			case "Character":
				info["type"] = "CH"
			case "Climax":
				info["type"] = "CX"
			}
		case "Color":
			if u, ok := dd.Find("img").First().Attr("src"); ok {
				_, colorName := path.Split(u)
				info["color"] = strings.ToUpper(strings.Split(colorName, ".")[0])
			} else {
				slog.With("cardnumber", cardNumber).Error("Failed to get color")
			}
		case "Cost":
			info["cost"] = ddText
		case "Expansion":
			info["expansion"] = ddText
		case "Level":
			info["level"] = ddText
		case "Power":
			info["power"] = ddText
		case "Rarity":
			info["rarity"] = ddText
		case "Side":
			if u, ok := dd.Find("img").First().Attr("src"); ok {
				_, side := path.Split(u)
				info["side"] = strings.ToUpper(strings.Split(side, ".")[0])
			} else {
				slog.With("cardnumber", cardNumber).Error("Failed to get side")
			}
		case "Soul":
			info["soul"] = strconv.Itoa(dd.Children().Length())
		case "Traits":
			info["specialAttribute"] = ddText
		case "Trigger":
			var res bytes.Buffer
			dd.Children().Each(func(i int, ss *goquery.Selection) {
				if i != 0 {
					res.WriteString(" ")
				}
				_, trigger := path.Split(ss.AttrOr("src", "yay"))
				res.WriteString(triggersMap[strings.Split(trigger, ".")[0]])
			})
			info["trigger"] = strings.ToUpper(strings.TrimSpace(res.String()))
		default:
			slog.With("cardnumber", cardNumber).Error(fmt.Sprintf("Unknown detail: %v", dt))
		}
	})

	// Flavor text
	flvr := strings.TrimSpace(txtArea.Find(".p-cards__detail-serif").Text())
	if flvr != "" && flvr != "-" {
		info["flavourText"] = flvr
	}

	ability, err := extractAbilities(mainHTML.Find(".p-cards__detail p").Last())
	if err != nil {
		slog.With("cardnumber", cardNumber).Error(fmt.Sprintf("Failed to get ability node: %v", err))
	}

	card := Card{
		CardNumber: cardNumber,
		SetID:      setID,
		// TODO: Figure out how to get EN set name. It's no longer on the card details page
		// SetName:     setName,
		ExpansionName: info["expansion"],
		Side:          info["side"],
		Release:       release,
		ReleasePackID: releasePackID,
		ID:            cardID,
		Language:      "EN",
		Type:          info["type"],
		Name:          cardName,
		Level:         filterDash(info["level"]),
		Cost:          filterDash(info["cost"]),
		FlavorText:    info["flavourText"],
		Color:         info["color"],
		Power:         filterDash(info["power"]),
		Rarity:        info["rarity"],
		Text:          ability,
		Version:       CardModelVersion,
	}
	if fullURL, err := joinPath(config.baseURL, imageCardURL); err == nil {
		card.ImageURL = fullURL.String()
	} else {
		slog.With("cardnumber", cardNumber).Error(fmt.Sprintf("Couldn't form full image URL: %v", err))
		card.ImageURL = imageCardURL
	}
	if info["specialAttribute"] != "" {
		card.Traits = strings.Split(info["specialAttribute"], "・")
	}
	if info["trigger"] != "" {
		card.Triggers = strings.Split(info["trigger"], " ")
	}
	if card.Type == "CH" {
		card.Soul = info["soul"]
	}
	return card
}

func extractDataJp(config siteConfig, mainHTML *goquery.Selection) Card {
	titleSpan := mainHTML.Find("h4 span").Last().Text()
	defer func() {
		if err := recover(); err != nil {
			slog.With("cardnumber", titleSpan).Error(fmt.Sprintf("Panic during card extraction=%v", err))
		}
	}()

	cardNumber := sanitizeCardNumber(titleSpan)
	slog.Debug(fmt.Sprintf("Start card: %s", titleSpan))

	setID, release, releasePackID, cardID := parseCardNumber(cardNumber)

	setName := strings.TrimSpace(strings.Split(mainHTML.Find("h4").Text(), ") -")[1])
	imageCardURL, _ := mainHTML.Find("a img").Attr("src")

	ability, err := extractAbilities(mainHTML.Find("span").Last())
	if err != nil {
		slog.With("cardnumber", titleSpan).Error(fmt.Sprintf("Failed to get ability node: %v", err))
	}

	infos := make(map[string]string)
	mainHTML.Find(".unit").Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		switch {
		// Color
		case strings.HasPrefix(txt, "色："):
			_, colorName := path.Split(s.Children().AttrOr("src", "yay"))
			infos["color"] = strings.ToUpper(strings.Split(colorName, ".")[0])
			// Card type
		case strings.HasPrefix(txt, "種類："):
			cType := strings.TrimSpace(strings.TrimPrefix(txt, "種類："))

			switch cType {
			case "イベント":
				infos["type"] = "EV"
			case "キャラ":
				infos["type"] = "CH"
			case "クライマックス":
				infos["type"] = "CX"
			}
			// Cost
		case strings.HasPrefix(txt, "コスト："):
			cost := strings.TrimSpace(strings.TrimPrefix(txt, "コスト："))
			infos["cost"] = cost
			// Flavor text
		case strings.HasPrefix(txt, "フレーバー："):
			flvr := strings.TrimSpace(strings.TrimPrefix(txt, "フレーバー："))
			infos["flavourText"] = flvr
			// Level
		case strings.HasPrefix(txt, "レベル："):
			lvl := strings.TrimSpace(strings.TrimPrefix(txt, "レベル："))
			infos["level"] = lvl
			// Power
		case strings.HasPrefix(txt, "パワー："):
			pwr := strings.TrimSpace(strings.TrimPrefix(txt, "パワー："))
			infos["power"] = pwr
			// Rarity
		case strings.HasPrefix(txt, "レアリティ："):
			rarity := strings.TrimSpace(strings.TrimPrefix(txt, "レアリティ："))
			infos["rarity"] = rarity
			// Side
		case strings.HasPrefix(txt, "サイド："):
			_, side := path.Split(s.Children().AttrOr("src", "yay"))
			infos["side"] = strings.ToUpper(strings.Split(side, ".")[0])
			// Soul
		case strings.HasPrefix(txt, "ソウル："):
			infos["soul"] = strconv.Itoa(s.Children().Length())
			// Trigger
		case strings.HasPrefix(txt, "トリガー："):
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
		case strings.HasPrefix(txt, "特徴："):
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
			slog.With("cardnumber", titleSpan).Error(fmt.Sprintf("Unknown detail: %q", txt))
		}
	})

	card := Card{
		CardNumber:    titleSpan,
		SetID:         setID,
		SetName:       setName,
		Side:          infos["side"],
		Release:       release,
		ReleasePackID: releasePackID,
		ID:            cardID,
		Language:      "JP",
		Type:          infos["type"],
		Name:          strings.TrimSpace(mainHTML.Find("h4 span").First().Text()),
		Level:         filterDash(infos["level"]),
		FlavorText:    infos["flavourText"],
		Color:         infos["color"],
		Power:         filterDash(infos["power"]),
		Cost:          filterDash(infos["cost"]),
		Rarity:        infos["rarity"],
		Text:          ability,
		Version:       CardModelVersion,
	}
	if fullURL, err := joinPath(config.baseURL, imageCardURL); err == nil {
		card.ImageURL = fullURL.String()
	} else {
		slog.With("cardnumber", titleSpan).Error(fmt.Sprintf("Couldn't form full image URL: %v", err))
		card.ImageURL = imageCardURL
	}
	if infos["specialAttribute"] != "" {
		card.Traits = strings.Split(infos["specialAttribute"], "・")
	}
	if infos["trigger"] != "" {
		card.Triggers = strings.Split(infos["trigger"], " ")
	}
	if card.Type == "CH" {
		card.Soul = infos["soul"]
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
		ability = append(ability, html.UnescapeString(line))
	}
	return ability, err
}

func sanitizeCardNumber(cn string) string {
	// The website sometimes shows "%2B" instead of + for some cards (eg. SSP+ rarity).
	cn = strings.ReplaceAll(cn, "%2B", "+")

	// The website sometimes puts a + in the displayed card number (eg. RWBY/BRO2021-01+PR) even though it shouldn't be there.
	plusCnt := strings.Count(cn, "+")
	if plusCnt > 0 {
		repCnt := plusCnt
		if strings.LastIndex(cn, "+") == len(cn)-1 {
			repCnt--
		}
		cn = strings.Replace(cn, "+", " ", repCnt)
	}
	return cn
}

func parseCardNumber(cn string) (setID, release, releasePackID, id string) {
	if matches := standardCardSuffixRE.FindStringSubmatch(cn); matches != nil {
		setID = matches[1]
		release = matches[2]
		id = matches[3]
		if relMatches := standardReleaseRE.FindStringSubmatch(release); relMatches != nil {
			releasePackID = relMatches[2]
			return
		}
		releasePackID = ""
		return
	}

	if strings.Contains(cn, "/") {
		setID = strings.Split(cn, "/")[0]
		setInfo := strings.Split(strings.Split(cn, "/")[1], "-")
		if len(setInfo) > 1 {
			release = setInfo[0]
			id = setInfo[1]

			if relMatches := standardReleaseRE.FindStringSubmatch(release); relMatches != nil {
				releasePackID = relMatches[2]
				return
			}
			releasePackID = ""
		}
		return
	} else {
		slog.With("cardnumber", cn).Error(fmt.Sprintf("Can't get set info from: %s", cn))
	}
	return
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
