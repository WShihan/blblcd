package cli

import (
	"blblcd/model"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Inject      *model.Injection
	cookieFile  string
	output      string
	mapping     bool
	workers     int
	corder      int
	imgDownload bool
	maxTryCount int
	maxDelaySec float64
)

var rootCmd = &cobra.Command{
	Use:   "blblcd",
	Short: "a command line tool for downloading bilibili comment",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("please type `blblcd -help` for more information")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cookieFile, "cookie", "c", "./cookie.text", "cookie文件路径")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "./output", "保存目录")
	rootCmd.PersistentFlags().BoolVarP(&mapping, "mapping", "m", false, "是否统计输出地图")
	rootCmd.PersistentFlags().BoolVarP(&imgDownload, "img-download", "i", false, "是否下载评论中的图片")
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 5, "最多协程数量")
	rootCmd.PersistentFlags().IntVarP(&maxTryCount, "max-try-count", "u", 3, "当爬取结果为空时请求最大尝试次数")
	rootCmd.PersistentFlags().Float64VarP(&maxDelaySec, "max-delay", "d", 1, "爬取最大延迟时间，单位秒")
	rootCmd.PersistentFlags().IntVarP(&corder, "corder", "v", 2, "爬取时评论排序方式，0：按热度，1：按热度+按时间，2：按时间，3：按热度")
}

func Execute(injection *model.Injection) {
	Inject = injection
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
