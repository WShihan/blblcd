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

	crypedApi, err := SignAndGenerateURL(api+params.Encode(), cookie)
	if err != nil {
		return videoList, err
	}

	resp, err := client.Get(crypedApi, map[string]string{
		"Origin":  "https://space.bilibili.com",
		"Referer": client.Origin + "/" + strconv.FormatInt(mid, 10),
		"Cookie":  cookie,
	}, nil)
	if err != nil {
		slog.Error("get json error", "err", err.Error())
		return
	}
	if resp.IsErrorState() {
		err = fmt.Errorf("请求视频列表失败: %s", resp.Status)
		slog.Error("get json state error", "err", err.Error(), "body", resp.String())
		return
	}

	err = resp.Unmarshal(&videoList)
	if err != nil {
		slog.Error("parse json error", "err", err.Error())
	}

	slog.Info(fmt.Sprintf("爬取up主视频列表成功,mid:%d，第%d页", mid, page))
	return
}
