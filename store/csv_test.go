package store

import (
	"blblcd/model"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestCMT2Record 测试评论转换为CSV记录
func TestCMT2Record(t *testing.T) {
	tests := []struct {
		name     string
		comment  model.Comment
		expected []string
	}{
		{
			name: "基本评论",
			comment: model.Comment{
				Bvid:          "BV1xxx",
				Uname:         "测试用户",
				Sex:           "男",
				Content:       "这是一条测试评论",
				Rpid:          123456789,
				Oid:           987654321,
				Mid:           111111111,
				Parent:        0,
				Fansgrade:     1,
				Ctime:         1609459200,
				Like:          100,
				Current_level: 5,
				Location:      "北京",
				Pictures:      []model.Picture{},
			},
			expected: []string{
				"BV1xxx", "测试用户", "男", "这是一条测试评论", "",
				"123456789", "987654321", "111111111", "0", "1",
				"1609459200", "100", "5", "北京",
			},
		},
		{
			name: "带图片的评论",
			comment: model.Comment{
				Bvid:          "BV2yyy",
				Uname:         "图片用户",
				Sex:           "女",
				Content:       "评论带图片",
				Rpid:          987654321,
				Oid:           123456789,
				Mid:           222222222,
				Parent:        123456789,
				Fansgrade:     0,
				Ctime:         1609545600,
				Like:          50,
				Current_level: 4,
				Location:      "上海",
				Pictures: []model.Picture{
					{Img_src: "https://example.com/pic1.jpg"},
					{Img_src: "https://example.com/pic2.jpg"},
				},
			},
			expected: []string{
				"BV2yyy", "图片用户", "女", "评论带图片",
				"https://example.com/pic1.jpg;https://example.com/pic2.jpg;",
				"987654321", "123456789", "222222222", "123456789", "0",
				"1609545600", "50", "4", "上海",
			},
		},
		{
			name: "空评论内容",
			comment: model.Comment{
				Bvid:          "BV3zzz",
				Uname:         "匿名用户",
				Sex:           "保密",
				Content:       "",
				Rpid:          111111111,
				Oid:           222222222,
				Mid:           333333333,
				Parent:        0,
				Fansgrade:     0,
				Ctime:         0,
				Like:          0,
				Current_level: 0,
				Location:      "",
				Pictures:      []model.Picture{},
			},
			expected: []string{
				"BV3zzz", "匿名用户", "保密", "", "",
				"111111111", "222222222", "333333333", "0", "0",
				"0", "0", "0", "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CMT2Record(tt.comment)

			if len(result) != len(tt.expected) {
				t.Errorf("记录长度不匹配: got %d, want %d", len(result), len(tt.expected))
			}

			for i := range tt.expected {
				if result[i] != tt.expected[i] {
					t.Errorf("字段 %d 不匹配: got %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestReadExistCommentRpids 测试读取已存在的评论ID
func TestReadExistCommentRpids(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		setup         func() string
		expectedLen   int
		expectedRpids []int64
		wantErr       bool
	}{
		{
			name: "文件不存在",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.csv")
			},
			expectedLen:   0,
			expectedRpids: []int64{},
			wantErr:       false,
		},
		{
			name: "空文件",
			setup: func() string {
				path := filepath.Join(tempDir, "empty.csv")
				file, _ := os.Create(path)
				file.Close()
				return path
			},
			expectedLen:   0,
			expectedRpids: []int64{},
			wantErr:       false,
		},
		{
			name: "有效的CSV文件",
			setup: func() string {
				path := filepath.Join(tempDir, "valid.csv")
				content := `bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location
BV1xxx,user1,男,content1,,123456,111,222,0,1,1609459200,10,5,北京
BV2yyy,user2,女,content2,,789012,333,444,0,0,1609545600,20,6,上海`
				os.WriteFile(path, []byte(content), 0644)
				return path
			},
			expectedLen:   2,
			expectedRpids: []int64{123456, 789012},
			wantErr:       false,
		},
		{
			name: "无rpid列的CSV",
			setup: func() string {
				path := filepath.Join(tempDir, "no_rpid.csv")
				content := `bvid,upname,sex,content
BV1xxx,user1,男,content1`
				os.WriteFile(path, []byte(content), 0644)
				return path
			},
			expectedLen:   0,
			expectedRpids: []int64{},
			wantErr:       false,
		},
		{
			name: "重复rpid",
			setup: func() string {
				path := filepath.Join(tempDir, "duplicate.csv")
				content := `bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location
BV1xxx,user1,男,content1,,123456,111,222,0,1,1609459200,10,5,北京
BV2yyy,user2,女,content2,,123456,333,444,0,0,1609545600,20,6,上海`
				os.WriteFile(path, []byte(content), 0644)
				return path
			},
			expectedLen:   1,
			expectedRpids: []int64{123456},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result, err := ReadExistCommentRpids(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadExistCommentRpids() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("结果长度 = %d, want %d", len(result), tt.expectedLen)
			}

			for _, rpid := range tt.expectedRpids {
				if !result[rpid] {
					t.Errorf("期望包含 rpid %d", rpid)
				}
			}
		})
	}
}

// TestSave2CSV 测试保存评论到CSV
func TestSave2CSV(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		basename    string
		comments    []model.Comment
		output      string
		downloadIMG bool
		setup       func(string)
		validate    func(string) bool
	}{
		{
			name:     "新建CSV文件",
			basename: "test_video",
			comments: []model.Comment{
				{
					Bvid:          "BV1xxx",
					Uname:         "用户1",
					Sex:           "男",
					Content:       "评论1",
					Rpid:          111,
					Oid:           222,
					Mid:           333,
					Parent:        0,
					Fansgrade:     1,
					Ctime:         1609459200,
					Like:          10,
					Current_level: 5,
					Location:      "北京",
				},
			},
			output:      tempDir,
			downloadIMG: false,
			validate: func(path string) bool {
				// 检查文件是否存在
				_, err := os.Stat(path)
				if err != nil {
					t.Logf("文件不存在: %v", err)
					return false
				}

				// 读取文件内容
				file, err := os.Open(path)
				if err != nil {
					t.Logf("无法打开文件: %v", err)
					return false
				}
				defer file.Close()

				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				if err != nil {
					t.Logf("读取CSV失败: %v", err)
					return false
				}

				// 应该有header + 1条数据
				if len(records) != 2 {
					t.Logf("记录数不正确: %d", len(records))
					return false
				}

				// 检查header
				if records[0][0] != "bvid" {
					t.Logf("header不正确: %s", records[0][0])
					return false
				}

				return true
			},
		},
		// 		{
		// 			name:     "追加到已有CSV",
		// 			basename: "append_video",
		// 			comments: []model.Comment{
		// 				{
		// 					Bvid:          "BV2yyy",
		// 					Uname:         "用户2",
		// 					Sex:           "女",
		// 					Content:       "评论2",
		// 					Rpid:          222,
		// 					Oid:           333,
		// 					Mid:           444,
		// 					Parent:        0,
		// 					Fansgrade:     0,
		// 					Ctime:         1609545600,
		// 					Like:          20,
		// 					Current_level: 6,
		// 					Location:      "上海",
		// 				},
		// 			},
		// 			output:      tempDir,
		// 			downloadIMG: false,
		// 			setup: func(dir string) {
		// 				// 预先创建CSV文件
		// 				path := filepath.Join(dir, "append_video.csv")
		// 				content := `bvid,upname,sex,content,pictures,rpid,oid,mid,parent,fans_grade,ctime,like,level,location
		// BV1xxx,用户1,男,评论1,,111,222,333,0,1,1609459200,10,5,北京`
		// 				os.WriteFile(path, []byte(content), 0644)
		// 			},
		// 			validate: func(path string) bool {
		// 				file, err := os.Open(path)
		// 				if err != nil {
		// 					return false
		// 				}
		// 				defer file.Close()

		// 				reader := csv.NewReader(file)
		// 				records, err := reader.ReadAll()
		// 				if err != nil {
		// 					return false
		// 				}

		// 				// 应该有header + 2条数据
		// 				return len(records) == 3
		// 			},
		// 		},
		{
			name:     "空评论列表",
			basename: "empty_comments",
			comments: []model.Comment{},
			output:   tempDir,
			validate: func(path string) bool {
				// 不应该创建文件
				_, err := os.Stat(path)
				return os.IsNotExist(err)
			},
		},
		{
			name:     "多评论写入",
			basename: "multi_comments",
			comments: []model.Comment{
				{
					Bvid:          "BV1xxx",
					Uname:         "用户1",
					Sex:           "男",
					Content:       "评论1",
					Rpid:          111,
					Oid:           222,
					Mid:           333,
					Parent:        0,
					Ctime:         1609459200,
					Like:          10,
					Current_level: 5,
					Location:      "北京",
				},
				{
					Bvid:          "BV1xxx",
					Uname:         "用户2",
					Sex:           "女",
					Content:       "评论2",
					Rpid:          222,
					Oid:           222,
					Mid:           444,
					Parent:        111,
					Ctime:         1609545600,
					Like:          20,
					Current_level: 6,
					Location:      "上海",
				},
				{
					Bvid:          "BV1xxx",
					Uname:         "用户3",
					Sex:           "保密",
					Content:       "评论3",
					Rpid:          333,
					Oid:           222,
					Mid:           555,
					Parent:        111,
					Ctime:         1609632000,
					Like:          30,
					Current_level: 4,
					Location:      "广州",
				},
			},
			output:      tempDir,
			downloadIMG: false,
			validate: func(path string) bool {
				file, err := os.Open(path)
				if err != nil {
					return false
				}
				defer file.Close()

				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				if err != nil {
					return false
				}

				// 应该有header + 3条数据
				return len(records) == 4
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(tt.output)
			}

			var csvMutex sync.Mutex
			var imageMutex sync.Mutex

			Save2CSV(&csvMutex, &imageMutex, tt.basename, tt.comments, tt.output, tt.downloadIMG)

			path := filepath.Join(tt.output, tt.basename+".csv")
			if !tt.validate(path) {
				fmt.Println(path)
				t.Error("验证失败")
			}
		})
	}
}

// TestSave2CSVConcurrent 测试并发写入CSV
func TestSave2CSVConcurrent(t *testing.T) {
	tempDir := t.TempDir()
	basename := "concurrent_test"

	var csvMutex sync.Mutex
	var imageMutex sync.Mutex

	// 并发写入多条评论
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			comments := []model.Comment{
				{
					Bvid:          "BV1xxx",
					Uname:         "用户",
					Content:       "评论",
					Rpid:          int64(1000 + index),
					Oid:           222,
					Mid:           int64(3000 + index),
					Ctime:         1609459200,
					Like:          int64(10 + index),
					Current_level: 5,
				},
			}
			Save2CSV(&csvMutex, &imageMutex, basename, comments, tempDir, false)
		}(i)
	}
	wg.Wait()

	// 验证文件内容
	path := filepath.Join(tempDir, basename+".csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("无法打开文件: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("读取CSV失败: %v", err)
	}

	// 应该有header + 5条数据
	if len(records) != 6 {
		t.Errorf("记录数不正确: got %d, want 6", len(records))
	}
}
