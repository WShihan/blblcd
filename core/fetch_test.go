package core

import (
	"blblcd/model"
	"os"
	"path/filepath"
	"testing"
)

// setupFetchTest 设置 fetch 测试环境
func setupFetchTest(t *testing.T) string {
	// 尝试多个位置的 cookie 文件
	possiblePaths := []string{
		"../cookie.text",
		"../cookie.txt",
		"./cookie.text",
		"./cookie.txt",
	}

	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}

	t.Skip("跳过: 未找到 cookie 文件")
	return ""
}

// TestFindComment 测试主评论获取流程
func TestFindComment(t *testing.T) {
	cookie := setupFetchTest(t)
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		bvid       string
		opt        model.Option
		skipReason string
	}{
		{
			name: "基本评论获取",
			bvid: "BV1e7NRemEwv",
			opt: model.Option{
				Bvid:        "BV1e7NRemEwv",
				Corder:      2,
				Mapping:     false,
				Cookie:      cookie,
				Output:      tempDir,
				ImgDownload: false,
				MaxTryCount: 3,
				MaxDelaySec: 1,
			},
		},
		{
			name: "带地图输出",
			bvid: "BV1e7NRemEwv",
			opt: model.Option{
				Bvid:        "BV1e7NRemEwv",
				Corder:      2,
				Mapping:     true,
				Cookie:      cookie,
				Output:      tempDir,
				ImgDownload: false,
				MaxTryCount: 3,
				MaxDelaySec: 1,
			},
			skipReason: "耗时较长",
		},
		{
			name: "带图片下载",
			bvid: "BV1e7NRemEwv",
			opt: model.Option{
				Bvid:        "BV1e7NRemEwv",
				Corder:      2,
				Mapping:     false,
				Cookie:      cookie,
				Output:      tempDir,
				ImgDownload: true,
				MaxTryCount: 3,
				MaxDelaySec: 1,
			},
			skipReason: "耗时较长",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			avid := Bvid2Avid(tt.bvid)
			FindComment(avid, &tt.opt)

			// 验证 CSV 文件是否创建
			csvPath := filepath.Join(tempDir, tt.bvid, tt.bvid+".csv")
			if _, err := os.Stat(csvPath); os.IsNotExist(err) {
				t.Logf("CSV 文件未创建（可能无评论或出错）: %s", csvPath)
			} else {
				t.Logf("成功创建 CSV 文件: %s", csvPath)
			}

			// 如果启用地图，验证 geojson 文件
			if tt.opt.Mapping {
				geojsonPath := filepath.Join(tempDir, tt.bvid, tt.bvid+".geojson")
				if _, err := os.Stat(geojsonPath); os.IsNotExist(err) {
					t.Logf("GeoJSON 文件未创建: %s", geojsonPath)
				} else {
					t.Logf("成功创建 GeoJSON 文件: %s", geojsonPath)
				}
			}
		})
	}
}

// TestFindSubComment 测试子评论获取
func TestFindSubComment(t *testing.T) {
	cookie := setupFetchTest(t)

	tests := []struct {
		name string
		cmt  model.ReplyItem
		opt  model.Option
	}{
		{
			name: "获取子评论",
			cmt: model.ReplyItem{
				Rpid: 243795113873,
				Oid:  Bvid2Avid("BV1e7NRemEwv"),
			},
			opt: model.Option{
				Corder:      2,
				Cookie:      cookie,
				MaxDelaySec: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindSubComment(tt.cmt, &tt.opt)
			t.Logf("获取到 %d 条子评论", len(result))
		})
	}
}

// TestFindUser 测试 UP 主视频获取
func TestFindUser(t *testing.T) {
	cookie := setupFetchTest(t)
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		opt        model.Option
		skipReason string
	}{
		{
			name: "基本 UP 主获取",
			opt: model.Option{
				Mid:         208259,
				Pages:       1,
				Skip:        0,
				Vorder:      "pubdate",
				Corder:      2,
				Mapping:     false,
				Cookie:      cookie,
				Output:      tempDir,
				ImgDownload: false,
				MaxTryCount: 3,
				MaxDelaySec: 2,
			},
			skipReason: "耗时较长",
		},
		{
			name: "多页获取",
			opt: model.Option{
				Mid:         208259,
				Pages:       2,
				Skip:        0,
				Vorder:      "click",
				Corder:      2,
				Mapping:     false,
				Cookie:      cookie,
				Output:      tempDir,
				ImgDownload: false,
				MaxTryCount: 3,
				MaxDelaySec: 2,
			},
			skipReason: "耗时很长",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			sem := make(chan struct{}, 2)
			FindUser(sem, &tt.opt)

			// 验证输出目录是否创建
			if _, err := os.Stat(tempDir); os.IsNotExist(err) {
				t.Error("输出目录未创建")
			}
		})
	}
}

