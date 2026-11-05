package core

import (
	"blblcd/client"
	"blblcd/model"
	"blblcd/store"
	"blblcd/utils"
	"errors"
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

const (
	defaultFetchConcurrency = 5
	subCommentPageSize      = 20
	// 防御异常数据：单条根评论最多翻这么多页楼中楼。
	maxSubCommentPages = 500
	// 网络错误 / 风控类错误的重试次数（不含首次请求）。
	maxRequestRetries = 2
	// 请求返回重复数据时的默认停止阈值。
	defaultMaxTryCount = 3
	// B站从 2022-07-25 11:00 (UTC+8) 开始为评论提供 IP 属地。
	ipLocationStart = int64(1658718000)
)

func concurrencyOf(opt *model.Option) int {
	if opt == nil || opt.Concurrency <= 0 {
		return defaultFetchConcurrency
	}
	return opt.Concurrency
}

func maxTryCountOf(opt *model.Option) int {
	if opt == nil || opt.MaxTryCount <= 0 {
		return defaultMaxTryCount
	}
	return opt.MaxTryCount
}

// retryableCode 判断接口错误码是否值得重试：多为风控、限流或服务端临时错误。
func retryableCode(code int) bool {
	switch code {
	case -352, -403, -412, -500, -503, -504, -509, -799:
		return true
	default:
		return false
	}
}

func retryDelay(attempt int) {
	base := time.Duration(attempt) * 400 * time.Millisecond
	jitter := time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
	time.Sleep(base + jitter)
}

// fetchCountWithRetry 拉取评论总数；失败重试几次，避免一次临时 412 导致整个视频放弃。
func fetchCountWithRetry(oid string, opt *model.Option) (int64, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRequestRetries; attempt++ {
		if attempt > 0 {
			retryDelay(attempt)
		}
		utils.RandomDelay(opt.MaxDelaySec)
		count, err := FetchCount(oid)
		if err == nil {
			return count, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

// fetchCommentPage 带随机延迟与重试地拉取一页主评论。
func fetchCommentPage(oid string, opt *model.Option, offsetStr string) (resp model.CommentResponse, err error) {
	var lastErr error
	riskControl := false
	for attempt := 0; attempt <= maxRequestRetries; attempt++ {
		if attempt > 0 {
			if riskControl {
				// 主评论必须继续翻页，412 时多等一会儿再试。
				time.Sleep(time.Duration(attempt) * 1500 * time.Millisecond)
			} else {
				retryDelay(attempt)
			}
		}
		utils.RandomDelay(opt.MaxDelaySec)
		resp, err = FetchComment(oid, opt.Corder, opt.Cookie, offsetStr)
		if err == nil && resp.Code == 0 {
			return resp, nil
		}
		if err == nil && !retryableCode(resp.Code) {
			return resp, nil
		}
		riskControl = errors.Is(err, client.ErrRiskControl)
		lastErr = err
		if lastErr == nil {
			lastErr = fmt.Errorf("接口错误码 %d: %s", resp.Code, resp.Message)
		}
		if attempt < maxRequestRetries {
			slog.Warn("请求评论失败，准备重试", "oid", oid, "attempt", attempt+1, "err", lastErr)
		}
	}
	return resp, lastErr
}

// fetchSubCommentPage 带随机延迟与重试地拉取一页楼中楼。
// 已登录时 412 重试一次；cookie 未登录时 412 多半是该楼中楼不可见，直接交给预览回复兜底。
func fetchSubCommentPage(oid string, rpid int64, pn int, opt *model.Option) (resp model.CommentResponse, err error) {
	var lastErr error
	riskControl := false
	for attempt := 0; attempt <= maxRequestRetries; attempt++ {
		if attempt > 0 {
			if riskControl {
				time.Sleep(1200 * time.Millisecond)
			} else {
				retryDelay(attempt)
			}
		}
		utils.RandomDelay(opt.MaxDelaySec)
		resp, err = FetchSubComment(oid, rpid, pn, opt.Cookie)
		if err == nil && resp.Code == 0 {
			return resp, nil
		}
		riskControl = errors.Is(err, client.ErrRiskControl)
		if riskControl && (!client.IsLoggedIn() || attempt >= 1) {
			return resp, err
		}
		if err == nil && !retryableCode(resp.Code) {
			return resp, nil
		}
		lastErr = err
		if lastErr == nil {
			lastErr = fmt.Errorf("接口错误码 %d: %s", resp.Code, resp.Message)
		}
		if attempt < maxRequestRetries {
			slog.Warn("请求子评论失败，准备重试", "oid", oid, "rpid", rpid, "pn", pn, "attempt", attempt+1, "err", lastErr)
		}
	}
	return resp, lastErr
}

func appendReplyItems(dst []model.ReplyItem, replies []model.ReplyItem) []model.ReplyItem {
	for _, k := range replies {
		dst = append(dst, k)
		dst = append(dst, k.Replies...)
	}
	return dst
}

func FindComment(avid int64, opt *model.Option) {
	if opt == nil {
		opt = &model.Option{}
	}

	var filename string
	if opt.Bvid != "" {
		filename = opt.Bvid
	} else {
		filename = Avid2Bvid(int64(avid))
	}

	oid := strconv.FormatInt(avid, 10)

	type pageResult struct {
		resp model.CommentResponse
		err  error
	}
	fetchPage := func(offset string) pageResult {
		resp, err := fetchCommentPage(oid, opt, offset)
		return pageResult{resp: resp, err: err}
	}

	// 主评论只能按游标串行翻页，但可以提前请求：
	// 第一页与评论总数并发获取，后续页在当前页处理楼中楼时预先请求，
	// 从而把网络往返和随机延迟从关键路径上隐藏掉。
	pageCh := make(chan pageResult, 1)
	go func() {
		pageCh <- fetchPage("")
	}()

	totalCount, err := fetchCountWithRetry(oid, opt)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info(fmt.Sprintf(">>>视频%s总共有%d条评论<<<\n", filename, totalCount))

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
	missingLocation := 0
	missingRootLocation := 0
	missingSubLocation := 0
	missingOldLocation := 0
	defer func() {
		endTime := time.Now().Unix()
		absPath, err := filepath.Abs(path.Join(opt.Output, opt.Bvid))
		if err != nil {
			log.Fatal(fmt.Sprintf("获取绝对路径异常：%s", err.Error()))
		}
		if totalCount == 0 || downloadedCount != totalCount {
			slog.Info(fmt.Sprintf("视频：%d 预期有 %d 条评论, 但实际获取了 %d 条", avid, totalCount, downloadedCount))
		}
		if missingLocation > 0 {
			detail := fmt.Sprintf("根评论 %d，楼中楼 %d，2022-07-25 前发布 %d", missingRootLocation, missingSubLocation, missingOldLocation)
			if !client.IsLoggedIn() {
				slog.Warn(fmt.Sprintf("有 %d 条评论未获取到 IP 属地（%s）：当前 cookie 未登录，B站不会返回 location，请更新 cookie 后重试", missingLocation, detail))
			} else {
				slog.Info(fmt.Sprintf("有 %d 条评论未获取到 IP 属地（%s），旧评论本身没有属地", missingLocation, detail))
				if missingRootLocation > 0 {
					slog.Warn("已登录但根评论缺少 IP 属地：请确认使用的是最新构建（旧版主评论请求不带 Cookie），并确认 cookie 中包含 SESSDATA", "root_missing", missingRootLocation)
				}
			}
		}
		slog.Info(fmt.Sprintf("***** 爬取视频：%s评论完成，用时%d秒，保存至 %s***** ", opt.Bvid, endTime-startTime, absPath))
	}()

	for {
		round++

		result := <-pageCh
		if result.err != nil {
			slog.Error(fmt.Sprintf("请求评论失败，视频%s，第%d页失败", filename, round), "err", result.err)
			break
		}
		cmtInfo := result.resp
		if cmtInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求评论失败，视频%s，第%d页失败", filename, round), "code", cmtInfo.Code)
			slog.Error(cmtInfo.Message)
			break
		}

		replies := cmtInfo.Data.Replies
		nextOffset := ""
		if len(replies) != 0 {
			nextOffset = cmtInfo.Data.Cursor.PaginationReply.NextOffset
		}
		if round == 1 && !client.IsLoggedIn() {
			slog.Warn("当前 cookie 未登录，楼中楼可能只能获取第一页，且评论不会返回 IP 属地；如需完整数据请更新 cookie")
		}

		// 先发起下一页请求，再去处理本页楼中楼。
		if nextOffset != "" && downloadedCount < totalCount {
			go func(offset string) {
				pageCh <- fetchPage(offset)
			}(nextOffset)
		}

		replyCollection := []model.ReplyItem{}
		if len(replies) != 0 {
			replyCollection = append(replyCollection, replies...)
			replyCollection = append(replyCollection, collectSubComments(replies, opt)...)
			if len(cmtInfo.Data.TopReplies) != 0 {
				for _, k := range cmtInfo.Data.TopReplies {
					replyCollection = append(replyCollection, k)
					replyCollection = append(replyCollection, k.Replies...)
				}
			}
		}

		// 如果接口返回空数据，直接停止
		if len(replyCollection) == 0 {
			slog.Info(fmt.Sprintf("视频：%s，第%d页未获取到评论，停止爬取，pagination_str: %s", oid, round, nextOffset))
			break
		}

		cmtCollection := []model.Comment{}
		isEmpty := true
		for _, k := range replyCollection {
			if _, ok := recordedMap[k.Rpid]; ok {
				slog.Debug("评论已存在，跳过", "rpid", k.Rpid)
				continue
			}
			isEmpty = false
			cmt := NewCMT(&k)
			if cmt.Location == "" {
				missingLocation++
				if k.Parent == 0 {
					missingRootLocation++
				} else {
					missingSubLocation++
				}
				if cmt.Ctime < ipLocationStart {
					missingOldLocation++
				}
			}
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
		}

		// 返回非空数据，但是重复，超过限制后直接停止
		if isEmpty {
			maxTryCount++
			slog.Info(fmt.Sprintf("请求返回空数据第%d次, round: %d, pagination_str: %s", maxTryCount, round, nextOffset))
			if maxTryCount >= maxTryCountOf(opt) {
				slog.Info(fmt.Sprintf("请求返回空数据达到%d次，停止爬取", maxTryCountOf(opt)))
				break
			}
		} else {
			maxTryCount = 0
		}

		downloadedCount += int64(len(cmtCollection))
		utils.ProgressBar(downloadedCount, totalCount)

		csvWaitGroup.Add(1)
		go func(cmts []model.Comment) {
			defer csvWaitGroup.Done()
			store.Save2CSV(&csvMutex, &imageMutex, filename, cmts, savePath, opt.ImgDownload)
		}(cmtCollection)

		if downloadedCount >= totalCount {
			break
		}
		if nextOffset == "" {
			break
		}
	}
	csvWaitGroup.Wait()

	if opt.Mapping {
		store.WriteGeoJSON(statMap, filename, savePath)
	}
}

// collectSubComments 并发拉取本页所有根评论的楼中楼，并按根评论顺序返回。
// 已经随主评论返回了完整预览的楼中楼不会重复请求。
func collectSubComments(roots []model.ReplyItem, opt *model.Option) []model.ReplyItem {
	fetched := make([][]model.ReplyItem, len(roots))
	jobs := make([]int, 0, len(roots))
	for i := range roots {
		k := roots[i]
		if k.Rcount == 0 {
			continue
		}
		if len(k.Replies) > 0 && int64(len(k.Replies)) == k.Rcount {
			continue
		}
		jobs = append(jobs, i)
	}

	if len(jobs) > 0 {
		sem := make(chan struct{}, concurrencyOf(opt))
		var wg sync.WaitGroup
		for _, idx := range jobs {
			wg.Add(1)
			sem <- struct{}{}
			go func(i int) {
				defer wg.Done()
				defer func() { <-sem }()
				fetched[i] = FindSubComment(roots[i], opt)
			}(idx)
		}
		wg.Wait()
	}

	collection := []model.ReplyItem{}
	fallbacks := 0
	for i := range roots {
		k := roots[i]
		if k.Rcount == 0 {
			continue
		}
		if len(k.Replies) > 0 && int64(len(k.Replies)) == k.Rcount {
			collection = append(collection, k.Replies...)
			continue
		}

		items := fetched[i]
		// 拉取失败（例如触发风控）时至少保留主评论里附带的预览回复，避免丢数据。
		if len(k.Replies) > 0 {
			if len(items) == 0 {
				fallbacks++
			}
			seen := make(map[int64]struct{}, len(items))
			for _, r := range items {
				seen[r.Rpid] = struct{}{}
			}
			for _, r := range k.Replies {
				if _, ok := seen[r.Rpid]; !ok {
					items = append(items, r)
				}
			}
		}
		collection = append(collection, items...)
	}
	if fallbacks > 0 {
		slog.Warn("部分楼中楼拉取失败，已用主评论预览回复兜底；如大量出现请检查 cookie 或降低并发", "count", fallbacks)
	}
	return collection
}

// FindSubComment 拉取一条根评论下的全部楼中楼。
// 首页之后的分页按 Rcount 计算页数并并发拉取，同时保留翻页顺序。
func FindSubComment(cmt model.ReplyItem, opt *model.Option) []model.ReplyItem {
	if opt == nil {
		opt = &model.Option{}
	}
	oid := strconv.FormatInt(cmt.Oid, 10)
	replyCollection := []model.ReplyItem{}

	first, err := fetchSubCommentPage(oid, cmt.Rpid, 1, opt)
	if err != nil {
		if errors.Is(err, client.ErrRiskControl) {
			slog.Debug("请求子评论触发412，使用预览回复兜底", "rpid", cmt.Rpid)
		} else {
			slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第1页失败", cmt.Rpid), "err", err)
		}
		return replyCollection
	}
	if first.Code != 0 {
		slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第1页失败，code=%d", cmt.Rpid, first.Code))
		slog.Error(first.Message)
		return replyCollection
	}
	firstPageCount := len(first.Data.Replies)
	replyCollection = appendReplyItems(replyCollection, first.Data.Replies)
	replyCollection = appendReplyItems(replyCollection, first.Data.TopReplies)

	if firstPageCount == 0 {
		slog.Debug(fmt.Sprintf("视频%s，评论%d，第1页未获取到子评论", oid, cmt.Rpid))
		return replyCollection
	}

	knownTotal := cmt.Rcount
	if pageCount := first.Data.Page.Count; pageCount > knownTotal {
		knownTotal = pageCount
	}
	unknownTotal := knownTotal <= 0
	if knownTotal <= int64(firstPageCount) {
		if firstPageCount < subCommentPageSize {
			return replyCollection
		}
		// 总数未知但首页已满，继续探测下一页。
		unknownTotal = true
		knownTotal = int64(firstPageCount) + 1
	}

	pages := int((knownTotal + subCommentPageSize - 1) / subCommentPageSize)
	if pages < 2 {
		pages = 2
	}
	if unknownTotal {
		// 总数未知时一直探测到空页为止（由 maxSubCommentPages 兜底）。
		pages = maxSubCommentPages
	}
	if pages > maxSubCommentPages {
		slog.Warn("楼中楼页数超过上限，已截断", "rpid", cmt.Rpid, "pages", pages, "limit", maxSubCommentPages)
		pages = maxSubCommentPages
	}

	// 先顺序探测第 2 页：未登录 / cookie 失效时接口只返回第一页，
	// 这样不会为每一条大热评发出几十个必然为空的请求。
	second, err := fetchSubCommentPage(oid, cmt.Rpid, 2, opt)
	if err != nil || second.Code != 0 || len(second.Data.Replies) == 0 {
		if err != nil && !errors.Is(err, client.ErrRiskControl) {
			slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第2页失败", cmt.Rpid), "err", err)
		}
		slog.Debug(fmt.Sprintf("视频%s，评论%d，第2页未获取到子评论，停止翻页", oid, cmt.Rpid))
		return replyCollection
	}
	replyCollection = appendReplyItems(replyCollection, second.Data.Replies)
	replyCollection = appendReplyItems(replyCollection, second.Data.TopReplies)

	// 剩余分页并发拉取；每批内保留页码顺序，遇到空页立即停止后续批次。
	batchSize := concurrencyOf(opt)
	for start := 3; start <= pages; start += batchSize {
		end := start + batchSize - 1
		if end > pages {
			end = pages
		}
		results := make([]model.CommentResponse, end-start+1)
		errs := make([]error, end-start+1)
		var wg sync.WaitGroup
		for pn := start; pn <= end; pn++ {
			wg.Add(1)
			go func(pn int) {
				defer wg.Done()
				resp, err := fetchSubCommentPage(oid, cmt.Rpid, pn, opt)
				results[pn-start] = resp
				errs[pn-start] = err
			}(pn)
		}
		wg.Wait()

		stop := false
		for i := range results {
			if errs[i] != nil {
				if errors.Is(errs[i], client.ErrRiskControl) {
					slog.Debug("楼中楼分页触发412，停止翻页", "rpid", cmt.Rpid, "pn", start+i)
				} else {
					slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第%d页失败", cmt.Rpid, start+i), "err", errs[i])
				}
				stop = true
				break
			}
			if results[i].Code != 0 {
				slog.Error(fmt.Sprintf("请求子评论失败，父评论%d，第%d页失败，code=%d", cmt.Rpid, start+i, results[i].Code))
				stop = true
				break
			}
			if len(results[i].Data.Replies) == 0 {
				stop = true
				break
			}
			replyCollection = appendReplyItems(replyCollection, results[i].Data.Replies)
			replyCollection = appendReplyItems(replyCollection, results[i].Data.TopReplies)
		}
		if stop {
			break
		}
	}
	return replyCollection
}

// 从评论回复列表提取感兴趣信息
func NewCMT(item *model.ReplyItem) model.Comment {
	location := strings.ReplaceAll(item.ReplyControl.Location, "IP属地：", "")
	location = strings.ReplaceAll(location, "IP属地:", "")
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
		Location:      strings.TrimSpace(location),
	}
}

func FindUser(sem chan struct{}, opt *model.Option) {
	if opt == nil {
		opt = &model.Option{}
	}
	round := opt.Skip + 1
	var videoCollection = []model.VideoItem{}
	for round <= opt.Pages+opt.Skip {
		slog.Info(fmt.Sprintf("爬取视频列表第%d页", round))
		tempVideoInfo, err := FetchVideoList(opt.Mid, round, opt.Vorder, opt.Cookie)
		if err != nil {
			slog.Error(fmt.Sprintf("请求up主视频列表失败，第%d页失败", round), "err", err)
			break
		}
		round++
		if tempVideoInfo.Code != 0 {
			slog.Error(fmt.Sprintf("请求up主视频列表失败，第%d页失败", round))
			slog.Error(tempVideoInfo.Message)
		}
		utils.RandomDelay(opt.MaxDelaySec)

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
	for _, k := range videoCollection {
		slog.Info(fmt.Sprintf("------启动爬取%d------", k.Aid))
		sem <- struct{}{}
		wg.Add(1)
		go func(avid int64) {
			defer wg.Done()
			defer func() {
				<-sem
			}()
			FindComment(avid, opt)
		}(k.Aid)
	}
	wg.Wait()
}
