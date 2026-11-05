package core

import (
	"blblcd/client"
	"blblcd/model"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

func riskControlError(statusCode int) error {
	if statusCode == http.StatusPreconditionFailed {
		return client.ErrRiskControl
	}
	return nil
}

func FetchCount(oid string) (count int64, err error) {
	reqUrl := fmt.Sprintf("https://api.bilibili.com/x/v2/reply/count?type=1&oid=%s", oid)
	data := model.CommentsCountResponse{}
	resp, err := client.Get(reqUrl, map[string]string{
		"Accept": "application/json",
	}, &data)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsErrorState() {
		err = fmt.Errorf("请求评论数失败: %s", resp.Status)
		slog.Error(err.Error())
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
	// 接口单页最多返回 30 条（实际 29 条），比默认 20 条少约 1.5 倍翻页次数。
	params.Set("ps", "30")
	params.Set("seek_rpid", "")
	params.Set("web_location", "1315875")

	reqUrl := "https://api.bilibili.com/x/v2/reply/wbi/main?" + params.Encode()
	newUrl, err := SignAndGenerateURL(reqUrl, cookie)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	oidInt, err := strconv.ParseInt(oid, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	headers := map[string]string{
		"Accept":         "application/json",
		"Origin":         client.Origin,
		"Referer":        client.Origin + "/video/" + Avid2Bvid(oidInt) + "/",
		"Sec-Fetch-Dest": "empty",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Site": "same-site",
	}
	if client.IsLoggedIn() {
		// B站只在登录态下返回 reply_control.location（IP属地）。
		// 未登录时带上失效 cookie 反而可能只返回少量评论，因此仅在确认登录后发送。
		headers["Cookie"] = cookie
	}
	resp, err := client.Get(newUrl, headers, &data)
	if err != nil {
		slog.Error("Erro:" + err.Error())
		return
	}
	if resp.IsErrorState() {
		if rcErr := riskControlError(resp.StatusCode); rcErr != nil {
			err = rcErr
		} else {
			err = fmt.Errorf("请求评论失败: %s", resp.Status)
		}
		slog.Error(err.Error())
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

	reqUrl := "https://api.bilibili.com/x/v2/reply/reply?" + params.Encode()

	oidInt, err := strconv.ParseInt(oid, 10, 64)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	res, err := client.Get(reqUrl, map[string]string{
		"Origin":         client.Origin,
		"Referer":        client.Origin + "/video/" + Avid2Bvid(oidInt) + "/",
		"Sec-Fetch-Dest": "empty",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Site": "same-site",
		"Cookie":         cookie,
	}, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	if res.IsErrorState() {
		if rcErr := riskControlError(res.StatusCode); rcErr != nil {
			// 412 通常是该楼中楼对当前 cookie 不可见或触发风控，交给上层用预览回复兜底。
			return data, rcErr
		}
		err = fmt.Errorf("请求子评论失败: %s", res.Status)
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
