package core

import (
	"blblcd/model"
	"blblcd/store"
	"blblcd/utils"
	"fmt"
	"log/slog"
	"math/rand"
	"path"
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
	var filename string
	if opt.Bvid != "" {
		filename = opt.Bvid
	} else {
		filename = Avid2Bvid(int64(avid))
	}
	oid := strconv.Itoa(avid)
	total, err := FetchCount(oid)
	slog.Info(fmt.Sprintf(">>>视频%s总共有%d条评论<<<\n", oid, total))
	downloadedCount := 0
	if err != nil {
		slog.Error(err.Error())
		return
	}
	savePath := path.Join(opt.Output, opt.Bvid)
	round := 0
	recordedMap := make(map[int64]bool)
	statMap := map[string]model.Stat{}
	maxTryCount := 0 // 当接口返回空数据后重试次数
	offsetStr := ""
	for {
		replyCollection := []model.ReplyItem{}
		// 停顿
		// delay := (rand.Float32() + 1) * 1e9
		// time.Sleep(time.Duration(delay))
		slog.Info(fmt.Sprintf("爬取视频评论%s", oid))
		round++
		cmtInfo, _ := FetchComment(oid, round, opt.Corder, opt.Cookie, offsetStr)

		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求评论失败，视频%s，第%d页失败", oid, round))
			slog.Error(cmtInfo.Message)
			break
		}
		if len(cmtInfo.Data.Replies) != 0 && len(replyCollection) < total {
			offsetStr = cmtInfo.Data.Cursor.PaginationReply.NextOffset
			replyCollection = append(replyCollection, cmtInfo.Data.Replies...)
			for _, k := range cmtInfo.Data.Replies {
				if k.Rcount == 0 {
					continue
				}
				if len(k.Replies) > 0 && len(k.Replies) == k.Rcount {
					replyCollection = append(replyCollection, k.Replies...)
				} else {
					subCmts := FindSubComment(k, opt)
					replyCollection = append(replyCollection, subCmts...)
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

		// 如果接口返回空数据，直接停止
		if len(replyCollection) == 0 {
			slog.Info(fmt.Sprintf("-----视频%s，第%d页未获取到评论，停止爬取-----", oid, round))
			break
		}

		var cmtCollection = []model.Comment{}
		isEmpty := true
		for _, k := range replyCollection[:] {
			if _, ok := recordedMap[k.Rpid]; !ok {
				isEmpty = false
				cmt := NewCMT(&k)
				recordedMap[cmt.Rpid] = true
				cmtCollection = append(cmtCollection, cmt)
				if opt.Mapping {
					if value, exist := statMap[cmt.Location]; exist {
						value.Location += 1
						value.Sex[cmt.Sex] += 1
						value.Like += cmt.Like
						value.Level[cmt.Current_level] += 1
						statMap[cmt.Location] = value
					} else {
						state := model.Stat{
							Name:     cmt.Location,
							Location: 1,
							Sex:      map[string]int{"男": 0, "女": 0, "保密": 0},
							Like:     cmt.Like,
						}
						state.Sex[cmt.Sex] += 1
						state.Level[cmt.Current_level] += 1
						statMap[cmt.Location] = state

					}
				}
			} else {
				slog.Info(fmt.Sprintf("评论%d已存在，跳过", k.Rpid))
			}

		}

		// 返回非空数据，但是重复，超过限制后直接停止
		if isEmpty {
			maxTryCount++
			if maxTryCount >= opt.MaxTryCount {
				slog.Info(fmt.Sprintf("请求返回空数据%d次，超过限度将停止爬取", opt.MaxTryCount))
				break
			}
		}

		downloadedCount += len(cmtCollection)
		utils.ProgressBar(downloadedCount, total)

		go store.Save2CSV(filename, cmtCollection, savePath, opt.ImgDownload)

		if downloadedCount >= total {
			slog.Info(fmt.Sprintf("*****爬取视频：%s评论完成*****", opt.Bvid))
			break
		}
	}
	if opt.Mapping {
		store.WriteGeoJSON(statMap, filename, savePath)

	}
}

func FindSubComment(cmt model.ReplyItem, opt *model.Option) []model.ReplyItem {
	oid := strconv.Itoa(cmt.Oid)
	round := 1
	replyCollection := []model.ReplyItem{}
	for {
		// 停顿
		delay := (rand.Float32() + 1) * 1e9
		time.Sleep(time.Duration(delay))
		cmtInfo, _ := FetchSubComment(oid, cmt.Rpid, round, opt.Cookie)
		round++
		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第%d页失败", cmt.Rpid, round))
			slog.Error(cmtInfo.Message)
			break
		}
		if len(cmtInfo.Data.Replies) > 0 {
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
		} else {
			slog.Info(fmt.Sprintf("******视频%s，评论%d，第%d页未获取到子评论，停止爬取子评论******", oid, cmt.Rpid, round))
			break
		}
	}
	return replyCollection

}

// 从评论回复列表提取感兴趣信息
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
		Pictures:      item.Content.Pictures,
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
	for round <= opt.Pages+opt.Skip {
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
