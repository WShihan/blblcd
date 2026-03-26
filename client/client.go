package client

import (
	"github.com/imroc/req/v3"
	"time"
)

var (
	Origin string = "https://www.bilibili.com"

	Client *req.Client
)

func init() {
	Client = req.
		C().
		ImpersonateChrome().
		SetTimeout(5*time.Second).
		SetCommonHeader("Accept-Language", "zh-CN,zh;q=0.9").
		SetCommonHeader("Referer", "https://www.bilibili.com/")
}
