package fetch

import (
	"log"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func equalSlice(sliceA []string, sliceB []string) bool {
	if len(sliceA) != len(sliceB) {
		log.Printf("wrong len sliceA %v, len sliceB %v", len(sliceA), len(sliceB))
		return false
	}

	for i := range sliceA {
		if sliceA[i] != sliceB[i] {
			log.Printf("wrong value sliceA %v, len sliceB %v", sliceA[i], sliceB[i])
			return false
		}
	}
	return true
}

func assertCardEquals(t *testing.T, got, want Card) {
	if got.SetID != want.SetID {
		t.Errorf("Incorrect Set: got %q, want %q", got.SetID, want.SetID)
	}
	if got.SetName != want.SetName {
		t.Errorf("Incorrect SetName: got %q, want %q", got.SetName, want.SetName)
	}
	if got.Side != want.Side {
		t.Errorf("Incorrect Side: got %q, want %q", got.Side, want.Side)
	}
	if got.Release != want.Release {
		t.Errorf("Incorrect Release: got %q, want %q", got.Release, want.Release)
	}
	if got.ID != want.ID {
		t.Errorf("Incorrect ID: got %q, want %q", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("Incorrect Name: got %q, want %q", got.Name, want.Name)
	}
	if got.Language != want.Language {
		t.Errorf("Incorrect Language: got %q, want %q", got.Language, want.Language)
	}
	if got.Type != want.Type {
		t.Errorf("Incorrect CardType: got %q, want %q", got.Type, want.Type)
	}
	if got.Color != want.Color {
		t.Errorf("Incorrect Colour: got %q, want %q", got.Color, want.Color)
	}
	if got.Level != want.Level {
		t.Errorf("Incorrect Level: got %q, want %q", got.Level, want.Level)
	}
	if got.Cost != want.Cost {
		t.Errorf("Incorrect Cost: got %q, want %q", got.Cost, want.Cost)
	}
	if got.Power != want.Power {
		t.Errorf("Incorrect Power: got %q, want %q", got.Power, want.Power)
	}
	if got.Soul != want.Soul {
		t.Errorf("Incorrect Soul: got %q, want %q", got.Soul, want.Soul)
	}
	if got.Rarity != want.Rarity {
		t.Errorf("Incorrect Rarity: got %q, want %q", got.Rarity, want.Rarity)
	}
	if got.FlavorText != want.FlavorText {
		t.Errorf("Incorrect FlavourText: got %q, want %q", got.FlavorText, want.FlavorText)
	}
	if !equalSlice(got.Triggers, want.Triggers) {
		t.Errorf("Incorrect Trigger: got %v, want %v", got.Triggers, want.Triggers)
	}
	if !equalSlice(got.Abilities, want.Abilities) {
		t.Errorf("Incorrect Ability: got\n %v,\nwant\n %v", got.Abilities, want.Abilities)
	}
	if !equalSlice(got.Traits, want.Traits) {
		t.Errorf("Incorrect SpecialAttrib: got %v, want %v", got.Traits, want.Traits)
	}
	if got.Version != want.Version {
		t.Errorf("Incorrect Version: got %q, want %q", got.Version, want.Version)
	}
	if got.ImageURL != want.ImageURL {
		t.Errorf("Incorrect ImageURL: got %q, want %q", got.ImageURL, want.ImageURL)
	}
	if got.CardNumber != want.CardNumber {
		t.Errorf("Incorrect Cardcode: got %q, want %q", got.CardNumber, want.CardNumber)
	}
}

func TestExtractData_jp(t *testing.T) {
	chara := `
	<th><a href="/cardlist/?cardno=BD/W63-036SPMa&amp;l"><img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/b/bd_w63/bd_w63_036spma.gif" alt="“私達、参上っ！”上原ひまり"/></a></th>
	<td>
	<h4><a href="/cardlist/?cardno=BD/W63-036SPMa&amp;l"><span>
	“私達、参上っ！”上原ひまり</span>(<span>BD/W63-036SPMa</span>)</a> -「バンドリ！ ガールズバンドパーティ！」Vol.2<br/></h4>
	<span class="unit">
	サイド：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/w.gif"/></span>
	<span class="unit">種類：キャラ</span>
	<span class="unit">レベル：2</span><br/>
	<span class="unit">色：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/green.gif"/></span>
	<span class="unit">パワー：6000</span>
	<span class="unit">ソウル：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/soul.gif"/><img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/soul.gif"/></span>
	<span class="unit">コスト：1</span><br/>
	<span class="unit">レアリティ：SPMa</span>
	<span class="unit">トリガー：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/soul.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/bounce.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/shot.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/treasure.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/standby.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/salvage.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/gate.gif"/>
	<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/draw.gif"/>
	</span>
	<span class="unit">特徴：<span>音楽・Afterglow</span></span><br/>
	<span class="unit">フレーバー：-</span><br/>
	<br/>
	<span>【永】 あなたのターン中、他のあなたの「“止まらずに、前へ”美竹蘭」がいるなら、このカードのパワーを＋6000。<br/>【自】［(1)］ このカードがアタックした時 、あなたはコストを払ってよい。そうしたら、そのアタック中、あなたはトリガーステップにトリガーチェックを2回行う。</span>
	</td>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(chara))
	expectedTrigger := []string{"SOUL", "RETURN", "SHOT", "TREASURE", "STANDBY", "COMEBACK", "GATE", "DRAW"}
	expectedTrait := []string{"音楽", "Afterglow"}
	expectedAbility := []string{
		"【永】 あなたのターン中、他のあなたの「“止まらずに、前へ”美竹蘭」がいるなら、このカードのパワーを＋6000。",
		"【自】［(1)］ このカードがアタックした時 、あなたはコストを払ってよい。そうしたら、そのアタック中、あなたはトリガーステップにトリガーチェックを2回行う。",
	}

	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[JP], doc.Clone())
	if card.Name != "“私達、参上っ！”上原ひまり" {
		t.Errorf("got %v: expected “私達、参上っ！”上原ひまり", card.Name)
	}
	if card.SetID != "BD" {
		t.Errorf("got %v: expected BD", card.SetID)
	}
	if card.Side != "W" {
		t.Errorf("got %v: expected W", card.Side)
	}
	if card.Release != "W63" {
		t.Errorf("got %v: expected W63", card.Release)
	}
	if card.ID != "036SPMa" {
		t.Errorf("got %v: expected 036SPMa", card.ID)
	}
	if card.Level != "2" {
		t.Errorf("got %v: expected 2", card.Level)
	}
	if card.Color != "GREEN" {
		t.Errorf("got %v: expected GREEN", card.Color)
	}
	if card.Power != "6000" {
		t.Errorf("got %v: expected 6000", card.Power)
	}
	if card.Soul != "2" {
		t.Errorf("got %v: expected 2", card.Soul)
	}
	if card.Cost != "1" {
		t.Errorf("got %v: expected 1", card.Cost)
	}
	if card.Type != "CH" {
		t.Errorf("got %v: expected CH", card.Type)
	}
	if card.Rarity != "SPMa" {
		t.Errorf("got %v: expected SPMa", card.Rarity)
	}
	if !equalSlice(card.Triggers, expectedTrigger) {
		t.Errorf("got %v: expected %v", card.Triggers, expectedTrigger)
	}
	if !equalSlice(card.Traits, expectedTrait) {
		t.Errorf("got %v: expected %v", card.Traits, expectedTrait)
	}
	if !equalSlice(card.Abilities, expectedAbility) {
		t.Errorf("got \n %v: expected \n %v", card.Abilities, expectedAbility)
	}
}

func TestExtractDataEvent_jp(t *testing.T) {
	chara := `
	<th><a href="/cardlist/?cardno=BD/W63-022&amp;l"><img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/b/bd_w63/bd_w63_022.gif" alt="ミッシェルからの伝言"></a></th>
	<td>
	<h4><a href="/cardlist/?cardno=BD/W63-022&amp;l"><span class="highlight_target">
	ミッシェルからの伝言</span>(<span class="highlight_target">BD/W63-022</span>)</a> -「バンドリ！ ガールズバンドパーティ！」Vol.2<br></h4>
	<span class="unit">
	サイド：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/w.gif"></span>
	<span class="unit">種類：イベント</span>
	<span class="unit">レベル：1</span><br>
	<span class="unit">色：<img src="https://s3-ap-northeast-1.amazonaws.com/static.ws-tcg.com/wordpress/wp-content/cardimages/_partimages/yellow.gif"></span>
	<span class="unit">パワー：-</span>
	<span class="unit">ソウル：-</span>
	<span class="unit">コスト：0</span><br>
	<span class="unit">レアリティ：U</span>
	<span class="unit">トリガー：－</span>
	<span class="unit">特徴：<span class="highlight_target">-・-</span></span><br>
	<span class="unit">フレーバー：美咲「あはは……ありがとう、はぐみ」</span><br>
	<br>
	<span class="highlight_target">このカードは、あなたの《ハロー、ハッピーワールド！》のキャラが2枚以下なら、手札からプレイできない。<br>あなたは自分の山札の上から2枚を、控え室に置き、自分の控え室のレベルＸ以下のキャラを1枚選び、手札に戻す。Ｘはそれらのカードのレベルの合計に等しい。（クライマックスのレベルは0として扱う）</span>
	</td>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(chara))
	var expectedTrigger []string

	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[JP], doc.Clone())
	if card.Name != "ミッシェルからの伝言" {
		t.Errorf("got %v: expected ミッシェルからの伝言", card.Name)
	}

	if !equalSlice(card.Triggers, expectedTrigger) {
		t.Errorf("got %v: expected %v", card.Triggers, expectedTrigger)
	}

	if card.Type != "EV" {
		t.Errorf("got %v: expected EV", card.Type)
	}

	if !equalSlice(card.Traits, []string{}) {
		t.Errorf("got %v: expected empty", card.Traits)
	}

	if card.Soul != "0" {
		t.Errorf("got %v: expected ''", card.Soul)
	}

	if card.Power != "0" {
		t.Errorf("got %v: expected 0", card.Power)
	}
}

