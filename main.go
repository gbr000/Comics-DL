package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/util"

	flag "github.com/spf13/pflag"
)

const (
	domain = "https://readcomicsonline.ru"
)

var (
	worker sync.WaitGroup
	count  = 0
)

func main() {
	flagComic := flag.String("comic-id", "", "")
	flagConcur := flag.Int("concurrency", 8, "The number of files to download simultaneously.")
	flag.Parse()

	id := *flagComic
	if len(id) == 0 {
		util.Log("Must send a valid comic ID")
		util.Log(">If you'd like to download https://readcomicsonline.ru/comic/justice-league-2016")
		util.Log(">then pass --comic-id justice-league-2016")
		return
	}
	util.Log("Saving comic:", id)

	ln := getDoc(domain + "/comic/" + id).Find("ul.chapters li").Length()
	for i := 0; i < ln; i++ {
		worker.Add(1)
		count++
		go getIssue(id, ln, i, &worker)
		if count <= *flagConcur {
			worker.Wait()
			count = 0
		}
	}
}

func getIssue(id string, ln, i int, wrk *sync.WaitGroup) {
	defer wrk.Done()
	for j := 0; true; j++ {
		dir := fmt.Sprintf("./results.ru/%s/Issue %02d/", id, i+1)
		os.MkdirAll(dir, os.ModePerm)
		pth := fmt.Sprintf("%s/%02d.jpg", dir, j+1)
		if util.DoesFileExist(pth) {
			continue
		}
		u := fmt.Sprintf("https://readcomicsonline.ru/uploads/manga/%s/chapters/%d/%02d.jpg", id, i+1, j+1)
		req, _ := http.NewRequest(http.MethodGet, u, strings.NewReader(""))
		res, _ := http.DefaultClient.Do(req)
		sc := res.StatusCode
		util.Log(sc, u, fmt.Sprintf("[issue %d/%d]", i+1, ln))
		if sc >= 400 {
			break
		}
		bys, _ := ioutil.ReadAll(res.Body)
		ioutil.WriteFile(pth, bys, os.ModePerm)
	}
}

func getDoc(lru string) *goquery.Document {
	req, _ := http.NewRequest(http.MethodGet, lru, strings.NewReader(""))
	res, _ := http.DefaultClient.Do(req)
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	return doc
}
