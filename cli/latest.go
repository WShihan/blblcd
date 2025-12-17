package cli

import (
	"blblcd/core"
	"blblcd/model"
	"blblcd/utils"
	"fmt"
	"path"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(latestCmd)
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "获取up主最新视频的评论",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("please provide up mid")
			return
		}
		mid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		cookie, err := utils.ReadTextFile(cookieFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		utils.PrintLogo()
		videoList, err := core.FetchVideoList(mid, 1, "pubdate", cookie)
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(videoList.Data.List.Vlist) == 0 {
			fmt.Println("no video found")
			return
		}

		bvid := videoList.Data.List.Vlist[0].Bvid
		if utils.FileOrPathExists(path.Join(output, bvid)) {
			fmt.Printf("This video's comments already downloaded, BVID: %s\n", bvid)
			return
		}
		opt := model.Option{
			Bvid:        bvid,
			Corder:      corder,
			Mapping:     mapping,
			Cookie:      cookie,
			Output:      output,
			ImgDownload: imgDownload,
			MaxTryCount: maxTryCount,
		}

		core.FindComment(core.Bvid2Avid(bvid), &opt)

	},
}
