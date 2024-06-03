package core

import (
	"blblcd/model"
	"blblcd/store"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

func FindComment(wg *sync.WaitGroup, avid int, opt *model.Option) {
	defer func() {
		wg.Done()
	}()
	oid := strconv.Itoa(avid)
	round := 0

	cmtInfo, _ := FetchComment(oid, 0, opt.Corder, opt.Cookie)
	total := cmtInfo.Data.Cursor.AllCount

	cmtCollection := make([]model.ReplyItem, total)
	collectionSlice := cmtCollection[:0]
	collectionSlice = append(collectionSlice, cmtInfo.Data.Replies...)
	for {
		round++
		// 停顿3s，模拟手工点击
		time.Sleep(3 * 1e9)
		slog.Info(fmt.Sprintf("爬取视频评论%s", oid))
		cmtInfo, _ := FetchComment(oid, round, opt.Corder, opt.Cookie)
		total := cmtInfo.Data.Cursor.AllCount
		if len(cmtInfo.Data.Replies) != 0 && len(collectionSlice) < total {
			collectionSlice = append(collectionSlice, cmtInfo.Data.Replies...)
			for _, k := range cmtInfo.Data.Replies {
				if len(k.Replies) > 0 {
					collectionSlice = append(collectionSlice, k.Replies...)
				}
			}
		} else {
			break
		}
	}
	var cmts = make([]model.Comment, total)
	sliceCmt := cmts[:0]
	for _, k := range cmtCollection[:] {
		cmt := NewCMT(&k)
		sliceCmt = append(sliceCmt, cmt)
	}
	ok := store.Save2CSV(oid, cmts[:], opt.Output)
	if ok {
		slog.Info("--爬取完成！--")
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
		Following:     item.Member.Following,
		Current_level: item.Member.LevelInfo.CurrentLevel,
		Location:      item.ReplyControl.Location,
		Time_desc:     item.ReplyControl.TimeDesc,
	}
}

func FindUserComments(opt *model.Option) {
	var wg sync.WaitGroup
	round := opt.Skip + 1
	pn := 30

	videoListInfo, err := FetchVideoList(opt.Mid, round, opt.Vorder, opt.Cookie)
	if err != nil {
		slog.Error(err.Error())
	}
	var videoCollection = make([]model.VideoItem, opt.Pages*pn)
	var videoListSlice = videoCollection[:0]
	videoListSlice = append(videoListSlice, videoListInfo.Data.List.Vlist...)
	for round < opt.Pages {
		round += 1
		// 停顿2s，避开机器扫描
		time.Sleep(2 * 1e9)
		slog.Info(fmt.Sprintf("爬取视频列表第%d页", round))
		tempVideoInfo, _ := FetchVideoList(opt.Mid, round, opt.Vorder, opt.Cookie)
		if len(tempVideoInfo.Data.List.Vlist) != 0 {
			videoListSlice = append(videoListSlice, tempVideoInfo.Data.List.Vlist...)
		} else {
			break
		}
	}

	slog.Info(fmt.Sprintf("%d查找到%d条视频", opt.Mid, len(videoListSlice)))
	for _, k := range videoCollection[:] {
		wg.Add(1)
		time.Sleep(3 * 1e9)
		slog.Info(fmt.Sprintf("------启动爬取%d------", k.Aid))
		go FindComment(&wg, k.Aid, opt)
	}
	wg.Wait()
}
