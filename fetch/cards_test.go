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

func TestRecentSwitch(t *testing.T) {
	expectedExpension := []string{
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
	recentTasks := getTasksForRecentReleases(doc)
	if len(recentTasks) != 8 {
		t.Errorf("Should be equal to 8: %v", recentTasks)
	}

	for _, task := range recentTasks {
		expansion := task.urlValues.Get("expansion")
		if slices.Contains(expectedExpension, expansion) == false {
			t.Errorf("Did not expect %v expansion", expansion)
		}
	}
}
