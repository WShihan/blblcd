package core

import (
	"blblcd/model"
	"blblcd/store"
	"blblcd/utils"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func FindComment(avid int64, opt *model.Option) {
	var filename string
	if opt.Bvid != "" {
		filename = opt.Bvid
	} else {
		filename = Avid2Bvid(int64(avid))
	}

	oid := strconv.FormatInt(avid, 10)
	totalCount, err := FetchCount(oid)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprintf(">>>视频%s总共有%d条评论<<<\n", filename, totalCount))
	time.Sleep(time.Duration((rand.Float32() + 1) * 1e8))

	savePath := path.Join(opt.Output, opt.Bvid)
	utils.PresetPath(savePath)

	csvPath := path.Join(savePath, filename+".csv")
	existMap, err := store.ReadExistCommentRpids(csvPath)
	if err != nil {
		slog.Error("读取已存在评论错误", "err", err)
	}

	downloadedCount := int64(0)
	if len(existMap) > 0 {
		downloadedCount = int64(len(existMap))
		slog.Info(fmt.Sprintf("从 %s 加载了 %d 条已存在的评论", filename+".csv", downloadedCount))
	}

	startTime := time.Now().Unix()
	var csvMutex sync.Mutex
	var imageMutex sync.Mutex
	var csvWaitGroup sync.WaitGroup
	recordedMap := map[int64]bool{}
	statMap := map[string]model.Stat{}
	round := 0
	maxTryCount := 0 // 当接口返回空数据后重试次数
	offsetStr := ""
	defer func() {
		endTime := time.Now().Unix()
		absPath, err := filepath.Abs(path.Join(opt.Output, opt.Bvid))
		if err != nil {
			log.Fatal(fmt.Sprintf("获取绝对路径异常：%s", err.Error()))
		}
		if totalCount == 0 || downloadedCount != totalCount {
			slog.Info(fmt.Sprintf("视频：%d 预期有 %d 条评论, 但实际获取了 %d 条", avid, totalCount, downloadedCount))
		}
		slog.Info(fmt.Sprintf("***** 爬取视频：%s评论完成，用时%d秒，保存至 %s***** ", opt.Bvid, endTime-startTime, absPath))
	}()

	for {
		round++

		replyCollection := []model.ReplyItem{}

		cmtInfo, _ := FetchComment(oid, opt.Corder, opt.Cookie, offsetStr)
		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求评论失败，视频%s，第%d页失败", filename, round))
			slog.Error(cmtInfo.Message)
			break
		}
		time.Sleep(time.Duration((rand.Float32() + 1) * 1e9))

		if len(cmtInfo.Data.Replies) != 0 {
			offsetStr = cmtInfo.Data.Cursor.PaginationReply.NextOffset
			replyCollection = append(replyCollection, cmtInfo.Data.Replies...)
			for _, k := range cmtInfo.Data.Replies {
				if k.Rcount == 0 {
					continue
				}
				if len(k.Replies) > 0 && int64(len(k.Replies)) == k.Rcount {
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
			slog.Info(fmt.Sprintf("视频：%s，第%d页未获取到评论，停止爬取，pagination_str: %s", oid, round, offsetStr))
			break
		}

		cmtCollection := []model.Comment{}
		isEmpty := true
		for _, k := range replyCollection[:] {
			if _, ok := recordedMap[k.Rpid]; !ok {
				isEmpty = false
				cmt := NewCMT(&k)
				recordedMap[cmt.Rpid] = true

				if _, ok := existMap[k.Rpid]; !ok {
					cmtCollection = append(cmtCollection, cmt)
				}

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
			slog.Info(fmt.Sprintf("请求返回空数据第%d次, round: %d, pagination_str: %s", maxTryCount, round, offsetStr))
			if maxTryCount >= opt.MaxTryCount {
				slog.Info(fmt.Sprintf("请求返回空数据达到%d次，停止爬取", opt.MaxTryCount))
				break
			}
		}

		downloadedCount += int64(len(cmtCollection))
		utils.ProgressBar(downloadedCount, totalCount)

		csvWaitGroup.Add(1)
		go func() {
			defer csvWaitGroup.Done()
			store.Save2CSV(&csvMutex, &imageMutex, filename, cmtCollection, savePath, opt.ImgDownload)
		}()

		if downloadedCount >= totalCount {
			break
		}
	}
	csvWaitGroup.Wait()

	if opt.Mapping {
		store.WriteGeoJSON(statMap, filename, savePath)
	}
}

func FindSubComment(cmt model.ReplyItem, opt *model.Option) []model.ReplyItem {
	oid := strconv.FormatInt(cmt.Oid, 10)
	round := 0
	replyCollection := []model.ReplyItem{}
	for {
		round++

		cmtInfo, _ := FetchSubComment(oid, cmt.Rpid, round, opt.Cookie)
		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第%d页失败", cmt.Rpid, round))
			slog.Error(cmtInfo.Message)
			break
		}
		time.Sleep(time.Duration((rand.Float32() + 1) * 1.5e9))

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
		Location:      strings.ReplaceAll(item.ReplyControl.Location, "IP属地：", ""),
	}
}

func FindUser(sem chan struct{}, opt *model.Option) {
	round := opt.Skip + 1
	var videoCollection = []model.VideoItem{}
	for round <= opt.Pages+opt.Skip {
		slog.Info(fmt.Sprintf("爬取视频列表第%d页", round))
		tempVideoInfo, _ := FetchVideoList(opt.Mid, round, opt.Vorder, opt.Cookie)
		round++
		if tempVideoInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求up主视频列表失败，第%d页失败", round))
			slog.Error(tempVideoInfo.Message)
		}
		time.Sleep(time.Duration((rand.Float32() + 1) * 3e9))

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
	wg := sync.WaitGroup{}
	for _, k := range videoCollection[:] {
		slog.Info(fmt.Sprintf("------启动爬取%d------", k.Aid))
		sem <- struct{}{}
		wg.Add(1)
		go func(avid int64) {
			defer wg.Done()
			defer func() {
				<-sem
			}()
			FindComment(k.Aid, opt)
		}(k.Aid)
		time.Sleep(time.Duration((rand.Float32() + 1) * 3e9))
	}
	wg.Wait()
}
