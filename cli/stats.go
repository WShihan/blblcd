package cli

import (
	"blblcd/model"
	"blblcd/store"
	"blblcd/utils"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	statsFormat string
	statExt     map[string]string = map[string]string{
		"json": ".json",
		"csv":  ".csv",
		"html": ".html",
	}
)

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().StringVarP(&statsFormat, "format", "f", "json", "输出格式：json/csv/html")
	statsCmd.MarkFlagRequired("input")
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "统计本地CSV文件数据",
	Long:  `统计本地CSV文件数据，支持区域分布、性别分布、等级分布等统计`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("错误：必须指定输入文件或目录路径（使用 -i 参数）")
			return
		}
		for _, p := range args {
			// 检查输入路径是否存在
			if _, err := os.Stat(p); os.IsNotExist(err) {
				if !strings.HasSuffix(strings.ToLower(p), ".csv") {
					fmt.Printf("错误：输入路径不存在：%s\n", p)
					return
				}

			}

			// 创建输出目录
			if err := os.MkdirAll(output, 0755); err != nil {
				fmt.Printf("错误：创建输出目录失败：%v\n", err)
				return
			}

			utils.PrintLogo()
			fileNamtWithEx := strings.Split(p, "/")[len(strings.Split(p, "/"))-1]
			fileName := strings.Replace(fileNamtWithEx, ".csv", "", 1)
			fmt.Printf("开始统计：%s\n", fileName)

			statMap := make(map[string]model.Stat)
			totalComments := 0

			comments, err := processCSVFile(p)
			if err != nil {
				fmt.Printf("警告：处理文件 %s 失败：%v\n", p, err)
				continue
			}

			totalComments += len(comments)
			updateStatistics(statMap, comments)

			fmt.Printf("总计处理 %d 条评论\n", totalComments)

			// 输出统计结果
			switch strings.ToLower(statsFormat) {
			case "json":
				if err := writeJSONStats(statMap, fileName, output); err != nil {
					fmt.Printf("错误：输出JSON统计结果失败：%v\n", err)
					return
				}
			case "csv":
				if err := writeCSVStats(statMap, fileName, output); err != nil {
					fmt.Printf("错误：输出CSV统计结果失败：%v\n", err)
					return
				}
			case "html":
				if err := writeHTMLStats(statMap, fileName, output); err != nil {
					fmt.Printf("错误：输出HTML统计结果失败：%v\n", err)
					return
				}
			default:
				fmt.Printf("错误：不支持的输出格式：%s（支持：json/csv/html）\n", statsFormat)
				return
			}

			fmt.Printf("统计完成！结果已保存到：%s\n", filepath.Join(output, fileName+statExt[statsFormat]))
		}

	},
}

