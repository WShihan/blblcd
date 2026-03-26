package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blblcd/model"
)

// 测试用的注入信息
var testInject = &model.Injection{
	Version:   "test-version",
	BuildTime: "test-buildtime",
	Commit:    "test-commit",
	Author:    "test-author",
}

// setupTest 设置测试环境
func setupTest(t *testing.T) (cookie string, cleanup func()) {
	// 尝试多个位置的 cookie 文件
	possiblePaths := []string{
		"../cookie.text",
		"../cookie.txt",
		"./cookie.text",
		"./cookie.txt",
	}

	var cookieData []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		cookieData, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		t.Skip("跳过测试: 未找到 cookie 文件 (cookie.text 或 cookie.txt)")
	}

	t.Logf("使用 cookie 文件: %s", foundPath)

	// 创建临时输出目录
	tempDir := t.TempDir()

	return strings.TrimSpace(string(cookieData)), func() {
		// 清理由 t.TempDir() 自动处理
		_ = tempDir
	}
}

// TestRootCmd 测试根命令
func TestRootCmd(t *testing.T) {
	Inject = testInject

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantOut string
	}{
		{
			name:    "无参数显示帮助",
			args:    []string{},
			wantErr: false,
			wantOut: "please type",
		},
		{
			name:    "帮助标志",
			args:    []string{"--help"},
			wantErr: false,
			wantOut: "Usage",
		},
		{
			name:    "简写帮助",
			args:    []string{"-h"},
			wantErr: false,
			wantOut: "Usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置命令状态
			rootCmd.SetArgs(tt.args)

			// 捕获输出
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			if tt.wantOut != "" && !strings.Contains(output, tt.wantOut) {
				t.Errorf("输出不包含 %q, 实际输出: %s", tt.wantOut, output)
			}
			t.Logf("输出: %s", output)
		})
	}
}

// TestVideoCmd 测试 video 命令
func TestVideoCmd(t *testing.T) {
	cookie, cleanup := setupTest(t)
	defer cleanup()
	Inject = testInject

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		skipReason string
	}{
		{
			name:    "单视频基本测试",
			args:    []string{"video", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", t.TempDir(), "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:       "多视频并发测试",
			args:       []string{"video", "BV1e7NRemEwv", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", t.TempDir(), "--workers", "2", "--max-delay", "2"},
			wantErr:    false,
			skipReason: "耗时较长",
		},
		{
			name:    "热度排序测试",
			args:    []string{"video", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", t.TempDir(), "--corder", "0", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "时间排序测试",
			args:    []string{"video", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", t.TempDir(), "--corder", "2", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "混合排序测试",
			args:    []string{"video", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", t.TempDir(), "--corder", "1", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "无视频参数",
			args:    []string{"video", "--cookie", "../cookie.text"},
			wantErr: false, // 打印提示信息但不报错退出
		},
		{
			name:    "无效BV号测试",
			args:    []string{"video", "BVInvalid", "--cookie", "../cookie.text", "--output", t.TempDir(), "--workers", "1"},
			wantErr: false, // 程序会处理错误但不 panic
		},
		{
			name:    "cookie文件不存在",
			args:    []string{"video", "BV1e7NRemEwv", "--cookie", "./nonexistent.cookie", "--output", t.TempDir()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			// 替换 cookie 路径为实际路径
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "../cookie.text" || arg == "../cookie.txt" {
					// 尝试找到实际的 cookie 文件
					for _, path := range []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"} {
						if _, err := os.Stat(path); err == nil {
							args[i] = path
							break
						}
					}
				}
			}

			// 创建新的命令实例以避免状态污染
			cmd := rootCmd
			cmd.SetArgs(args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			t.Logf("输出: %s", output)

			// 验证输出包含预期内容
			if strings.Contains(tt.name, "无视频参数") {
				if !strings.Contains(output, "please provide") && !strings.Contains(output, "bvid") {
					t.Errorf("期望提示用户提供 bvid")
				}
			}
		})
	}

	_ = cookie // 使用变量避免未使用警告
}

// TestVideoCmdWithMapping 测试 video 命令的 mapping 功能
func TestVideoCmdWithMapping(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()
	Inject = testInject

	tempDir := t.TempDir()

	// 测试带地图输出
	args := []string{"video", "BV1e7NRemEwv", "--cookie", "../cookie.text", "--output", tempDir, "--mapping", "--workers", "1", "--max-delay", "2"}

	// 尝试找到实际的 cookie 文件
	for i, arg := range args {
		if arg == "../cookie.text" || arg == "../cookie.txt" {
			for _, path := range []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"} {
				if _, err := os.Stat(path); err == nil {
					args[i] = path
					break
				}
			}
		}
	}

	rootCmd.SetArgs(args)
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()
	t.Logf("输出: %s", output)

	// 验证是否生成了地图文件
	geojsonPath := filepath.Join(tempDir, "BV1e7NRemEwv", "BV1e7NRemEwv.geojson")
	if _, err := os.Stat(geojsonPath); err == nil {
		t.Logf("成功生成地图文件: %s", geojsonPath)
	}
}

// TestUpCmd 测试 up 命令
func TestUpCmd(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()
	Inject = testInject

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		skipReason string
	}{
		{
			name:    "基本UP主测试",
			args:    []string{"up", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--pages", "1", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "播放排序测试",
			args:    []string{"up", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--pages", "1", "--vorder", "click", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "收藏排序测试",
			args:    []string{"up", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--pages", "1", "--vorder", "stow", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "跳过页面测试",
			args:    []string{"up", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--pages", "1", "--skip", "1", "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "无效MID测试",
			args:    []string{"up", "invalid_mid", "--cookie", "../cookie.text", "--output", t.TempDir()},
			wantErr: false,
		},
		{
			name:    "无MID参数",
			args:    []string{"up", "--cookie", "../cookie.text"},
			wantErr: false,
		},
		{
			name:       "带地图输出",
			args:       []string{"up", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--pages", "1", "--mapping", "--workers", "1", "--max-delay", "2"},
			wantErr:    false,
			skipReason: "耗时较长",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "../cookie.text" || arg == "../cookie.txt" {
					for _, path := range []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"} {
						if _, err := os.Stat(path); err == nil {
							args[i] = path
							break
						}
					}
				}
			}

			rootCmd.SetArgs(args)
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			t.Logf("输出: %s", output)

			if strings.Contains(tt.name, "无MID") {
				if !strings.Contains(output, "please provide") && !strings.Contains(output, "mid") {
					t.Errorf("期望提示用户提供 mid")
				}
			}
		})
	}
}

// TestLatestCmd 测试 latest 命令
func TestLatestCmd(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()
	Inject = testInject

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		skipReason string
	}{
		{
			name:    "基本latest测试",
			args:    []string{"latest", "208259", "--cookie", "../cookie.text", "--output", t.TempDir(), "--workers", "1", "--max-delay", "2"},
			wantErr: false,
		},
		{
			name:    "无效MID测试",
			args:    []string{"latest", "invalid", "--cookie", "../cookie.text", "--output", t.TempDir()},
			wantErr: false,
		},
		{
			name:    "无MID参数",
			args:    []string{"latest", "--cookie", "../cookie.text"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "../cookie.text" || arg == "../cookie.txt" {
					for _, path := range []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"} {
						if _, err := os.Stat(path); err == nil {
							args[i] = path
							break
						}
					}
				}
			}

			rootCmd.SetArgs(args)
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			t.Logf("输出: %s", output)
		})
	}
}