func TestExtractDataCX_jp(t *testing.T) {
	chara := `
<tr>
	<th><a href="/cardlist/?cardno=BD/W63-025&amp;l"><img src="/wordpress/wp-content/images/cardlist/b/bd_w63/bd_w63_025.png" alt="キラキラのお日様"></a></th>
	<td>
	<h4><a href="/cardlist/?cardno=BD/W63-025&amp;l"><span class="highlight_target">
	キラキラのお日様</span>(<span class="highlight_target">BD/W63-025</span>)</a> -「バンドリ！ ガールズバンドパーティ！」Vol.2<br></h4>
	<span class="unit">
	サイド：<img src="/wordpress/wp-content/images/cardlist/_partimages/w.gif"></span>
	<span class="unit">種類：クライマックス</span>
	<span class="unit">レベル：-</span><br>
	<span class="unit">色：<img src="/wordpress/wp-content/images/cardlist/_partimages/yellow.gif"></span>
	<span class="unit">パワー：-</span>
	<span class="unit">ソウル：-</span>
	<span class="unit">コスト：-</span><br>
	<span class="unit">レアリティ：CR</span>
	<span class="unit">トリガー：<img src="/wordpress/wp-content/images/cardlist/_partimages/soul.gif"><img src="/wordpress/wp-content/images/cardlist/_partimages/bounce.gif"></span>
	<span class="unit">特徴：<span class="highlight_target">-</span></span><br>
	<span class="unit">フレーバー：楽しい気持ちは誰かといると生まれるものってこと！</span><br>
	<br>
	<span class="highlight_target">【永】 あなたのキャラすべてに、パワーを＋1000し、ソウルを＋1。<br>（<img src="/wordpress/wp-content/images/cardlist/_partimages/bounce.gif">：このカードがトリガーした時、あなたは相手のキャラを1枚選び、手札に戻してよい）</span>
	</td>
</tr>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(chara))
	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[JP], doc.Clone())

	expectedCard := Card{
		Name:       "キラキラのお日様",
		SetID:      "BD",
		SetName:    "「バンドリ！ ガールズバンドパーティ！」Vol.2",
		Side:       "W",
		CardNumber: "BD/W63-025",
		Release:    "W63",
		ID:         "025",
		Color:      "YELLOW",
		Language:   "JP",
		Type:       "CX",
		Soul:       "0",
		Level:      "0",
		Cost:       "0",
		FlavorText: "楽しい気持ちは誰かといると生まれるものってこと！",
		Power:      "0",
		Rarity:     "CR",
		ImageURL:   "https://ws-tcg.com/wordpress/wp-content/images/cardlist/b/bd_w63/bd_w63_025.png",
		Version:    CardModelVersion,
		Triggers:   []string{"SOUL", "RETURN"},
		Abilities: []string{
			"【永】 あなたのキャラすべてに、パワーを＋1000し、ソウルを＋1。",
			"（[RETURN]：このカードがトリガーした時、あなたは相手のキャラを1枚選び、手札に戻してよい）",
		},
	}
	assertCardEquals(t, card, expectedCard)
}

func TestExtractData_en(t *testing.T) {
	chara := `
<div class="p-cards__detail-wrapper">
	<div class="p-cards__detail-wrapper-inner">
		<div class="image"><img src="/wp/wp-content/images/cardimages/f/fs_s64/FS_BCS_2019_03.png" alt="EGOISTIC, Sakura" decoding="async">
		</div>
		<div class="p-cards__detail-textarea">
		<p class="number">FS/BCS2019-03</p>
		<p class="ttl u-mt-14 u-mt-16-sp">EGOISTIC, Sakura</p>
		<div class="p-cards__detail-type u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Expansion</dt>
			<dd>PR Card 【Schwarz Side】</dd>
			</dl>
			<dl>
			<dt>Traits</dt>
			<dd>Master・Love</dd>
			</dl>
			<dl>
			<dt>Card Type</dt>
			<dd>Character</dd>
			</dl>
			<dl>
			<dt>Rarity</dt>
			<dd>PR</dd>
			</dl>
			<dl>
			<dt>Side</dt>
			<dd>
								<img src="/cardlist/partimages/s.gif" alt="" decoding="async">
								</dd>
			</dl>
			<dl>
			<dt>Color</dt>
			<dd><img src="/wp/wp-content/images/partimages/green.gif"></dd>
			</dl>
		</div>
		<div class="p-cards__detail-status u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Level</dt>
			<dd>0</dd>
			</dl>
			<dl>
			<dt>Cost</dt>
			<dd>0</dd>
			</dl>
			<dl>
			<dt>Power</dt>
			<dd>2000</dd>
			</dl>
			<dl>
			<dt>Trigger</dt>
			<dd>-</dd>
			</dl>
			<dl>
			<dt>Soul</dt>
			<dd><img src="/wp/wp-content/images/partimages/soul.gif"></dd>
			</dl>
		</div>
		<div class="p-cards__detail u-mt-22 u-mt-40-sp">
			<p>【AUTO】 When this card is placed on the stage from your hand, choose 1 of your 《Master》 or 《Servant》 characters, and that character gets +1500 power until end of turn.</p>
		</div>
		<div class="p-cards__detail-serif u-mt-22 u-mt-40-sp">
			<p>I wish someone like this didn't exist.</p>
		</div>
		<p class="p-cards__detail-copyrights u-mt-22 u-mt-40-sp">©TYPE-MOON, ufotable, FSNPC</p>
		</div>
	</div>
</div>
`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(chara))
	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[EN], doc.Clone())
	expectedCard := Card{
		Name:          "EGOISTIC, Sakura",
		ExpansionName: "PR Card 【Schwarz Side】",
		CardNumber:    "FS/BCS2019-03",
		SetID:         "FS",
		Side:          "S",
		Release:       "BCS2019",
		ReleasePackID: "2019",
		ID:            "03",
		Level:         "0",
		Color:         "GREEN",
		Power:         "2000",
		Soul:          "1",
		Cost:          "0",
		Language:      "EN",
		Type:          "CH",
		Rarity:        "PR",
		FlavorText:    "I wish someone like this didn't exist.",
		Traits:        []string{"Master", "Love"},
		Abilities:     []string{"【AUTO】 When this card is placed on the stage from your hand, choose 1 of your 《Master》 or 《Servant》 characters, and that character gets +1500 power until end of turn."},
		ImageURL:      "https://en.ws-tcg.com/wp/wp-content/images/cardimages/f/fs_s64/FS_BCS_2019_03.png",
		Version:       CardModelVersion,
	}
	assertCardEquals(t, card, expectedCard)
}

func TestExtractData_en_multiIconAbility(t *testing.T) {
	character := `
<div class="p-cards__detail-wrapper">
	<div class="p-cards__detail-wrapper-inner">
		<div class="image"><img src="/wp/wp-content/images/cardimages/ATLA/BP/ATLA_WX04_007S.png" alt="Aang: Learning Avatar State" decoding="async">
		</div>
		<div class="p-cards__detail-textarea">
		<p class="number">ATLA/WX04-007S</p>
		<p class="ttl u-mt-14 u-mt-16-sp">Aang: Learning Avatar State</p>
		<div class="p-cards__detail-type u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Expansion</dt>
			<dd>Avatar: The Last Airbender</dd>
			</dl>
			<dl>
			<dt>Traits</dt>
			<dd>World of Avatar・Air Nomads</dd>
			</dl>
			<dl>
			<dt>Card Type</dt>
			<dd>Character</dd>
			</dl>
			<dl>
			<dt>Rarity</dt>
			<dd>SR</dd>
			</dl>
			<dl>
			<dt>Side</dt>
			<dd>
								<img src="/cardlist/partimages/w.gif" alt="" decoding="async">
								</dd>
			</dl>
			<dl>
			<dt>Color</dt>
			<dd><img src="/wp/wp-content/images/partimages/yellow.gif"></dd>
			</dl>
		</div>
		<div class="p-cards__detail-status u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Level</dt>
			<dd>2</dd>
			</dl>
			<dl>
			<dt>Cost</dt>
			<dd>1</dd>
			</dl>
			<dl>
			<dt>Power</dt>
			<dd>1000</dd>
			</dl>
			<dl>
			<dt>Trigger</dt>
			<dd><img src="/wp/wp-content/images/partimages/soul.gif"></dd>
			</dl>
			<dl>
			<dt>Soul</dt>
			<dd>-</dd>
			</dl>
		</div>
		<div class="p-cards__detail u-mt-22 u-mt-40-sp">
			<p>【CONT】 If your climax area has a climax with <img src="/wp/wp-content/images/partimages/choice.gif"> in its trigger icon, this card in all of your zones get <img src="/wp/wp-content/images/partimages/choice.gif"> in the trigger icon. If there is a climax with <img src="/wp/wp-content/images/partimages/treasure.gif"> in its trigger icon, this card in all of your zones get <img src="/wp/wp-content/images/partimages/treasure.gif"> in the trigger icon. If there is a climax with <img src="/wp/wp-content/images/partimages/standby.gif"> in its trigger icon, this card in all of your zones get <img src="/wp/wp-content/images/partimages/standby.gif"> in the trigger icon. If there is a climax with <img src="/wp/wp-content/images/partimages/gate.gif"> in its trigger icon, this card in all of your zones get <img src="/wp/wp-content/images/partimages/gate.gif"> in the trigger icon.<br>【AUTO】 【CLOCK】 Alarm If this card is the top card of your clock, and you have 4 or more 《World of Avatar》 characters, at the beginning of your climax phase, you may put the top card of your deck into your stock.</p>
		</div>
		<div class="p-cards__detail-serif u-mt-22 u-mt-40-sp">
			<p>-</p>
		</div>
		<p class="p-cards__detail-copyrights u-mt-22 u-mt-40-sp">©2023 Viacom International Inc. All Rights Reserved.</p>
		</div>
	</div>
</div>
`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(character))
	if err != nil {
		t.Fatal(err)
	}

	expectedCard := Card{
		CardNumber:    "ATLA/WX04-007S",
		SetID:         "ATLA",
		ExpansionName: "Avatar: The Last Airbender",
		Side:          "W",
		Release:       "WX04",
		ReleasePackID: "WX",
		ID:            "007S",
		Language:      "EN",
		Type:          "CH",
		Name:          "Aang: Learning Avatar State",
		Color:         "YELLOW",
		Soul:          "0",
		Level:         "2",
		Cost:          "1",
		FlavorText:    "",
		Power:         "1000",
		Rarity:        "SR",
		ImageURL:      "https://en.ws-tcg.com/wp/wp-content/images/cardimages/ATLA/BP/ATLA_WX04_007S.png",
		Triggers:      []string{"SOUL"},
		Traits:        []string{"World of Avatar", "Air Nomads"},
		Abilities: []string{
			"【CONT】 If your climax area has a climax with [CHOICE] in its trigger icon, this card in all of your zones get [CHOICE] in the trigger icon. If there is a climax with [TREASURE] in its trigger icon, this card in all of your zones get [TREASURE] in the trigger icon. If there is a climax with [STANDBY] in its trigger icon, this card in all of your zones get [STANDBY] in the trigger icon. If there is a climax with [GATE] in its trigger icon, this card in all of your zones get [GATE] in the trigger icon.",
			"【AUTO】 【CLOCK】 Alarm If this card is the top card of your clock, and you have 4 or more 《World of Avatar》 characters, at the beginning of your climax phase, you may put the top card of your deck into your stock.",
		},
		Version: CardModelVersion,
	}

	card := extractData(siteConfigs[EN], doc.Clone())
	assertCardEquals(t, card, expectedCard)
}

func TestExtractDataEvent_en(t *testing.T) {
	event := `
<div class="p-cards__detail-wrapper">
	<div class="p-cards__detail-wrapper-inner">
		<div class="image"><img src="/wp/wp-content/images/cardimages/SS/WE41_E17.png" alt="The Day Yuji Disappeared" decoding="async">
		</div>
		<div class="p-cards__detail-textarea">
		<p class="number">SS/WE41-E17</p>
		<p class="ttl u-mt-14 u-mt-16-sp">The Day Yuji Disappeared</p>
		<div class="p-cards__detail-type u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Expansion</dt>
			<dd>[EX] Shakugan no Shana</dd>
			</dl>
			<dl>
			<dt>Traits</dt>
			<dd></dd>
			</dl>
			<dl>
			<dt>Card Type</dt>
			<dd>Event</dd>
			</dl>
			<dl>
			<dt>Rarity</dt>
			<dd>N</dd>
			</dl>
			<dl>
			<dt>Side</dt>
			<dd>
								<img src="/cardlist/partimages/w.gif" alt="" decoding="async">
								</dd>
			</dl>
			<dl>
			<dt>Color</dt>
			<dd><img src="/wp/wp-content/images/partimages/yellow.gif"></dd>
			</dl>
		</div>
		<div class="p-cards__detail-status u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Level</dt>
			<dd>2</dd>
			</dl>
			<dl>
			<dt>Cost</dt>
			<dd>1</dd>
			</dl>
			<dl>
			<dt>Power</dt>
			<dd>-</dd>
			</dl>
			<dl>
			<dt>Trigger</dt>
			<dd>－</dd>
			</dl>
			<dl>
			<dt>Soul</dt>
			<dd>-</dd>
			</dl>
		</div>
		<div class="p-cards__detail u-mt-22 u-mt-40-sp">
			<p>Search your deck for up to 2 《Flame》 characters, reveal them to your opponent, put them into your hand, choose 1 card in your hand, put it into your waiting room, and shuffle your deck.<br>Put this card into your memory.<br></p>
		</div>
		<div class="p-cards__detail-serif u-mt-22 u-mt-40-sp">
			<p>Yuji...</p>
		</div>
		<p class="p-cards__detail-copyrights u-mt-22 u-mt-40-sp">© YASHICHIRO TAKAHASHI/NOIZI ITO/ASCII MEDIA WORKS/「Shakugan no Shana F」committee</p>
		</div>
	</div>
</div>
`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(event))

	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[EN], doc.Clone())

	if card.Type != "EV" {
		t.Errorf("got %v: expected EV", card.Type)
	}

	if card.Name != "The Day Yuji Disappeared" {
		t.Errorf("got %v: expected The Day Yuji Disappeared", card.Name)
	}

	var expectedTrigger []string
	if !equalSlice(card.Triggers, expectedTrigger) {
		t.Errorf("got %v: expected %v", card.Triggers, expectedTrigger)
	}

	if !equalSlice(card.Traits, []string{}) {
		t.Errorf("got %v: expected empty", card.Traits)
	}

	if card.Level != "2" {
		t.Errorf("got %v: expected 2", card.Level)
	}

	if card.Color != "YELLOW" {
		t.Errorf("got %v: expected YELLOW", card.Color)
	}

	if card.Soul != "0" {
		t.Errorf("got %v: expected ''", card.Soul)
	}

	if card.Power != "0" {
		t.Errorf("got %v: expected 0", card.Power)
	}
}

func TestExtractDataCX_en(t *testing.T) {
	climax := `
<div class="p-cards__detail-wrapper">
	<div class="p-cards__detail-wrapper-inner">
		<div class="image"><img src="/wp/wp-content/images/cardimages/SS/WE41_E59SHP.png" alt="Direct Confrontation!" decoding="async">
		</div>
		<div class="p-cards__detail-textarea">
		<p class="number">SS/WE41-E59SHP</p>
		<p class="ttl u-mt-14 u-mt-16-sp">Direct Confrontation!</p>
		<div class="p-cards__detail-type u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Expansion</dt>
			<dd>[EX] Shakugan no Shana</dd>
			</dl>
			<dl>
			<dt>Traits</dt>
			<dd></dd>
			</dl>
			<dl>
			<dt>Card Type</dt>
			<dd>Climax</dd>
			</dl>
			<dl>
			<dt>Rarity</dt>
			<dd>SHP</dd>
			</dl>
			<dl>
			<dt>Side</dt>
			<dd>
								<img src="/cardlist/partimages/w.gif" alt="" decoding="async">
								</dd>
			</dl>
			<dl>
			<dt>Color</dt>
			<dd><img src="/wp/wp-content/images/partimages/blue.gif"></dd>
			</dl>
		</div>
		<div class="p-cards__detail-status u-mt-22 u-mt-40-sp">
			<dl>
			<dt>Level</dt>
			<dd>-</dd>
			</dl>
			<dl>
			<dt>Cost</dt>
			<dd>-</dd>
			</dl>
			<dl>
			<dt>Power</dt>
			<dd>-</dd>
			</dl>
			<dl>
			<dt>Trigger</dt>
			<dd><img src="/wp/wp-content/images/partimages/soul.gif"><img src="/wp/wp-content/images/partimages/gate.gif"></dd>
			</dl>
			<dl>
			<dt>Soul</dt>
			<dd>-</dd>
			</dl>
		</div>
		<div class="p-cards__detail u-mt-22 u-mt-40-sp">
			<p>【CONT】 All of your characters get +1000 power and +1 soul.<br>(<img src="/wp/wp-content/images/partimages/gate.gif">: When this card triggers, you may choose 1 climax in your waiting room, and return it to your hand)<br></p>
		</div>
		<div class="p-cards__detail-serif u-mt-22 u-mt-40-sp">
			<p>Flow inside, O energy.</p>
		</div>
		<p class="p-cards__detail-copyrights u-mt-22 u-mt-40-sp">© YASHICHIRO TAKAHASHI/NOIZI ITO/ASCII MEDIA WORKS/「SHAKUGAN NO ShanaⅡ」COMMITTEE/MBS</p>
		</div>
	</div>
</div>
`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(climax))
	if err != nil {
		t.Fatal(err)
	}

	card := extractData(siteConfigs[EN], doc.Clone())

	if card.Type != "CX" {
		t.Errorf("got %v: expected CX", card.Type)
	}

	if card.Name != "Direct Confrontation!" {
		t.Errorf("got %v: expected Direction Confrontation!", card.Name)
	}

	if card.Color != "BLUE" {
		t.Errorf("got %v: expected BLUE", card.Color)
	}

	if card.Soul != "0" {
		t.Errorf("got %v: expected ''", card.Soul)
	}

	if card.Level != "0" {
		t.Errorf("got %v: expected 0", card.Level)
	}

	if card.Cost != "0" {
		t.Errorf("got %v: expected 0", card.Cost)
	}

	expectedTrigger := []string{"SOUL", "GATE"}
	if !equalSlice(card.Triggers, expectedTrigger) {
		t.Errorf("got %v: expected %v", card.Triggers, expectedTrigger)
	}

	expectedAbility := []string{
		"【CONT】 All of your characters get +1000 power and +1 soul.",
		"([GATE]: When this card triggers, you may choose 1 climax in your waiting room, and return it to your hand)",
	}
	if !equalSlice(card.Abilities, expectedAbility) {
		t.Errorf("Incorrect ability. Got %v, want %v", card.Abilities, expectedAbility)
	}
}
