package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Inject     *Injection
	cookieFile string
	output     string
	mapping    bool
	workers    int
	corder     int
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
	rootCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 5, "最多协程数量")
	rootCmd.PersistentFlags().IntVarP(&workers, "corder", "v", 1, "爬取时评论排序方式，0：按时间，1：按点赞数，2：按回复数")
}

func Execute(injection *Injection) {
	Inject = injection
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
