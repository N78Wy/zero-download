package main

import (
	"fmt"
	"log"
	"os"

	service "github.com/N78Wy/zero-download/services"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate)

	conf := &service.Config{}
	conf.LoadConfig("./config.json")

	cli := service.NewClient()

	zd := &service.ZeroDownload{
		Cookie:  conf.Cookie,
		OutPath: conf.OutPath,
		Client:  cli,
		Limit:   conf.Limit,
	}

	zd.DownloadComic(conf.Urls)

	log.Printf("DONE! Press any key to exit!/下载完成! 任意按键退出!")
	fmt.Scanln()
}
