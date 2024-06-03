package store

import (
	"blblcd/model"
	"blblcd/utils"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func parseInt32(num int32) string {
	return fmt.Sprint(num)
}

func parseInt64(num int64) string {
	return fmt.Sprint(num)
}

func parseInt(num int) string {
	return strconv.Itoa(num)
}

func CMT2Record(cmt model.Comment) (record []string) {
	return []string{
		cmt.Uname, cmt.Sex, cmt.Content, cmt.Bvid,
		parseInt64(cmt.Rpid), parseInt(cmt.Oid), parseInt(cmt.Mid),
		parseInt(cmt.Parent), parseInt(cmt.Fansgrade), parseInt(cmt.Ctime),
		parseInt(cmt.Like), fmt.Sprint(cmt.Following), parseInt(cmt.Current_level), cmt.Location,
	}
}

func Save2CSV(filename string, cmts []model.Comment, ooutput string) (ok bool) {
	csv_path := fmt.Sprintf("%s/data_%s.csv", ooutput, filename)
	if utils.FileOrPathExists(csv_path) {
		file, err := os.OpenFile(csv_path, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.Error(fmt.Sprintf("打开csv文件错误，oid:%d", cmts[0].Oid))
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// writer := csv.NewWriter(transform.NewWriter(file, simplifiedchinese.GBK.NewEncoder()))

		defer writer.Flush()

		for _, cmt := range cmts {
			if cmt.Uname == "" {
				continue
			}
			record := CMT2Record(cmt)
			err = writer.Write(record)
			if err != nil {
				slog.Error(fmt.Sprintf("追加评论至csv文件错误，oid:%d", cmt.Oid))
			}
		}

		slog.Info(fmt.Sprintf("追加评论至csv文件成功，oid:%d", cmts[0].Oid))
		ok = true

	} else {
		file, err := os.Create(csv_path)
		if err != nil {
			slog.Error(fmt.Sprintf("创建csv文件错误，oid:%d", cmts[0].Oid))
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// writer := csv.NewWriter(transform.NewWriter(file, simplifiedchinese.GBK.NewEncoder()))
		defer writer.Flush()
		headers := "upname,sex,content,bvid,rpid,oid,mid,parent,fans_grade,ctime,like,following,level,location"
		headerErr := writer.Write(strings.Split(headers, ","))
		if headerErr != nil {
			slog.Error(fmt.Sprintf("写入csv文件字段错误，oid:%d", cmts[0].Oid))
			return
		}

		for _, cmt := range cmts {
			if cmt.Uname == "" {
				continue
			}
			record := CMT2Record(cmt)
			err := writer.Write(record)
			if err != nil {
				slog.Error(fmt.Sprintf("写入csv文件错误，oid:%d", cmt.Oid))
				return
			}
		}

		slog.Info(fmt.Sprintf("写入csv文件成功，oid:%d", cmts[0].Oid))
		ok = true
	}
	return

}
