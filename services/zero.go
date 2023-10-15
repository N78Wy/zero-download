package services

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ZeroDownload struct {
	Client  *http.Client
	Cookie  string
	OutPath string
	Limit   int
}

type Comic struct {
	Title string
	Pages []Page
}

type Page struct {
	Name    string
	Total   int
	PageUrl string
	Urls    []string
}

func (zd *ZeroDownload) Requert(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.43")
	req.Header.Add("Cookie", zd.Cookie)
	res, err := zd.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (zd *ZeroDownload) DownloadComic(urls []string) {
	if err := os.MkdirAll(zd.OutPath, 0744); err != nil {
		log.Fatal(err)
	}

	for _, comicUrl := range urls {
		comic := zd.GetComicPageInfo(comicUrl)

		log.Printf("Start downloading/开始下载: %s", comic.Title)
		for _, page := range comic.Pages {
			zd.DownloadPage(page, zd.OutPath+"/"+comic.Title)
		}
		log.Printf("Download completed/下载完成: %s", comic.Title)
	}
}

func (zd *ZeroDownload) DownloadPage(page Page, path string) {
	if len(page.Urls) == 0 {
		log.Printf(`The chapter image list is empty, non members cannot view the last three chapters. You need to recharge the member, find the cookie, and fill it in config.json before re executing.
章节图片列表为空，非会员不能查看最后的3个章节，需要充值会员后找到cookie填入config.json里再重新执行: %s`, page.Name)
		return
	}

	if err := os.MkdirAll(path+"/"+page.Name, 0744); err != nil {
		log.Fatal(err)
	}

	dlUrls := page.Urls

	ch := make(chan struct{}, zd.Limit)
	for i := 0; i < zd.Limit; i++ {
		ch <- struct{}{}
	}

	wg := sync.WaitGroup{}
	for _, url := range dlUrls {
		<-ch
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			var retry int
			zd.DownloadImage(url, path+"/"+page.Name, &retry)
			ch <- struct{}{}
		}(url)
	}
	wg.Wait()
}

func (zd *ZeroDownload) DownloadImage(url, path string, retry *int) {
	imagePath := zd.getFullPath(url, path)

	_, err := os.Stat(imagePath)
	if err == nil {
		log.Printf("Skip existing images/跳过已存在的图片: %s", imagePath)
		return
	}

	resp, err := zd.Requert(url)
	if err != nil {
		zd.retryDownLoad(url, path, retry)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Download failed/下载图片失败: %s", resp.Status)
	}

	file, err := os.Create(imagePath)
	if err != nil {
		log.Fatalf("Created failed/创建图片失败: %s %s", imagePath, err.Error())
	}
	wt := bufio.NewWriter(file)
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()

		// Timed out images may not be downloaded completely and need to be deleted and downloaded again
		if os.IsTimeout(err) {
			log.Printf("Download Timeout/下载图片超时: %s", imagePath)
			if _, err := os.Stat(imagePath); err == nil {
				log.Printf("Preparing to delete potentially incomplete images/准备删除可能不完整的图片: %s", imagePath)
				time.Sleep(time.Second * 2)
				if err := os.Remove(imagePath); err != nil {
					log.Fatalf("Image deletion failed/图片删除失败: %s", err.Error())
				}
			}
		}

		zd.retryDownLoad(url, path, retry)
		return
	}
	wt.Flush()

	log.Printf("Success/下载图片成功: %s", imagePath)
}

func (zd *ZeroDownload) GetComicPageInfo(url string) *Comic {
	res, err := zd.Requert(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("request failure/请求失败: %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	title := doc.Find("title").First().Text()
	reg := regexp.MustCompile(`[ \s]+`)
	til := reg.ReplaceAll([]byte(title), []byte{})

	comic := &Comic{
		Pages: []Page{},
		Title: string(til),
	}

	log.Printf("Preparing to obtain chapter information/准备获取章节信息: %s", comic.Title)

	wg := &sync.WaitGroup{}
	pages := doc.Find(".uk-grid-collapse .muludiv a")
	comic.Pages = make([]Page, pages.Length())

	pages.Each(func(i int, s *goquery.Selection) {
		page := Page{
			Name: s.Text(),
		}

		href, _ := s.Attr("href")
		reg := regexp.MustCompile(`^(https|http|ftp)\:\/\/[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}`)
		matchArr := reg.FindStringSubmatch(url)
		if len(matchArr) <= 0 {
			log.Fatalf("URL extraction error, please check the URL format/url提取错误，请检查url格式")
		}
		baseUrl := matchArr[0]

		page.PageUrl = baseUrl + "/" + href

		wg.Add(1)

		go func(pageUrl string) {
			defer wg.Done()

			res, err := zd.Requert(page.PageUrl)
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			imgs := doc.Find(".uk-zjimg img")
			page.Total = imgs.Length()
			page.Urls = make([]string, page.Total)

			imgs.Each(func(i int, s *goquery.Selection) {
				if imageUrl, ok := s.Attr("src"); ok {
					page.Urls[i] = imageUrl
				}
			})

			log.Printf("%s, chapter/章节: %s, total/总数: %d", comic.Title, page.Name, page.Total)

			comic.Pages[i] = page
		}(page.PageUrl)
	})

	wg.Wait()

	return comic
}

func (zd *ZeroDownload) retryDownLoad(url, path string, retry *int) {
	imagePath := zd.getFullPath(url, path)
	log.Printf("Download failed/下载图片失败: %s, Retry/重试次数: %d, Retrying after 5 seconds/5s后进行重试", imagePath, *retry)
	if *retry > 8 {
		log.Printf("Exceeded retry count/超过重试次数: %s", imagePath)
	}
	*retry++
	time.Sleep(time.Second * 5)
	zd.DownloadImage(url, path, retry)
}

func (zd *ZeroDownload) getFullPath(url, path string) string {
	tempArr := strings.Split(url, "/")
	fileName := tempArr[len(tempArr)-1]
	return path + "/" + fileName
}
