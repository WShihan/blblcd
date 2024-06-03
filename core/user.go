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

func FetchVideoList(mid int, page int, order string, cookie string) (videoList model.VideoListResponse, err error) {
	api := "https://api.bilibili.com/x/space/wbi/arc/search?"
	params := url.Values{}
	params.Set("mid", fmt.Sprint(mid))
	params.Set("order", order)
	params.Set("platform", "web")
	params.Set("pn", fmt.Sprint(page))
	params.Set("ps", "30")
	params.Set("tid", "0")

	client := http.Client{}
	crypedApi, _ := SignAndGenerateURL(api+params.Encode(), cookie)

	req, _ := http.NewRequest("GET", crypedApi, strings.NewReader(""))

	req.Header.Add("Origin", "https://space.bilibili.com")
	req.Header.Add("Host", Host)
	req.Header.Add("Referer", Origin)
	req.Header.Add("User-agent", UserAgent)
	req.Header.Add("Cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("parse json error:" + err.Error())
	}
	defer resp.Body.Close()

	jsonByte, _ := io.ReadAll(resp.Body)
	slog.Info(resp.Status)
	json.Unmarshal(jsonByte, &videoList)
	// fmt.Printf("%v", videoList)
	slog.Info("finish vidoe list index:" + fmt.Sprint(page))
	return

}
