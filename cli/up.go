package cli

import (
	"blblcd/core"
	"blblcd/model"
	"blblcd/utils"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	pages  int
	skip   int
	vorder string
)

func init() {
	upCmd.Flags().IntVarP(&pages, "pages", "p", 3, "获取的页数")
	upCmd.Flags().IntVarP(&skip, "skip", "s", 0, "跳过视频的页数")
	upCmd.Flags().StringVarP(&vorder, "corder", "t", "pubdate", "爬取up主视频列表时排序方式，最新发布：pubdate最多播放：click最多收藏：stow")

	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "批量获取UP主视频列表的评论",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("please provide mid")
			return
		}
		mid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		utils.PresetPath(output)
		cookie, err := utils.ReadTextFile(cookieFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		opt := model.Option{
			Mid:     int(mid),
			Pages:   pages,
			Skip:    skip,
			Vorder:  vorder,
			Bvid:    "",
			Corder:  corder,
			Mapping: mapping,
			Cookie:  cookie,
			Output:  output,
		}
		sem := make(chan struct{}, workers)
		core.FindUser(sem, &opt)

	},
}
