package cli

import (
	"blblcd/core"
	"blblcd/model"
	"blblcd/utils"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(videoCmd)
}

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "获取视频评论，支持单个和多个视频",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("please provide bvid")
			return
		}
		cookie, err := utils.ReadTextFile(cookieFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		utils.PrintLogo()

		sem := make(chan struct{}, workers)
		wg := sync.WaitGroup{}
		for i := range args {
			bvid := args[i]
			opt := model.Option{
				Bvid:        bvid,
				Corder:      corder,
				Mapping:     mapping,
				Cookie:      cookie,
				Output:      output,
				ImgDownload: imgDownload,
				MaxTryCount: maxTryCount,
				MaxDelaySec: maxDelaySec,
			}

			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					<-sem
				}()
				core.FindComment(core.Bvid2Avid(bvid), &opt)
			}()
		}
		wg.Wait()
	},
}
