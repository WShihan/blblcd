package store

import (
	"blblcd/client"
	"blblcd/model"
	"blblcd/utils"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func fetchImage(imageUrl string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	response, err := client.
		Client.
		R().
		SetHeader("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8").
		SetHeader("Priority", "i").
		SetHeader("Sec-Fetch-Dest", "image").
		SetHeader("Sec-Fetch-Mode", "no-cors").
		SetHeader("Sec-Fetch-Site", "cross-site").
		SetHeader("Sec-Fetch-Storage-Access", "none").
		SetOutput(buffer).
		Get(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s for URL: %s", response.Status, imageUrl)
	}

	return buffer.Bytes(), nil
}

func DownloadImg(sem chan struct{}, mu *sync.Mutex, imageUrl string, output string, prefix string) {
	imgName := strings.Split(imageUrl, "/")[len(strings.Split(imageUrl, "/"))-1]
	imgFileName := filepath.Join(output, prefix+"_"+imgName)
	if utils.FileOrPathExists(imgFileName) {
		return
	}

	imageData, err := func() ([]byte, error) {
		sem <- struct{}{}
		defer func() { <-sem }()
		return fetchImage(imageUrl)
	}()
	if err != nil {
		slog.Error(fmt.Sprintf("获取图片错误:%s", err))
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// 判断保存路径是否存在
	utils.PresetPath(output)

	// 创建本地文件
	outFile, err := os.OpenFile(imgFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			slog.Error(fmt.Sprintf("创建文件错误:%s", err))
		}
		return
	}
	defer outFile.Close()
	defer outFile.Sync()

	_, err = outFile.Write(imageData)
	if err != nil {
		slog.Error(fmt.Sprintf("保存图片错误:%s", err))
		return
	}

	slog.Info("图片下载成功：" + imageUrl)
}

func WriteImage(mu *sync.Mutex, uname string, pics []model.Picture, output string) {
	sem := make(chan struct{}, 5)
	wg := sync.WaitGroup{}
	for _, pic := range pics {
		wg.Add(1)
		go func(imageUrl string) {
			defer wg.Done()
			DownloadImg(sem, mu, imageUrl, output, uname)
		}(pic.Img_src)

		time.Sleep(time.Duration((rand.Float32() + 1) * 2e7))
	}
	wg.Wait()
}
