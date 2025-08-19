package core

import (
	"blblcd/client"
	"blblcd/model"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
)

func FetchVideoList(mid int64, page int, order string, cookie string) (videoList model.VideoListResponse, err error) {
	api := "https://api.bilibili.com/x/space/wbi/arc/search?"
	params := url.Values{}
	params.Set("mid", fmt.Sprint(mid))
	params.Set("order", order)
	params.Set("platform", "web")
	params.Set("pn", fmt.Sprint(page))
	params.Set("ps", "30")
	params.Set("tid", "0")

	crypedApi, _ := SignAndGenerateURL(api+params.Encode(), cookie)

	resp, err := client.
		Client.
		R().
		SetHeader("Origin", "https://space.bilibili.com").
		SetHeader("Referer", client.Origin+"/"+strconv.FormatInt(mid, 10)).
		SetHeader("Cookie", cookie).
		Get(crypedApi)
	if err != nil {
		slog.Error("get json error", "err", err.Error())
	}
	if resp.IsErrorState() {
		slog.Error("get json state error", "body", resp.String())
		return
	}

	err = resp.Unmarshal(&videoList)
	if err != nil {
		slog.Error("parse json error", "err", err.Error())
	}

	slog.Info(fmt.Sprintf("爬取up主视频列表成功,mid:%d，第%d页", mid, page))
	return
}