// processCSVFile 处理单个CSV文件
func processCSVFile(filePath string) ([]model.Comment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// 创建字段索引映射
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.TrimSpace(header)] = i
	}

	var comments []model.Comment

	// 读取数据行
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn(fmt.Sprintf("读取CSV行失败：%v", err))
			continue
		}

		comment := model.Comment{}

		// 解析字段
		if idx, ok := headerMap["bvid"]; ok && idx < len(record) {
			comment.Bvid = record[idx]
		}
		if idx, ok := headerMap["upname"]; ok && idx < len(record) {
			comment.Uname = record[idx]
		}
		if idx, ok := headerMap["sex"]; ok && idx < len(record) {
			comment.Sex = record[idx]
		}
		if idx, ok := headerMap["content"]; ok && idx < len(record) {
			comment.Content = record[idx]
		}
		if idx, ok := headerMap["pictures"]; ok && idx < len(record) {
			// CSV中pictures字段存储的是字符串，需要解析
			// 这里暂时不处理图片URL
		}
		if idx, ok := headerMap["rpid"]; ok && idx < len(record) {
			if rpid, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Rpid = rpid
			}
		}
		if idx, ok := headerMap["oid"]; ok && idx < len(record) {
			if oid, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Oid = oid
			}
		}
		if idx, ok := headerMap["mid"]; ok && idx < len(record) {
			if mid, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Mid = mid
			}
		}
		if idx, ok := headerMap["parent"]; ok && idx < len(record) {
			if parent, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Parent = parent
			}
		}
		if idx, ok := headerMap["fans_grade"]; ok && idx < len(record) {
			if fansGrade, err := strconv.Atoi(record[idx]); err == nil {
				comment.Fansgrade = fansGrade
			}
		}
		if idx, ok := headerMap["ctime"]; ok && idx < len(record) {
			if ctime, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Ctime = ctime
			}
		}
		if idx, ok := headerMap["like"]; ok && idx < len(record) {
			if like, err := strconv.ParseInt(record[idx], 10, 64); err == nil {
				comment.Like = like
			}
		}
		if idx, ok := headerMap["level"]; ok && idx < len(record) {
			if level, err := strconv.Atoi(record[idx]); err == nil {
				comment.Current_level = level
			}
		}
		if idx, ok := headerMap["location"]; ok && idx < len(record) {
			comment.Location = record[idx]
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

// updateStatistics 更新统计信息
func updateStatistics(statMap map[string]model.Stat, comments []model.Comment) {
	for _, comment := range comments {
		location := comment.Location
		if location == "" {
			location = "未知"
		}

		stat, exists := statMap[location]
		if !exists {
			stat = model.Stat{
				Name: location,
				Sex:  make(map[string]int),
			}
		}

		// 更新统计
		stat.Location++
		stat.Like += comment.Like

		// 性别统计
		sex := comment.Sex
		if sex == "" {
			sex = "保密"
		}
		stat.Sex[sex]++

		// 等级统计
		if comment.Current_level >= 1 && comment.Current_level <= 7 {
			stat.Level[comment.Current_level-1]++
		}

		statMap[location] = stat
	}
}

// writeJSONStats 输出JSON格式统计结果
func writeJSONStats(statMap map[string]model.Stat, fileName string, outputDir string) error {
	return writeTextStats(statMap, fileName, outputDir)
}

// writeCSVStats 输出CSV格式统计结果
func writeCSVStats(statMap map[string]model.Stat, filename string, outputDir string) error {
	outputPath := filepath.Join(outputDir, filename+"_stat.csv")
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"地区", "评论数", "总点赞数", "男", "女", "保密", "等级1", "等级2", "等级3", "等级4", "等级5", "等级6", "等级7"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据
	for _, stat := range statMap {
		record := []string{
			stat.Name,
			strconv.Itoa(stat.Location),
			strconv.FormatInt(stat.Like, 10),
			strconv.Itoa(stat.Sex["男"]),
			strconv.Itoa(stat.Sex["女"]),
			strconv.Itoa(stat.Sex["保密"]),
			strconv.Itoa(stat.Level[0]),
			strconv.Itoa(stat.Level[1]),
			strconv.Itoa(stat.Level[2]),
			strconv.Itoa(stat.Level[3]),
			strconv.Itoa(stat.Level[4]),
			strconv.Itoa(stat.Level[5]),
			strconv.Itoa(stat.Level[6]),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// writeHTMLStats 输出HTML格式统计结果（包含地图）
func writeHTMLStats(statMap map[string]model.Stat, fileName string, outputDir string) error {
	store.WriteGeoJSON(statMap, fileName, outputDir)
	return writeTextStats(statMap, fileName, outputDir)
}

// writeTextStats 输出文本格式统计结果
func writeTextStats(statMap map[string]model.Stat, fileName string, outputDir string) error {
	outputPath := filepath.Join(outputDir, fileName+"_stat.txt")
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 计算总计
	totalComments := 0
	totalLikes := int64(0)
	totalSex := map[string]int{"男": 0, "女": 0, "保密": 0}
	totalLevel := [7]int{}

	for _, stat := range statMap {
		totalComments += stat.Location
		totalLikes += stat.Like
		for sex, count := range stat.Sex {
			totalSex[sex] += count
		}
		for i := 0; i < 7; i++ {
			totalLevel[i] += stat.Level[i]
		}
	}

	// 写入统计摘要
	fmt.Fprintf(file, "=== 统计摘要 ===\n")
	fmt.Fprintf(file, "总计评论数：%d\n", totalComments)
	fmt.Fprintf(file, "总计点赞数：%d\n", totalLikes)
	fmt.Fprintf(file, "性别分布：男 %d，女 %d，保密 %d\n", totalSex["男"], totalSex["女"], totalSex["保密"])
	fmt.Fprintf(file, "等级分布：")
	for i, count := range totalLevel {
		fmt.Fprintf(file, "L%d:%d ", i+1, count)
	}
	fmt.Fprintf(file, "\n\n")

	// 写入地区详细统计
	fmt.Fprintf(file, "=== 地区详细统计 ===\n")
	fmt.Fprintf(file, "%-10s %-10s %-12s %-8s %-8s %-8s %s\n",
		"地区", "评论数", "点赞数", "男", "女", "保密", "等级分布")

	for _, stat := range statMap {
		levelStr := ""
		for i, count := range stat.Level {
			if count > 0 {
				levelStr += fmt.Sprintf("L%d:%d ", i+1, count)
			}
		}
		if levelStr == "" {
			levelStr = "无"
		}

		fmt.Fprintf(file, "%-10s %-10d %-12d %-8d %-8d %-8d %s\n",
			stat.Name,
			stat.Location,
			stat.Like,
			stat.Sex["男"],
			stat.Sex["女"],
			stat.Sex["保密"],
			levelStr)
	}

	return nil
}
