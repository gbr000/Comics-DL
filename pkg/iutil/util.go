package iutil

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/The-Eye-Team/Comics-DL/pkg/idata"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/mbpp"
)

func DoSite(place *url.URL, rootDir string) {
	h, ok := idata.Hosts[place.Host]
	if !ok {
		return
	}
	id := strings.Split(place.Path, "/")[h.IDPathIndex]
	job := place.Host + " / " + id
	mbpp.CreateJob(job, h.DownloadFunc(place.Host, id, place.Path, rootDir+"/"+place.Host))
}

func GetDoc(urlS string) *goquery.Document {
	res, _ := http.Get(urlS)
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	return doc
}

func Trim(x string) string {
	return strings.Trim(x, " \n\r\t")
}

func PackCbzArchive(dirIn string, title string, bar *mbpp.BarProxy) {
	outf, _ := os.Create(dirIn + ".cbz")
	outz := zip.NewWriter(outf)
	files, _ := ioutil.ReadDir(dirIn)
	bar.AddToTotal(int64(len(files) + 2))
	for _, item := range files {
		zw, _ := outz.Create(item.Name())
		bs, _ := ioutil.ReadFile(dirIn + "/" + item.Name())
		zw.Write(bs)
		bar.Increment(1)
	}
	outz.Close()
	bar.Increment(1)
	if !idata.KeepJpg {
		os.RemoveAll(dirIn)
	}
	bar.Increment(1)
	bar.Increment(1)
}

func FixTitleForFilename(t string) string {
	n := Trim(t)
	n = strings.Replace(n, ":", "", -1)
	n = strings.Replace(n, "\\", "-", -1)
	n = strings.Replace(n, "/", "-", -1)
	n = strings.Replace(n, "*", "-", -1)
	n = strings.Replace(n, "?", "-", -1)
	n = strings.Replace(n, "\"", "-", -1)
	n = strings.Replace(n, "<", "-", -1)
	n = strings.Replace(n, ">", "-", -1)
	n = strings.Replace(n, "|", "-", -1)
	return n
}

func PadPgNum(n int) string {
	return fmt.Sprintf("%04d", n)
}

func ParseInt(s string) int {
	x, _ := strconv.ParseInt(s, 10, 32)
	return int(x)
}
