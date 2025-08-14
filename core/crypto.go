package core

import (
	"blblcd/client"
	"crypto/md5"
	"encoding/hex"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

var (
	mixinKeyEncTab = []int{
		46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
		33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
		61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
		36, 20, 34, 44, 52,
	}
	cache          sync.Map
	lastUpdateTime time.Time

	XOR_CODE = int64(23442827791579)
	MAX_CODE = int64(2251799813685247)
	CHARTS   = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
)

func SignAndGenerateURL(urlStr string, cookie string) (string, error) {
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	imgKey, subKey := getWbiKeysCached(cookie)
	query := urlObj.Query()
	params := map[string]string{}
	for k, v := range query {
		params[k] = v[0]
	}
	newParams := encWbi(params, imgKey, subKey)
	for k, v := range newParams {
		query.Set(k, v)
	}
	urlObj.RawQuery = query.Encode()
	newUrlStr := urlObj.String()
	return newUrlStr, nil
}

func encWbi(params map[string]string, imgKey, subKey string) map[string]string {
	mixinKey := getMixinKey(imgKey + subKey)
	currTime := strconv.FormatInt(time.Now().Round(time.Second).Unix(), 10)
	params["wts"] = currTime

	// Build URL parameters
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	queryStr := query.Encode() // sorted by key

	// Calculate w_rid
	hash := md5.Sum([]byte(queryStr + mixinKey))
	params["w_rid"] = hex.EncodeToString(hash[:])
	return params
}

func getMixinKey(orig string) string {
	var str strings.Builder
	for _, v := range mixinKeyEncTab {
		if v < len(orig) {
			str.WriteByte(orig[v])
		}
	}
	return str.String()[:32]
}

func updateCache(cookie string) {
	if time.Since(lastUpdateTime).Minutes() < 10 {
		return
	}
	imgKey, subKey := getWbiKeys(cookie)
	cache.Store("imgKey", imgKey)
	cache.Store("subKey", subKey)
	lastUpdateTime = time.Now()
}

func getWbiKeysCached(cookie string) (string, string) {
	updateCache(cookie)
	imgKeyI, _ := cache.Load("imgKey")
	subKeyI, _ := cache.Load("subKey")
	return imgKeyI.(string), subKeyI.(string)
}

func getWbiKeys(cookie string) (string, string) {
	resp, err := client.Client.R().SetHeader("Cookie", string(cookie)).Get("https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		slog.Error("Error sending request", "err", err)
		return "", ""
	}
	json := gjson.Parse(resp.String())
	imgURL := json.Get("data.wbi_img.img_url").String()
	subURL := json.Get("data.wbi_img.sub_url").String()
	imgKey := strings.Split(strings.Split(imgURL, "/")[len(strings.Split(imgURL, "/"))-1], ".")[0]
	subKey := strings.Split(strings.Split(subURL, "/")[len(strings.Split(subURL, "/"))-1], ".")[0]
	return imgKey, subKey
}

func swapString(s string, x, y int) string {
	chars := []rune(s)
	chars[x], chars[y] = chars[y], chars[x]
	return string(chars)
}

func Bvid2Avid(bvid string) (avid int64) {
	s := swapString(swapString(bvid, 3, 9), 4, 7)
	bv1 := string([]rune(s)[3:])
	temp := int64(0)
	for _, c := range bv1 {
		idx := strings.IndexRune(CHARTS, c)
		temp = temp*int64(58) + int64(idx)
	}
	avid = (temp & MAX_CODE) ^ XOR_CODE
	return
}

func Avid2Bvid(avid int64) (bvid string) {
	arr := [12]string{"B", "V", "1"}
	bvIdx := len(arr) - 1
	temp := (avid | (MAX_CODE + 1)) ^ XOR_CODE
	for temp > 0 {
		idx := temp % 58
		arr[bvIdx] = string(CHARTS[idx])
		temp /= 58
		bvIdx--
	}
	raw := strings.Join(arr[:], "")
	bvid = swapString(swapString(raw, 3, 9), 4, 7)
	return
}
