package core

import (
	"blblcd/client"
	"blblcd/model"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
)

func FetchCount(oid string) (count int64, err error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/count?type=1&oid=%s", oid)
	data := model.CommentsCountResponse{}
	resp, err := client.
		Client.
		R().
		SetSuccessResult(&data).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsErrorState() {
		slog.Error("Erro:" + resp.String())
		return
	}

	count = data.Data.Count
	return
}

func FetchComment(oid string, order int, cookie string, offsetStr string) (data model.CommentResponse, err error) {
	var fmtOffsetStr string
	if offsetStr == "" {
		fmtOffsetStr = `{"offset":""}`
	} else {
		fmtOffsetStr = fmt.Sprintf(`{"offset":%q}`, offsetStr)
	}

	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("mode", strconv.Itoa(order))
	params.Set("pagination_str", fmtOffsetStr)
	params.Set("plat", "1")
	params.Set("seek_rpid", "")
	params.Set("web_location", "1315875")

	url := "https://api.bilibili.com/x/v2/reply/wbi/main?" + params.Encode()
	newUrl, err := SignAndGenerateURL(url, cookie)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	oidInt, err := strconv.ParseInt(oid, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	resp, err := client.
		Client.
		R().
		SetSuccessResult(&data).
		SetHeader("Accept", "application/json").
		SetHeader("Origin", client.Origin).
		SetHeader("Referer", client.Origin+"/video/"+Avid2Bvid(oidInt)+"/").
		SetHeader("Sec-Fetch-Dest", "empty").
		SetHeader("Sec-Fetch-Mode", "cors").
		SetHeader("Sec-Fetch-Site", "same-site").
		Get(newUrl)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsErrorState() {
		slog.Error("Erro:" + resp.String())
		return
	}
	return
}

func FetchSubComment(oid string, rpid int64, next int, cookie string) (data model.CommentResponse, err error) {
	params := url.Values{}
	params.Set("oid", oid)
	params.Set("type", "1")
	params.Set("root", fmt.Sprint(rpid))
	params.Set("ps", "20")
	params.Set("pn", fmt.Sprint(next))
	params.Set("web_location", "333.788")

	url := "https://api.bilibili.com/x/v2/reply/reply?" + params.Encode()

	oidInt, err := strconv.ParseInt(oid, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	res, err := client.
		Client.
		R().
		SetHeader("Origin", client.Origin).
		SetHeader("Referer", client.Origin+"/video/"+Avid2Bvid(oidInt)+"/").
		SetHeader("Sec-Fetch-Dest", "empty").
		SetHeader("Sec-Fetch-Mode", "cors").
		SetHeader("Sec-Fetch-Site", "same-site").
		SetHeader("Cookie", cookie).
		Get(url)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	err = res.UnmarshalJson(&data)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	return
}
