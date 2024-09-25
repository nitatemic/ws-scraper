package fetch

import (
	"os"
	"slices"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// func TestGetLastPage(t *testing.T) {
// 	f, err := os.Open("mockws/bd.html")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer f.Close()

// 	doc, err := goquery.NewDocumentFromReader(f)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	last := getLastPage(doc)
// 	if last != 69 {
// 		t.Errorf("%v is not last", last)
// 	}
// }

func TestRecentSwitch_en(t *testing.T) {
	expectedExpansion := []string{
		"228",
		"227",
		"226",
		"225",
	}
	f, err := os.Open("mockws-en/recent.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}
	recentTasks := getTasksForRecentReleases(siteConfigs[English], doc)
	if len(recentTasks) != len(expectedExpansion) {
		t.Errorf("Didn't get enough tasks. Got %d, want %d", len(recentTasks), len(expectedExpansion))
	}

	for _, task := range recentTasks {
		expansion := task.urlValues.Get("expansion")
		if !slices.Contains(expectedExpansion, expansion) {
			t.Errorf("Did not expect %q expansion", expansion)
		}
	}
}

func TestRecentSwitch_jp(t *testing.T) {
	expectedExpansion := []string{
		"444",
		"439",
		"443",
		"442",
		"438",
		"437",
		"441",
		"440",
	}
	f, err := os.Open("mockws/recent.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}
	recentTasks := getTasksForRecentReleases(siteConfigs[Japanese], doc)
	if len(recentTasks) != 8 {
		t.Errorf("Should be equal to 8: %v", recentTasks)
	}

	for _, task := range recentTasks {
		expansion := task.urlValues.Get("expansion")
		if slices.Contains(expectedExpansion, expansion) == false {
			t.Errorf("Did not expect %q expansion", expansion)
		}
	}
}
