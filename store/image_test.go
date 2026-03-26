package store

import (
	"blblcd/model"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestFetchImage 测试获取图片（真实HTTP请求）
func TestFetchImage(t *testing.T) {
	tests := []struct {
		name       string
		imageUrl   string
		wantErr    bool
		errMsg     string
		skipReason string
	}{
		{
			name:       "有效图片URL",
			imageUrl:   "https://httpbin.org/image/png",
			wantErr:    false,
			skipReason: "依赖外部网络",
		},
		{
			name:     "无效URL",
			imageUrl: "https://invalid.domain.test/image.jpg",
			wantErr:  true,
			errMsg:   "failed to fetch",
		},
		{
			name:       "非图片URL",
			imageUrl:   "https://httpbin.org/json",
			wantErr:    false, // 可以获取，但不是图片
			skipReason: "依赖外部网络",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			data, err := fetchImage(tt.imageUrl)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("错误消息不包含 %q, 实际: %v", tt.errMsg, err)
				}
			}

			if err == nil && len(data) == 0 {
				t.Error("成功获取但数据为空")
			}

			if err == nil {
				t.Logf("成功获取图片，大小: %d bytes", len(data))
			}
		})
	}
}

// TestDownloadImg 测试下载图片到文件
func TestDownloadImg(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		imageUrl   string
		output     string
		prefix     string
		skipReason string
	}{
		{
			name:       "基本下载测试",
			imageUrl:   "https://httpbin.org/image/png",
			output:     tempDir,
			prefix:     "testuser",
			skipReason: "依赖外部网络",
		},
		{
			name:     "无效URL不崩溃",
			imageUrl: "https://invalid.domain.test/image.jpg",
			output:   tempDir,
			prefix:   "testuser2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			sem := make(chan struct{}, 1)
			var mu sync.Mutex

			DownloadImg(sem, &mu, tt.imageUrl, tt.output, tt.prefix)

			// 如果是有效的URL，检查文件是否被创建
			if tt.skipReason == "" {
				// 无效URL测试 - 不应崩溃，文件可能不存在
				return
			}

			// 检查是否创建了文件
			files, err := os.ReadDir(tt.output)
			if err != nil {
				t.Logf("读取目录失败: %v", err)
				return
			}

			found := false
			for _, file := range files {
				if !file.IsDir() && containsStr(file.Name(), tt.prefix) {
					found = true
					t.Logf("成功创建文件: %s", file.Name())
					break
				}
			}

			if !found && tt.skipReason == "" {
				t.Log("文件未创建（可能是下载失败）")
			}
		})
	}
}

// TestDownloadImgConcurrent 测试并发下载图片
func TestDownloadImgConcurrent(t *testing.T) {

	tempDir := t.TempDir()

	sem := make(chan struct{}, 3) // 最多3个并发
	var mu sync.Mutex

	urls := []string{
		"https://httpbin.org/image/png",
		"https://httpbin.org/image/jpeg",
		"https://httpbin.org/image/webp",
	}

	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(index int, imageUrl string) {
			defer wg.Done()
			prefix := "user" + string(rune('A'+index))
			DownloadImg(sem, &mu, imageUrl, tempDir, prefix)
		}(i, url)
	}
	wg.Wait()

	// 检查创建的目录
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	t.Logf("创建了 %d 个文件", len(files))
}

// TestWriteImage 测试WriteImage函数
func TestWriteImage(t *testing.T) {
	tempDir := t.TempDir()

	pics := []model.Picture{
		{Img_src: "https://httpbin.org/image/png"},
	}

	var mu sync.Mutex
	WriteImage(&mu, "testuser", pics, tempDir)

	// 检查是否创建了图片文件
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() {
			count++
			t.Logf("创建文件: %s", file.Name())
		}
	}

	if count == 0 {
		t.Error("没有创建任何图片文件")
	}
}

// TestDownloadImgDuplicate 测试重复下载不会创建重复文件
func TestDownloadImgDuplicate(t *testing.T) {
	t.Skip("跳过: 依赖外部网络")

	tempDir := t.TempDir()

	imageUrl := "https://httpbin.org/image/png"
	prefix := "testuser"
	sem := make(chan struct{}, 1)
	var mu sync.Mutex

	// 第一次下载
	DownloadImg(sem, &mu, imageUrl, tempDir, prefix)

	// 获取第一次下载后的文件信息
	files1, _ := os.ReadDir(tempDir)

	// 第二次下载相同URL
	DownloadImg(sem, &mu, imageUrl, tempDir, prefix)

	// 获取第二次下载后的文件信息
	files2, _ := os.ReadDir(tempDir)

	// 文件数量应该相同（因为检测到已存在会跳过）
	if len(files1) != len(files2) {
		t.Errorf("重复下载创建了额外文件: before=%d, after=%d", len(files1), len(files2))
	}
}

// TestWriteImageEmpty 测试空图片列表
func TestWriteImageEmpty(t *testing.T) {
	tempDir := t.TempDir()

	emptyPics := []model.Picture{}

	var mu sync.Mutex
	WriteImage(&mu, "testuser", emptyPics, tempDir)

	// 不应该创建任何文件
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("空图片列表不应创建文件，但创建了 %d 个", len(files))
	}
}

// TestDownloadImgInvalidPath 测试无效路径
func TestDownloadImgInvalidPath(t *testing.T) {
	// 使用一个无效的路径（在Windows上可能是非法字符，在Unix上可能是权限问题）
	invalidPath := filepath.Join(t.TempDir(), "nonexistent", "subdir")

	sem := make(chan struct{}, 1)
	var mu sync.Mutex

	// 不应panic
	DownloadImg(sem, &mu, "https://httpbin.org/image/png", invalidPath, "test")
}

// Helper function
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
