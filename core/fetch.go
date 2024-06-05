package core

import (
	"blblcd/model"
	"blblcd/store"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func FindComment(sem chan struct{}, wg *sync.WaitGroup, avid int, opt *model.Option) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("爬取视频：%d失败", avid))
			slog.Error(fmt.Sprint(err))
		}
		wg.Done()
		<-sem
	}()

	oid := strconv.Itoa(avid)
	round := 1
	recordedMap := make(map[int64]bool)
	locationStat := map[string]int{}
	for {
		replyCollection := []model.ReplyItem{}

		// 停顿
		delay := (rand.Float32() + 1) * 1e9
		time.Sleep(time.Duration(delay))
		slog.Info(fmt.Sprintf("爬取视频评论%s", oid))
		cmtInfo, _ := FetchComment(oid, round, opt.Corder, opt.Cookie)
		round++
		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求评论失败，视频%s，第%d页失败", oid, round))
			slog.Error(cmtInfo.Message)
			break
		}
		total := cmtInfo.Data.Page.Acount
		if len(cmtInfo.Data.Replies) != 0 && len(replyCollection) < total {
			replyCollection = append(replyCollection, cmtInfo.Data.Replies...)
			for _, k := range cmtInfo.Data.Replies {
				if len(k.Replies) > 0 {
					replyCollection = append(replyCollection, k.Replies...)
				}
			}
			if len(cmtInfo.Data.TopReplies) != 0 {
				replyCollection = append(replyCollection, cmtInfo.Data.TopReplies...)
				for _, k := range cmtInfo.Data.TopReplies {
					if len(k.Replies) > 0 {
						replyCollection = append(replyCollection, k.Replies...)
					}
				}
			}
		}
		if len(replyCollection) == 0 {
			slog.Info(fmt.Sprintf("-----视频%s，第%d页未获取到评论-----", oid, round))
			break
		}

		var cmtCollection = []model.Comment{}
		for _, k := range replyCollection[:] {
			if _, ok := recordedMap[k.Rpid]; !ok {
				cmt := NewCMT(&k)
				recordedMap[cmt.Rpid] = true
				cmtCollection = append(cmtCollection, cmt)
				if opt.Geojson {
					locationStat[cmt.Location] += 1
				}
			} else {
				slog.Info(fmt.Sprintf("评论%d已存在，跳过", k.Rpid))
			}

		}
		ok := store.Save2CSV(oid, cmtCollection, opt.Output)
		if ok {
			slog.Info(fmt.Sprintf("-----爬取评论%s，第%d页完成！----", oid, round))
		}
	}
	if opt.Geojson {
		store.WriteGeoJSON(locationStat, oid, opt.Output)
	}

}

func NewCMT(item *model.ReplyItem) model.Comment {
	return model.Comment{
		Uname:         item.Member.Uname,
		Sex:           item.Member.Sex,
		Content:       item.Content.Message,
		Rpid:          item.Rpid,
		Oid:           item.Oid,
		Bvid:          Avid2Bvid(int64(item.Oid)),
		Mid:           item.Mid,
		Parent:        item.Parent,
		Ctime:         item.Ctime,
		Like:          item.Like,
		Following:     item.ReplyControl.Following,
		Current_level: item.Member.LevelInfo.CurrentLevel,
		Location:      strings.Replace(item.ReplyControl.Location, "IP属地：", "", -1),
	}
}

func FindUser(sem chan struct{}, opt *model.Option) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("爬取up：%d失败", opt.Mid))
			slog.Error(fmt.Sprint(err))
		}
	}()

	wg := sync.WaitGroup{}
	round := opt.Skip + 1
	var videoCollection = []model.VideoItem{}
	for round < opt.Pages+opt.Skip {
		// 停顿2s
		time.Sleep(2 * 1e9)
		slog.Info(fmt.Sprintf("爬取视频列表第%d页", round))
		tempVideoInfo, _ := FetchVideoList(opt.Mid, round, opt.Vorder, opt.Cookie)
		round++
		if tempVideoInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求up主视频列表失败，第%d页失败", round))
			slog.Error(tempVideoInfo.Message)
		}
		if len(tempVideoInfo.Data.List.Vlist) != 0 {
			videoCollection = append(videoCollection, tempVideoInfo.Data.List.Vlist...)

		} else {
			break
		}
	}

	if len(videoCollection) == 0 {
		slog.Info(fmt.Sprintf("up主：%d未获取到视频", opt.Mid))
		return
	}

	slog.Info(fmt.Sprintf("%d查找到%d条视频", opt.Mid, len(videoCollection)))
	for _, k := range videoCollection[:] {
		time.Sleep(3 * 1e9)
		slog.Info(fmt.Sprintf("------启动爬取%d------", k.Aid))
		wg.Add(1)
		sem <- struct{}{}
		go FindComment(sem, &wg, k.Aid, opt)
	}
	wg.Wait()
}
