package core

import (
	"sync"
	"testing"
)

var (
	cookie = `buvid3=C1D23DC8-915B-1365-74E3-E1370E33656B00176infoc; b_nut=1710602900; i-wanna-go-back=-1; _uuid=2C10E8CC5-3B11-A9AC-81B5-E932DE2310610500418infoc; enable_web_push=DISABLE; buvid4=BB9A5B3A-FD02-75A4-9E30-7DAD735E85E001253-024031615-qa58iDNdnMeppfWVlzC2Gg%3D%3D; header_theme_version=CLOSE; DedeUserID=479424003; DedeUserID__ckMd5=3095bad518dcee61; rpdid=|(umRkuY)J||0J'u~u|Y|Rm)R; CURRENT_FNVAL=4048; buvid_fp_plain=undefined; b_ut=5; FEED_LIVE_VERSION=V_DYN_LIVING_UP; hit-dyn-v2=1; PVID=1; home_feed_column=5; fingerprint=74fcedc2cfd24a790b3329dce34bbe3d; CURRENT_QUALITY=80; buvid_fp=74fcedc2cfd24a790b3329dce34bbe3d; SESSDATA=9f9fd6cd%2C1732638111%2C5660f%2A51CjAF5emcapoYjxXEQKeUou154cH6yUeEDqAsDuVPTMG_s1szsfZ5jGJtLxBtGU4Bp3MSVlNOaFpkcnJPTWdsWVZlZjVJSHFfM2FfdERqTE1xR2lzR3g0cTA0XzVnNkRqYmVuVW5oMEdnWUsxa2NCbUlIYzBoVkRnZWxOc2NUdy1LaHdkc3AzTzZRIIEC; bili_jct=23f94e856c62646b8d769e39085aa920; sid=4pvk7l33; browser_resolution=1440-738; bsource=search_google; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTc0MTc5MzksImlhdCI6MTcxNzE1ODY3OSwicGx0IjotMX0.8zbaul3kh-yXZUk11eWnvqFe-JXk1c3s--8QNtCFI8M; bili_ticket_expires=1717417879; bp_t_offset_479424003=937732441859162243; b_lsid=71AE9432_18FD06B9441`
)

func GetCMT(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	FetchComment("1205445487", 17, 2, cookie)

}
func TestFetchCmt(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go GetCMT(&wg)
	wg.Wait()
	// fmt.Println(FetchComment("1205445487", 17, 2, cookie))
}