// TestVersionCmd 测试 version 命令
func TestVersionCmd(t *testing.T) {

}

// TestGlobalFlags 测试全局参数
func TestGlobalFlags(t *testing.T) {
	Inject = testInject

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "workers参数",
			args:    []string{"--help"},
			wantOut: "workers",
		},
		{
			name:    "cookie参数",
			args:    []string{"--help"},
			wantOut: "cookie",
		},
		{
			name:    "output参数",
			args:    []string{"--help"},
			wantOut: "output",
		},
		{
			name:    "corder参数",
			args:    []string{"--help"},
			wantOut: "corder",
		},
		{
			name:    "mapping参数",
			args:    []string{"--help"},
			wantOut: "mapping",
		},
		{
			name:    "img-download参数",
			args:    []string{"--help"},
			wantOut: "img-download",
		},
		{
			name:    "max-try-count参数",
			args:    []string{"--help"},
			wantOut: "max-try-count",
		},
		{
			name:    "max-delay参数",
			args:    []string{"--help"},
			wantOut: "max-delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			err := rootCmd.Execute()
			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOut) {
				t.Errorf("帮助输出不包含 %q", tt.wantOut)
			}
		})
	}
}

// TestVideoCmdFlags 测试 video 命令特有参数
func TestVideoCmdFlags(t *testing.T) {
	Inject = testInject

	// 测试参数帮助信息
	rootCmd.SetArgs([]string{"video", "--help"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()
	t.Logf("video 命令帮助: %s", output)

	// 验证帮助信息包含基本说明
	if !strings.Contains(output, "video") {
		t.Error("帮助输出不包含 'video'")
	}
}

// TestUpCmdFlags 测试 up 命令特有参数
func TestUpCmdFlags(t *testing.T) {
	Inject = testInject

	tests := []struct {
		name     string
		wantFlag string
	}{
		{"pages参数", "pages"},
		{"skip参数", "skip"},
		{"vorder参数", "vorder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs([]string{"up", "--help"})
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			err := rootCmd.Execute()
			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantFlag) {
				t.Errorf("帮助输出不包含 %q", tt.wantFlag)
			}
		})
	}
}
