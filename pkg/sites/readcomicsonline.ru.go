package sites

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/The-Eye-Team/Comics-DL/pkg/idata"
	"github.com/The-Eye-Team/Comics-DL/pkg/itypes"
	"github.com/The-Eye-Team/Comics-DL/pkg/iutil"

	"github.com/PuerkitoBio/goquery"
	"github.com/nektro/go-util/alias"
	"github.com/nektro/go-util/mbpp"
	"github.com/nektro/go-util/util"
)

func init() {
	idata.Hosts["readcomicsonline.ru"] = itypes.HostVal{2, func(host string, id string, path string, outputDir string) func(*mbpp.BarProxy, *sync.WaitGroup) {
		return func(mbar *mbpp.BarProxy, _ *sync.WaitGroup) {
			//
			d := iutil.GetDoc("https://" + host + "/comic/" + id)
			s := d.Find("ul.chapters li")
			n := iutil.FixTitleForFilename(iutil.Trim(d.Find("h2.listmanga-header").Eq(0).Text()))
			mbar.AddToTotal(int64(s.Length()))
			s.Each(func(i int, el *goquery.Selection) {
				is0, _ := el.Children().First().Children().First().Attr("href")
				is1 := strings.Split(is0, "/")
				is2 := is1[len(is1)-1]
				is3, _ := url.ParseQuery("x=" + is2)
				//
				name := n
				issue := is3["x"][0]
				go mbpp.CreateJob(name+" / "+issue, func(jbar *mbpp.BarProxy, wg *sync.WaitGroup) {
					defer mbar.Increment(1)
					//
					dir2 := outputDir + "/" + name
					dir := dir2 + "/Issue " + issue
					finp := dir + ".cbz"

					if util.DoesFileExist(finp) {
						return
					}
					os.MkdirAll(dir2, os.ModePerm)
					if !util.DoesFileExist(finp) {
						os.MkdirAll(dir, os.ModePerm)
						jbar.AddToTotal(1)
						for j := 1; true; j++ {
							pth := dir + "/" + iutil.PadPgNum(j) + ".jpg"
							if util.DoesFileExist(pth) {
								continue
							}
							u := alias.F("https://readcomicsonline.ru/uploads/manga/%s/chapters/%s/%02d.jpg", id, issue, j)
							res, _ := http.Head(u)
							if res.StatusCode != 200 {
								break
							}
							jbar.AddToTotal(1)
							wg.Add(1)
							go mbpp.CreateDownloadJob(u, pth, wg, jbar)
						}
						wg.Wait()
						iutil.PackCbzArchive(dir, finp, jbar)
					}
				})
			})
		}
	}}
}