package store

import (
	"blblcd/model"
	"blblcd/utils"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

func DwonloadImg(wg *sync.WaitGroup, sem chan struct{}, imageUrl string, output string) {
	// 判断保存路径是否存在
	if !utils.FileOrPathExists(output) {
		os.MkdirAll(output, os.ModePerm)
	}
	// 发送 GET 请求
	response, err := http.Get(imageUrl)
	if err != nil {
		slog.Error(fmt.Sprintf("获取图片错误:%s", err))
		return
	}
	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("访问图片响应错误:%s", response.Status))
		return
	}
	// 创建本地文件
	imgName := strings.Split(imageUrl, "/")[len(strings.Split(imageUrl, "/"))-1]
	outFile, err := os.Create(output + "/" + imgName)
	if err != nil {
		slog.Error(fmt.Sprintf("创建文件错误:%s", err))
		return
	}
	defer func() {
		response.Body.Close()
		outFile.Close()
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("写入图片失败:%s", imageUrl))
			slog.Error(fmt.Sprint(err))
		}
		wg.Done()
		<-sem
	}()

	// 将响应体写入文件
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("报错图片错误:%s", err))
		return
	}

	slog.Info("图片下载成功：" + imageUrl)
}
func WriteImage(uname string, pics []model.Picture, output string) {
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, 5)
	for _, pic := range pics {
		wg.Add(1)
		sem <- struct{}{}
		go DwonloadImg(&wg, sem, pic.Img_src, output+"/"+uname)
	}
	wg.Wait()

}
