package core

import (
	"blblcd/model"
	"log"
	"os"
	"strconv"
	"testing"
)

var (
	Cookie string
)

func init() {
	// 尝试多个位置的cookie文件
	possiblePaths := []string{
		"../cookie.text",
		"../cookie.txt",
		"./cookie.text",
		"./cookie.txt",
	}

	var found bool
	for _, path := range possiblePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			Cookie = string(data)
			found = true
			log.Printf("使用cookie文件: %s", path)
			break
		}
	}

	if !found {
		log.Println("警告: 未找到cookie文件，部分测试将被跳过")
	}
}

// TestFetchCmt 测试获取主评论（真实API）
func TestFetchCmt(t *testing.T) {
	if Cookie == "" {
		t.Skip("跳过: 未找到cookie文件")
	}

	tests := []struct {
		name       string
		bvid       string
		order      int
		offset     string
		skipReason string
	}{
		{
			name:  "按时间排序",
			bvid:  "BV1e7NRemEwv",
			order: 2,
		},
		{
			name:  "按热度排序",
			bvid:  "BV1e7NRemEwv",
			order: 0,
		},
		{
			name:  "混合排序",
			bvid:  "BV1e7NRemEwv",
			order: 1,
		},
		{
			name:   "带分页偏移",
			bvid:   "BV1e7NRemEwv",
			order:  2,
			offset: `{"offset":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			oid := Bvid2Avid(tt.bvid)
			result, err := FetchComment(strconv.FormatInt(oid, 10), tt.order, Cookie, tt.offset)

			if err != nil {
				t.Errorf("FetchComment() error = %v", err)
				return
			}

			// 验证响应结构
			if result.Code != 0 {
				t.Errorf("API返回错误: code=%d, message=%s", result.Code, result.Message)
				return
			}

			t.Logf("获取到 %d 条评论", len(result.Data.Replies))
			t.Logf("Cursor: %+v", result.Data.Cursor)

			// 验证字段
			if result.Data.Page.Size > 0 {
				t.Logf("页面大小: %d", result.Data.Page.Size)
			}
		})
	}
}

// TestFetchSubCmt 测试获取子评论
func TestFetchSubCmt(t *testing.T) {
	if Cookie == "" {
		t.Skip("跳过: 未找到cookie文件")
	}

	tests := []struct {
		name string
		bvid string
		rpid int64
		next int
	}{
		{
			name: "获取子评论第一页",
			bvid: "BV1e7NRemEwv",
			rpid: 243795113873,
			next: 1,
		},
		{
			name: "获取子评论多页",
			bvid: "BV1e7NRemEwv",
			rpid: 243795113873,
			next: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid := Bvid2Avid(tt.bvid)
			result, err := FetchSubComment(strconv.FormatInt(oid, 10), tt.rpid, tt.next, Cookie)

			if err != nil {
				t.Errorf("FetchSubComment() error = %v", err)
				return
			}

			if result.Code != 0 {
				t.Errorf("API返回错误: code=%d, message=%s", result.Code, result.Message)
				return
			}

			t.Logf("获取到 %d 条子评论", len(result.Data.Replies))
		})
	}
}

// TestFetchSubCmtCount 测试获取评论数量
func TestFetchSubCmtCount(t *testing.T) {
	tests := []struct {
		name string
		bvid string
	}{
		{
			name: "获取评论数量",
			bvid: "BV1e7NRemEwv",
		},
		{
			name: "另一个视频",
			bvid: "BV1e7NRemEwv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oid := Bvid2Avid(tt.bvid)
			count, err := FetchCount(strconv.FormatInt(oid, 10))

			if err != nil {
				t.Errorf("FetchCount() error = %v", err)
				return
			}

			t.Logf("视频 %s 共有 %d 条评论", tt.bvid, count)

			if count < 0 {
				t.Error("评论数量不能为负数")
			}
		})
	}
}

// TestFetchVideoList 测试获取UP主视频列表
func TestFetchVideoList(t *testing.T) {
	if Cookie == "" {
		t.Skip("跳过: 未找到cookie文件")
	}

	tests := []struct {
		name  string
		mid   int64
		page  int
		order string
	}{
		{
			name:  "按发布时间",
			mid:   208259,
			page:  1,
			order: "pubdate",
		},
		{
			name:  "按播放量",
			mid:   208259,
			page:  1,
			order: "click",
		},
		{
			name:  "按收藏数",
			mid:   208259,
			page:  1,
			order: "stow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FetchVideoList(tt.mid, tt.page, tt.order, Cookie)

			if err != nil {
				t.Errorf("FetchVideoList() error = %v", err)
				return
			}

			if result.Code != 0 {
				t.Errorf("API返回错误: code=%d, message=%s", result.Code, result.Message)
				return
			}

			t.Logf("获取到 %d 个视频", len(result.Data.List.Vlist))

			if len(result.Data.List.Vlist) > 0 {
				first := result.Data.List.Vlist[0]
				t.Logf("第一个视频: %s - %s", first.Bvid, first.Title)
			}
		})
	}
}

// TestNewCMT 测试评论数据转换
func TestNewCMT(t *testing.T) {
	tests := []struct {
		name     string
		item     model.ReplyItem
		expected model.Comment
	}{
		{
			name: "基本评论",
			item: model.ReplyItem{
				Rpid:   123456789,
				Oid:    987654321,
				Mid:    111111111,
				Parent: 0,
				Like:   100,
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
					Uname: "测试用户",
					Sex:   "男",
					LevelInfo: struct {
						CurrentLevel int `json:"current_level"`
						CurrentMin   int `json:"current_min"`
						CurrentExp   int `json:"current_exp"`
						NextExp      int `json:"next_exp"`
					}{CurrentLevel: 5},
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
					Message:  "这是一条测试评论",
					Pictures: []model.Picture{},
				},
				ReplyControl: struct {
					Following bool   `json:"following"`
					MaxLine   int    `json:"max_line"`
					TimeDesc  string `json:"time_desc"`
					Location  string `json:"location"`
				}{
					Location: "IP属地：北京",
				},
			},
			expected: model.Comment{
				Uname:         "测试用户",
				Sex:           "男",
				Content:       "这是一条测试评论",
				Rpid:          123456789,
				Oid:           987654321,
				Mid:           111111111,
				Parent:        0,
				Ctime:         1609459200,
				Like:          100,
				Current_level: 5,
				Location:      "北京",
			},
		},
		{
			name: "子评论",
			item: model.ReplyItem{
				Rpid:   987654321,
				Oid:    987654321,
				Mid:    222222222,
				Parent: 123456789,
				Like:   50,
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
					Uname: "回复用户",
					Sex:   "女",
					LevelInfo: struct {
						CurrentLevel int `json:"current_level"`
						CurrentMin   int `json:"current_min"`
						CurrentExp   int `json:"current_exp"`
						NextExp      int `json:"next_exp"`
					}{CurrentLevel: 3},
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
					Message: "这是一条回复",
				},
				ReplyControl: struct {
					Following bool   `json:"following"`
					MaxLine   int    `json:"max_line"`
					TimeDesc  string `json:"time_desc"`
					Location  string `json:"location"`
				}{
					Location: "IP属地：上海",
				},
			},
			expected: model.Comment{
				Uname:         "回复用户",
				Sex:           "女",
				Content:       "这是一条回复",
				Rpid:          987654321,
				Oid:           987654321,
				Mid:           222222222,
				Parent:        123456789,
				Current_level: 3,
				Location:      "上海",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewCMT(&tt.item)

			if result.Uname != tt.expected.Uname {
				t.Errorf("Uname = %v, want %v", result.Uname, tt.expected.Uname)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content = %v, want %v", result.Content, tt.expected.Content)
			}
			if result.Rpid != tt.expected.Rpid {
				t.Errorf("Rpid = %v, want %v", result.Rpid, tt.expected.Rpid)
			}
			if result.Parent != tt.expected.Parent {
				t.Errorf("Parent = %v, want %v", result.Parent, tt.expected.Parent)
			}
			if result.Location != tt.expected.Location {
				t.Errorf("Location = %v, want %v", result.Location, tt.expected.Location)
			}

			// 验证Bvid转换
			if result.Bvid == "" {
				t.Error("Bvid不应为空")
			}
			t.Logf("生成的Bvid: %s", result.Bvid)
		})
	}
}

// TestCorderValues 测试评论排序参数
func TestCorderValues(t *testing.T) {
	if Cookie == "" {
		t.Skip("跳过: 未找到cookie文件")
	}

	bvid := "BV1e7NRemEwv"
	oid := Bvid2Avid(bvid)

	corders := []int{0, 1, 2, 3}
	for _, corder := range corders {
		t.Run(strconv.Itoa(corder), func(t *testing.T) {
			result, err := FetchComment(strconv.FormatInt(oid, 10), corder, Cookie, "")
			if err != nil {
				t.Errorf("corder=%d 时出错: %v", corder, err)
				return
			}
			if result.Code != 0 {
				t.Errorf("corder=%d 时API返回错误: %s", corder, result.Message)
				return
			}
			t.Logf("corder=%d 成功", corder)
		})
	}
}
