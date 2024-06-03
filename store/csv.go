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
		parseInt(cmt.Like), parseInt(cmt.Following), parseInt(cmt.Current_level),
		cmt.Location, cmt.Time_desc,
	}
}

func Save2CSV(filename string, cmts []model.Comment, ooutput string) (ok bool) {
	csv_path := fmt.Sprintf("%s/data_%s.csv", ooutput, filename)
	if utils.FileOrPathExists(csv_path) {
		// 打开已存在的 CSV 文件以追加数据
		file, err := os.OpenFile(csv_path, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
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
				fmt.Println("Error writing record to CSV:", err)
			}
		}

		fmt.Println("Multiple records appended to CSV file successfully.")
		ok = true

	} else {
		file, err := os.Create(csv_path)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// writer := csv.NewWriter(transform.NewWriter(file, simplifiedchinese.GBK.NewEncoder()))
		defer writer.Flush()
		headers := "upname,sex,content,bvid,rpid,oid,mid,parent,fans_grade,ctime,like,following,level,location,time_desc"
		headerErr := writer.Write(strings.Split(headers, ","))
		if headerErr != nil {
			slog.Error("Write csv header error:" + headerErr.Error())
			return
		}

		for _, cmt := range cmts {
			if cmt.Uname == "" {
				continue
			}
			record := CMT2Record(cmt)
			err := writer.Write(record)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}

		slog.Info(fmt.Sprintf("CSV file %s created successfully", csv_path))
		ok = true
	}
	return

}
