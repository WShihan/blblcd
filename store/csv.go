package store

import (
	"blblcd/model"
	"blblcd/utils"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func parseInt64(num int64) string {
	return fmt.Sprint(num)
}

func parseInt(num int) string {
	return strconv.Itoa(num)
}

func CMT2Record(cmt model.Comment) (record []string) {
	picURLs := ""
	for _, pic := range cmt.Pictures {
		picURLs += pic.Img_src + ";"
	}
	return []string{
		cmt.Bvid, cmt.Uname, cmt.Sex, cmt.Content, picURLs,
		parseInt64(cmt.Rpid), parseInt64(cmt.Oid), parseInt64(cmt.Mid),
		parseInt64(cmt.Parent), parseInt(cmt.Fansgrade), parseInt64(cmt.Ctime),
		parseInt64(cmt.Like), parseInt(cmt.Current_level), cmt.Location,
	}
}

func ReadExistCommentRpids(filename string) (map[int64]bool, error) {
	existRpids := make(map[int64]bool)
	if !utils.FileOrPathExists(filename) {
		return existRpids, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		slog.Error("打开csv文件错误", "err", err)
		return existRpids, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return existRpids, nil // empty file or read error
	}

	rpidIndex := -1
	for i, col := range header {
		if col == "rpid" {
			rpidIndex = i
			break
		}
	}

	if rpidIndex == -1 {
		slog.Error("在csv文件中未找到rpid字段", "header", header)
		return existRpids, nil
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("读取csv记录错误", "err", err)
			return existRpids, err
		}

		if len(record) > rpidIndex {
			rpid, err := strconv.ParseInt(record[rpidIndex], 10, 64)
			if err == nil {
				existRpids[rpid] = true
			}
		}
	}

	return existRpids, nil
}

// Save2CSV 把一页评论追加到 CSV，图片下载放在 CSV 锁之外，
// 避免图片下载期间阻塞其他页的写盘。
func Save2CSV(csvMutex *sync.Mutex, imageMutex *sync.Mutex, basename string, cmts []model.Comment, output string, downloadIMG bool) {
	if len(cmts) == 0 {
		return
	}

	records := make([]model.Comment, 0, len(cmts))
	for _, cmt := range cmts {
		if cmt.Uname != "" {
			records = append(records, cmt)
		}
	}
	if len(records) == 0 {
		return
	}

	csvPath := filepath.Join(output, basename+".csv")
	utils.PresetPath(output)

	csvMutex.Lock()
	appended := utils.FileOrPathExists(csvPath)
	err := appendRecords(csvPath, records)
	csvMutex.Unlock()
	if err != nil {
		slog.Error(fmt.Sprintf("写入csv文件错误，oid:%d", records[0].Oid), "err", err)
		return
	}
	if appended {
		slog.Info(fmt.Sprintf("追加评论至csv文件成功，oid:%d", records[0].Oid))
	} else {
		slog.Info(fmt.Sprintf("写入csv文件成功，oid:%d", records[0].Oid))
	}

	if !downloadIMG {
		return
	}

	var wg sync.WaitGroup
	for _, cmt := range records {
		if len(cmt.Pictures) == 0 {
			continue
		}
		wg.Add(1)
		go func(c model.Comment) {
			defer wg.Done()
			WriteImage(imageMutex, c.Uname, c.Pictures, filepath.Join(output, "images"))
		}(cmt)
	}
	wg.Wait()
}

// appendRecords 打开并追加一页记录；文件为空时补上表头。
// 不使用 fsync，依赖操作系统缓冲，避免每页一次同步写盘。
func appendRecords(csvPath string, records []model.Comment) error {
	file, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	info, statErr := file.Stat()
	writer := csv.NewWriter(file)
	if statErr == nil && info.Size() == 0 {
		headers := "bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location"
		if err := writer.Write(strings.Split(headers, ",")); err != nil {
			file.Close()
			return err
		}
	}

	for _, cmt := range records {
		if err := writer.Write(CMT2Record(cmt)); err != nil {
			file.Close()
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
