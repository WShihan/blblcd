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

	"github.com/go-resty/resty/v2"
)

var (
	UserAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0"
	Origin    string = "https://www.bilibili.com"
	Host      string = "https://www.bilibili.com"
)

func FetchCount(oid string) (count int, err error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/count?type=1&oid=%s", oid)
	client := resty.New()
	data := model.CommentsCountResponse{}
	resp, err := client.R().
		SetResult(&data).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsError() {
		slog.Error("Erro:" + resp.String())
		return
	}

	count = data.Data.Count
	return
}

func FetchComment(oid string, next int, order int, cookie string, offsetStr string) (data model.CommentResponse, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("=====爬取主评论,oid:%s，第%d页失败=====", oid, next))
			slog.Error(fmt.Sprint(err))
		}
	}()
	client := resty.New()
	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("mode", "3")
	params.Set("plat", "1")
	params.Set("web_location", "1315875")
	params.Set("pagination_str", offsetStr)

	url := "https://api.bilibili.com/x/v2/reply/wbi/main?" + params.Encode()
	newUrl, err := SignAndGenerateURL(url, cookie)
	resp, err := client.R().
		SetResult(&data).
		SetHeader("Accept", "application/json").
		Get(newUrl)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsError() {
		slog.Error("Erro:" + resp.String())
		return
	}
	return

}

func FetchSubComment(oid string, rpid int64, next int, cookie string) (data model.CommentResponse, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("xxxxx爬取子评论,oid:%s，第%d页失败xxxxx", oid, next))
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
	slog.Info(fmt.Sprintf("xxxxx完成子评论获取，oid: %s, 第%d页xxxxx", oid, next))
	return

}
