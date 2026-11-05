package model

type Option struct {
	Cookie      string
	Mid         int64
	Pages       int
	Skip        int
	Vorder      string
	Bvid        string
	Corder      int
	Output      string
	Mapping     bool
	ImgDownload bool
	MaxTryCount int
	MaxDelaySec float64
	// Concurrency 单个视频内并发请求数（主评论翻页、楼中楼及其分页）。
	// <=0 时使用默认值。
	Concurrency int
}
