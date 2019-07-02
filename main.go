package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	flag "github.com/spf13/pflag"
)

const (
	s01Host = "readcomicsonline.ru"
	s02Host = "www.tsumino.com"
)

var (
	outputDir string
	uilist    *widgets.List
	waitgroup *sync.WaitGroup
	concurr   int
	count     = 0
	keepJpg   bool
	useTermui bool
)

func main() {
	flagComicID := flag.String("comic-id", "", "readcomicsonline.ru comic ID")
	flagConcur := flag.Int("concurrency", 4, "The number of files to download simultaneously.")
	flagOutDir := flag.String("output-dir", "./results", "Output directory")
	flagKeepJpg := flag.Bool("keep-jpg", false, "Flag to keep/delete .jpg files of individual pages.")
	flagURL := flag.String("url", "", "URL of comic to download.")
	flagUseTermui := flag.Bool("use-termui", false, "")
	flag.Parse()

	//

	outputDir, _ = filepath.Abs(*flagOutDir)
	outputDir = strings.Replace(outputDir, string(filepath.Separator), "/", -1)
	outputDir += "/"
	log("Saving all files to", outputDir)

	concurr = *flagConcur

	wg := sync.WaitGroup{}
	waitgroup = &wg

	keepJpg = *flagKeepJpg

	useTermui = *flagUseTermui

	//

	if len(*flagComicID) > 0 {
		*flagURL = "https://readcomicsonline.ru/comic/" + *flagComicID
	}

	//

	urlO, err := url.Parse(*flagURL)
	if err != nil {
		return
	}

	if useTermui {
		if err := termui.Init(); err != nil {
			log("failed to initialize termui:", err)
		}
	}

	switch urlO.Host {
	case s01Host:
		outputDir += s01Host
		s01GetComic(strings.Split(urlO.Path, "/")[2])
	case s02Host:
		outputDir += s02Host
		s02GetComic(strings.Split(urlO.Path, "/")[3])
	default:
		closeTermui()
		log("Site not supported!")
		return
	}

	closeTermui()
}

func getDoc(urlS string) *goquery.Document {
	doc, _ := goquery.NewDocumentFromReader(doRequest(urlS).Body)
	return doc
}

func trim(x string) string {
	return strings.Trim(x, " \n\r\t")
}

func doesFileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func log(message ...interface{}) {
	fmt.Print("[" + time.Now().UTC().String()[0:19] + "] ")
	fmt.Println(message...)
}

func doRequest(urlS string) *http.Response {
	req, _ := http.NewRequest(http.MethodGet, urlS, strings.NewReader(""))
	res, _ := http.DefaultClient.Do(req)
	return res
}

func setRowText(row int, text string) {
	if !useTermui {
		log(text)
		return
	}
	uilist.Rows[row] = text
	termui.Render(uilist)
}

func findNextOpenRow(iss string) int {
	for i, v := range uilist.Rows {
		if strings.HasPrefix(v, "[x]") {
			uilist.Rows[i] = "[r] Reserved for " + iss
			return i
		}
	}
	return 0
}

func doesDirectoryExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !s.IsDir() {
		return false
	}
	return true
}

// F is an shorthand alias to fmt.Sprintf
func F(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func setupUIList(name string, id string) {
	if !useTermui {
		return
	}
	uilist = widgets.NewList()
	uilist.Title = "Comics-DL ---- " + name + " [" + id + "] ---- " + outputDir + " "
	uilist.Rows = strings.Split(strings.Repeat("[x] ,", concurr), ",")
	uilist.WrapText = false
	uilist.SetRect(0, 0, 100, concurr*2)
	termui.Render(uilist)
}

func packCbzArchive(dirIn string, fileOut string) {
	outf, _ := os.Create(fileOut)
	outz := zip.NewWriter(outf)
	files, _ := ioutil.ReadDir(dirIn)
	for _, item := range files {
		zw, _ := outz.Create(item.Name())
		bs, _ := ioutil.ReadFile(dirIn + item.Name())
		zw.Write(bs)
	}
	outz.Close()

	if !keepJpg {
		os.RemoveAll(dirIn)
	}
}

func closeTermui() {
	if useTermui {
		termui.Close()
	}
}
