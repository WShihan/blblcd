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

func FetchComment(oid string, next int, mode int, cookie string) (data model.CommentResponse, err error) {
	client := http.Client{}
	payload := strings.NewReader("")

	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("mode", fmt.Sprint(mode))
	params.Set("next", fmt.Sprint(next))

	url := "https://api.bilibili.com/x/v2/reply/wbi/main?" + params.Encode()
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
	fmt.Println(res.Status)
	fmt.Println(newUrl)
	slog.Info("finish vidoe comment oid:" + fmt.Sprint(oid) + ", index:" + fmt.Sprint(next))
	return

}
