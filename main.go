package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/N78Wy/zero-download/src"
)

func main() {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatal(err)
	}
	logFile, err := os.OpenFile("./logs/"+time.Now().Local().Format("20060102")+".log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)

	conf := &src.Config{}
	conf.LoadConfig("./config.json")

	cli := src.NewClient()

	zd := &src.ZeroDownload{
		Cookie:  conf.Cookie,
		OutPath: conf.OutPath,
		Client:  cli,
		Limit:   conf.Limit,
	}

	zd.DownloadComic(conf.Urls)

	log.Printf("DONE!")
}
