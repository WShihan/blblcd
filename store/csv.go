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

func Save2CSV(csvMutex *sync.Mutex, imageMutex *sync.Mutex, basename string, cmts []model.Comment, output string, downloadIMG bool) {
	csvMutex.Lock()
	defer csvMutex.Unlock()

	utils.PresetPath(output)
	if len(cmts) == 0 {
		return
	}
	var wg sync.WaitGroup
	csvPath := filepath.Join(output, basename+".csv")
	if utils.FileOrPathExists(csvPath) {
		file, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Error(fmt.Sprintf("打开csv文件错误，oid:%d", cmts[0].Oid))
			return
		}
		defer file.Close()
		defer file.Sync()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		for _, cmt := range cmts {
			if cmt.Uname == "" {
				continue
			}
			if downloadIMG {
				if len(cmt.Pictures) != 0 {
					wg.Add(1)
					go func(c model.Comment) {
						defer wg.Done()
						WriteImage(imageMutex, c.Uname, c.Pictures, output+"/"+"images")
					}(cmt)
				}
			}

			record := CMT2Record(cmt)
			err = writer.Write(record)
			if err != nil {
				slog.Error(fmt.Sprintf("追加评论至csv文件错误，oid:%d", cmt.Oid))
			}
		}

		slog.Info(fmt.Sprintf("追加评论至csv文件成功，oid:%d", cmts[0].Oid))

	} else {
		file, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			slog.Error(fmt.Sprintf("创建csv文件错误，oid:%d", cmts[0].Oid))
			return
		}
		defer file.Close()
		defer file.Sync()

		writer := csv.NewWriter(file)
		defer writer.Flush()
		headers := "bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location"
		headerErr := writer.Write(strings.Split(headers, ","))
		if headerErr != nil {
			slog.Error(fmt.Sprintf("写入csv文件字段错误，oid:%d", cmts[0].Oid))
			return
		}

		for _, cmt := range cmts {
			if cmt.Uname == "" {
				continue
			}
			if downloadIMG {
				if len(cmt.Pictures) != 0 {
					wg.Add(1)
					go func(c model.Comment) {
						defer wg.Done()
						WriteImage(imageMutex, c.Uname, c.Pictures, output+"/"+"images")
					}(cmt)
				}
			}

			record := CMT2Record(cmt)
			err := writer.Write(record)
			if err != nil {
				slog.Error(fmt.Sprintf("写入csv文件错误，oid:%d", cmt.Oid))
				return
			}
		}
		slog.Info(fmt.Sprintf("写入csv文件成功，oid:%d", cmts[0].Oid))
	}

	wg.Wait()
}