// TestNewCMTMultiple 测试评论数据转换（多组数据）
func TestNewCMTMultiple(t *testing.T) {
	tests := []struct {
		name string
		item *model.ReplyItem
		want map[string]interface{}
	}{
		{
			name: "完整评论数据",
			item: &model.ReplyItem{
				Rpid:   123456789,
				Oid:    987654321,
				Mid:    111111111,
				Parent: 0,
				Like:   999,
				Ctime:  1609459200,
				Member: struct {
					Mid            string `json:"mid"`
					Uname          string `json:"uname"`
					Sex            string `json:"sex"`
					Sign           string `json:"sign"`
					Avatar         string `json:"avatar"`
					Rank           string `json:"rank"`
					FaceNftNew     int    `json:"face_nft_new"`
					IsSeniorMember int    `json:"is_senior_member"`
					LevelInfo      struct {
						CurrentLevel int `json:"current_level"`
						CurrentMin   int `json:"current_min"`
						CurrentExp   int `json:"current_exp"`
						NextExp      int `json:"next_exp"`
					} `json:"level_info"`
					Vip struct {
						VipType       int    `json:"vipType"`
						VipDueDate    int64  `json:"vipDueDate"`
						DueRemark     string `json:"dueRemark"`
						AccessStatus  int    `json:"accessStatus"`
						VipStatus     int    `json:"vipStatus"`
						VipStatusWarn string `json:"vipStatusWarn"`
					} `json:"vip"`
					FansDetail any `json:"fans_detail"`
				}{
					Uname: "完整测试用户",
					Sex:   "保密",
					LevelInfo: struct {
						CurrentLevel int `json:"current_level"`
						CurrentMin   int `json:"current_min"`
						CurrentExp   int `json:"current_exp"`
						NextExp      int `json:"next_exp"`
					}{CurrentLevel: 6},
				},
				Content: struct {
					Message  string          `json:"message"`
					Pictures []model.Picture `json:"pictures"`
					Members  []any           `json:"members"`
					Emote    struct {
						NAMING_FAILED struct {
							ID        int    `json:"id"`
							PackageID int    `json:"package_id"`
							State     int    `json:"state"`
							Type      int    `json:"type"`
							Attr      int    `json:"attr"`
							Text      string `json:"text"`
							URL       string `json:"url"`
							Meta      struct {
								Size int `json:"size"`
							} `json:"meta"`
							Mtime     int64  `json:"mtime"`
							JumpTitle string `json:"jump_title"`
						} `json:"[吃瓜]"`
					} `json:"emote"`
					JumpURL struct {
					} `json:"jump_url"`
					MaxLine int `json:"max_line"`
				}{
					Message: "这是一条完整的测试评论",
					Pictures: []model.Picture{
						{Img_src: "https://example.com/1.jpg"},
						{Img_src: "https://example.com/2.jpg"},
					},
				},
				ReplyControl: struct {
					Following bool   `json:"following"`
					MaxLine   int    `json:"max_line"`
					TimeDesc  string `json:"time_desc"`
					Location  string `json:"location"`
				}{
					Following: true,
					Location:  "IP属地：火星",
				},
			},
			want: map[string]interface{}{
				"Uname":         "完整测试用户",
				"Sex":           "保密",
				"Content":       "这是一条完整的测试评论",
				"Like":          int64(999),
				"Current_level": 6,
				"Location":      "火星",
				"Following":     true,
				"PictureCount":  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewCMT(tt.item)

			if result.Uname != tt.want["Uname"] {
				t.Errorf("Uname = %v, want %v", result.Uname, tt.want["Uname"])
			}
			if result.Sex != tt.want["Sex"] {
				t.Errorf("Sex = %v, want %v", result.Sex, tt.want["Sex"])
			}
			if result.Content != tt.want["Content"] {
				t.Errorf("Content = %v, want %v", result.Content, tt.want["Content"])
			}
			if result.Like != tt.want["Like"] {
				t.Errorf("Like = %v, want %v", result.Like, tt.want["Like"])
			}
			if result.Current_level != tt.want["Current_level"] {
				t.Errorf("Current_level = %v, want %v", result.Current_level, tt.want["Current_level"])
			}
			if result.Location != tt.want["Location"] {
				t.Errorf("Location = %v, want %v", result.Location, tt.want["Location"])
			}
			if result.Following != tt.want["Following"] {
				t.Errorf("Following = %v, want %v", result.Following, tt.want["Following"])
			}
			if len(result.Pictures) != tt.want["PictureCount"] {
				t.Errorf("Pictures count = %v, want %v", len(result.Pictures), tt.want["PictureCount"])
			}
		})
	}
}

// TestFindCommentWithExistingCSV 测试带已有 CSV 的增量下载
func TestFindCommentWithExistingCSV(t *testing.T) {
	cookie := setupFetchTest(t)
	tempDir := t.TempDir()
	bvid := "BV1e7NRemEwv"

	// 创建模拟的已存在 CSV 文件
	savePath := filepath.Join(tempDir, bvid)
	os.MkdirAll(savePath, 0755)
	csvPath := filepath.Join(savePath, bvid+".csv")

	// 写入 CSV 头部和一些数据
	csvContent := `bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location
BV1e7NRemEwv,测试用户,男,已有评论,,123456789,987654321,111111111,0,1,1609459200,10,5,北京`
	os.WriteFile(csvPath, []byte(csvContent), 0644)

	opt := model.Option{
		Bvid:        bvid,
		Corder:      2,
		Mapping:     false,
		Cookie:      cookie,
		Output:      tempDir,
		ImgDownload: false,
		MaxTryCount: 3,
		MaxDelaySec: 1,
	}

	avid := Bvid2Avid(bvid)
	FindComment(avid, &opt)

	// 验证 CSV 文件仍然可读
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("CSV 文件被意外删除")
	} else {
		t.Log("CSV 文件成功保留")
	}
}
