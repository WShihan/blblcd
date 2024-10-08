package core

import (
	"blblcd/model"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

var (
	UserAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0"
	Origin    string = "https://www.bilibili.com"
	Host      string = "https://www.bilibili.com"
)

func FetchComment(oid string, next int, order int, cookie string) (data model.CommentResponse, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("爬取评论失败,oid:%s，第%d页失败", oid, next))
			slog.Error(fmt.Sprint(err))
		}
	}()
	slog.Info(cookie)
	client := http.Client{}
	payload := strings.NewReader("")

	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("sort", fmt.Sprint(order))
	params.Set("nohot", "1")
	params.Set("ps", "20")
	params.Set("pn", fmt.Sprint(next))

	url := "https://api.bilibili.com/x/v2/reply?" + params.Encode()
	newUrl, err := SignAndGenerateURL(url, cookie)
	if err != nil {
		slog.Error(err.Error())
	}

	req, err := http.NewRequest("GET", newUrl, payload)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	req.Header.Add("User-agent", UserAgent)
	req.Header.Add("Origin", Origin)
	req.Header.Add("Host", Host)
	req.Header.Add("Referer", "https://www.bilibili.com/video/BV12u411g7go/?spm_id_from=333.788.top_right_bar_window_history.content.click&vd_source=10d0f86227f3c318f8237345caac47c8")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Cookie", cookie)

	res, err := client.Do(req)
	if err != nil {
		slog.Error("Erro:" + err.Error())
	}
	body := res.Body
	defer body.Close()
	dataStr, _ := io.ReadAll(res.Body)
	json.Unmarshal(dataStr, &data)
	slog.Info(fmt.Sprintf("完成评论获取，oid: %s, 第%d页", oid, next))
	return

}

func FetchSubComment(oid string, rpid int64, next int, cookie string) (data model.CommentResponse, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("爬取评论失败,oid:%s，第%d页失败", oid, next))
			slog.Error(fmt.Sprint(err))
		}
	}()
	client := http.Client{}
	payload := strings.NewReader("")

	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("root", fmt.Sprint(rpid))
	params.Set("ps", "20")
	params.Set("pn", fmt.Sprint(next))

	url := "https://api.bilibili.com/x/v2/reply/reply?" + params.Encode()
	newUrl, err := SignAndGenerateURL(url, cookie)
	if err != nil {
		slog.Error(err.Error())
	}

	req, err := http.NewRequest("GET", newUrl, payload)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	req.Header.Add("User-agent", UserAgent)
	req.Header.Add("Origin", Origin)
	req.Header.Add("Host", Host)
	req.Header.Add("Referer", "https://www.bilibili.com/video/BV12u411g7go/?spm_id_from=333.788.top_right_bar_window_history.content.click&vd_source=10d0f86227f3c318f8237345caac47c8")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Cookie", cookie)

	res, err := client.Do(req)
	if err != nil {
		slog.Error("Erro:" + err.Error())
	}
	body := res.Body
	defer body.Close()
	dataStr, _ := io.ReadAll(res.Body)
	json.Unmarshal(dataStr, &data)
	slog.Info(fmt.Sprintf("完成子评论获取，oid: %s, 第%d页", oid, next))
	return

}
