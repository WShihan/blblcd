package main

import (
	"blblcd/core"
	"blblcd/model"
	"blblcd/utils"
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
)

func main() {
	cookie := flag.String("cookie", "cookie.text", "保存cookie的文件，类型为text")
	mid := flag.Int("mid", 0, "up主mid，爬取该up主若干视频评论")
	pages := flag.Int("pages", 3, "获取的页数,仅当指定mid时有效")
	skip := flag.Int("skip", 0, "跳过视频的页数，仅当指定mid时有效")
	vorder := flag.String("vorder", "pubdate", "爬取up主视频列表时排序方式，最新发布：pubdate最多播放：click最多收藏：stow")
	bvid := flag.String("bvid", "", "视频bvid，爬取该视频评论")
	corder := flag.Int("corder", 2, "爬取视频评论，排序方式，0 3：仅按热度，1：按热度+按时间，2：仅按时间")
	output := flag.String("output", "./output", "评论文件输出位置，默认程序运行位置")
	flag.Parse()

	opt := model.Option{
		Mid:    *mid,
		Pages:  *pages,
		Skip:   *skip,
		Vorder: *vorder,
		Bvid:   *bvid,
		Corder: *corder,
		Output: *output,
	}
	if !utils.FileOrPathExists(opt.Output) {
		os.MkdirAll(opt.Output, os.ModePerm)
	}

	if *cookie != "" {
		cookieStr := ""
		file, err := os.Open(*cookie)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			cookieStr += line
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}
		opt.Cookie = cookieStr
	}

	if opt.Mid != 0 {
		core.FindUserComments(&opt)
	} else if opt.Bvid != "" {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go core.FindComment(&wg, int(core.Bvid2Avid(fmt.Sprint(opt.Bvid))), &opt)
		wg.Wait()

	} else {
		fmt.Printf("请指定up主mid或视频bvid")
	}
}