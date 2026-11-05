package client

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imroc/req/v3"
)

const DefaultMaxConcurrency = 5

var (
	Origin string = "https://www.bilibili.com"

	Client *req.Client

	// ErrRiskControl 表示 B站返回 412，通常是触发风控或该资源对当前 cookie 不可见。
	ErrRiskControl = errors.New("bilibili risk control (412)")

	semMu sync.RWMutex
	sem   chan struct{}

	// loggedIn 记录当前 cookie 是否为已登录状态，由 nav 接口结果更新。
	loggedIn atomic.Bool
)

func init() {
	Client = req.
		C().
		ImpersonateChrome().
		SetTimeout(10*time.Second).
		SetCommonHeader("Accept-Language", "zh-CN,zh;q=0.9").
		SetCommonHeader("Referer", "https://www.bilibili.com/")

	// 默认每个 host 只保留 2 个空闲连接，高并发时会导致频繁新建 TLS 连接。
	transport := Client.GetTransport()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 32
	transport.IdleConnTimeout = 90 * time.Second

	SetMaxConcurrency(DefaultMaxConcurrency)
}

// SetLoggedIn 更新当前 cookie 的登录状态。
func SetLoggedIn(v bool) {
	loggedIn.Store(v)
}

// IsLoggedIn 返回当前 cookie 是否已登录。
func IsLoggedIn() bool {
	return loggedIn.Load()
}

// SetMaxConcurrency 设置整个进程内同时进行的 HTTP 请求数。
// 应在开始爬取前调用；运行中调用也安全，已在途的请求不受影响。
func SetMaxConcurrency(n int) {
	if n < 1 {
		n = 1
	}
	s := make(chan struct{}, n)
	semMu.Lock()
	sem = s
	semMu.Unlock()
}

func acquire() chan struct{} {
	semMu.RLock()
	s := sem
	semMu.RUnlock()
	if s == nil {
		return nil
	}
	s <- struct{}{}
	return s
}

func release(s chan struct{}) {
	if s != nil {
		<-s
	}
}

// Do 在全局并发限制内执行一次请求，避免多视频、多楼中楼并发时请求数失控。
func Do[T any](fn func() (T, error)) (T, error) {
	s := acquire()
	defer release(s)
	return fn()
}

// Get 在全局并发限制内发起一次 GET 请求。
// successResult 非 nil 时会把响应 JSON 解析进去；headers 会覆盖同名请求头。
func Get(url string, headers map[string]string, successResult any) (*req.Response, error) {
	return Do(func() (*req.Response, error) {
		r := Client.R()
		if successResult != nil {
			r.SetSuccessResult(successResult)
		}
		for _, k := range sortedKeys(headers) {
			r.SetHeader(k, headers[k])
		}
		return r.Get(url)
	})
}

// sortedKeys 固定请求头写入顺序，避免 map 随机顺序影响浏览器指纹。
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
