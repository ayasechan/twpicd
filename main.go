package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

// 限制并发
const maxGoroutines = 50

var Version string

var (
	api         *anaconda.TwitterApi
	screen_name = flag.String("name", "", "用户的名字")
)

type ImageInfo struct {
	URL string
	ID  string
}

func main() {

	flag.Parse()

	if *screen_name == "" {
		flag.Usage()
		os.Exit(-1)
	}

	// 等待所有go程结束
	api = anaconda.NewTwitterApiWithCredentials(
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_SECRET"),
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_SECRET"),
	)

	images, err := GetImages()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 创建文件夹
	os.Mkdir(*screen_name, os.ModePerm)

	var wg sync.WaitGroup
	guard := make(chan int, maxGoroutines)
	defer close(guard)

	var downloadFinishCount int64
	for _, img := range images {
		wg.Add(1)
		guard <- 1
		go func(img ImageInfo) {
			defer func() {
				wg.Done()
				<-guard
			}()
			// originFileUrl := fmt.Sprintf("%s?name=orig", url)
			pngUrl := strings.Replace(img.URL, ".jpg", ".png", 1)
			fname := fmt.Sprintf(
				"%s_%s%s",
				img.ID,
				strconv.FormatInt(time.Now().UnixNano(), 10),
				filepath.Ext(pngUrl),
			)

			fpath := filepath.Join(*screen_name, fname)

			err := DownloadFile(fpath, pngUrl)
			if err != nil {
				fmt.Println(err)
				return
			}
			atomic.AddInt64(&downloadFinishCount, 1)
			fmt.Printf("%d/%d %s\n", downloadFinishCount, len(images), fpath)
		}(img)
	}

	wg.Wait()
	fmt.Println("end!")
}

func GetImages() ([]ImageInfo, error) {
	// 设置请求参数
	v := url.Values{}
	v.Set("screen_name", *screen_name)
	v.Set("count", "200")
	v.Set("exclude_replies", "true")
	v.Set("include_rts", "false")

	var maxID int64
	images := []ImageInfo{}
	pageCount := 0

	for {
		if maxID != 0 {
			v.Set("max_id", strconv.FormatInt(maxID, 10))
		}
		searchResult, err := api.GetUserTimeline(v)
		if err != nil {
			return nil, err
		}
		pageCount += 1
		fmt.Println("request page", pageCount)

		if len(searchResult) < 1 {
			break
		}

		maxID = searchResult[len(searchResult)-1].Id - 1

		for _, v := range searchResult {
			for _, m := range v.ExtendedEntities.Media {
				images = append(images, ImageInfo{
					URL: m.Media_url_https,
					ID:  v.IdStr,
				})
			}
		}
		time.Sleep(time.Second)
	}
	return images, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
	client := new(http.Client)

	// Get the data
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode, url)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
