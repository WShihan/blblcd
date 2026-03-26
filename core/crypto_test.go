package core

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestBvid2Avid 测试BV号转AV号
func TestBvid2Avid(t *testing.T) {
	tests := []struct {
		name     string
		bvid     string
		expected int64
	}{
		{
			name:     "常见BV号",
			bvid:     "BV1e7NRemEwv",
			expected: 113976755094913,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Bvid2Avid(tt.bvid)
			t.Logf("BV号 %s -> AV号 %d", tt.bvid, result)

			// 验证可逆性
			back := Avid2Bvid(result)
			if back != tt.bvid {
				t.Errorf("转换不可逆: Bvid2Avid(%s) = %d, Avid2Bvid(%d) = %s",
					tt.bvid, result, result, back)
			}

			if result != tt.expected && tt.expected != 0 {
				t.Errorf("期望值 %d, 实际 %d", tt.expected, result)
			}
		})
	}
}

// TestAvid2Bvid 测试AV号转BV号
func TestAvid2Bvid(t *testing.T) {
	tests := []struct {
		name     string
		avid     int64
		expected string
	}{
		{
			name:     "常见AV号",
			avid:     113976755094913,
			expected: "BV1e7NRemEwv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Avid2Bvid(tt.avid)
			t.Logf("AV号 %d -> BV号 %s", tt.avid, result)

			// 验证可逆性
			back := Bvid2Avid(result)
			if back != tt.avid {
				t.Errorf("转换不可逆: Avid2Bvid(%d) = %s, Bvid2Avid(%s) = %d",
					tt.avid, result, result, back)
			}

			if result != tt.expected && tt.expected != "" {
				t.Errorf("期望值 %s, 实际 %s", tt.expected, result)
			}
		})
	}
}

// TestBidirectionalConversion 测试双向转换一致性
func TestBidirectionalConversion(t *testing.T) {
	// 测试一系列AV号的往返转换
	testAvids := []int64{
		113976755094913,
	}

	for _, avid := range testAvids {
		bvid := Avid2Bvid(avid)
		back := Bvid2Avid(bvid)

		if back != avid {
			t.Errorf("双向转换失败: avid=%d -> bvid=%s -> avid=%d", avid, bvid, back)
		} else {
			t.Logf("双向转换成功: avid=%d <-> bvid=%s", avid, bvid)
		}
	}
}

// TestSwapString 测试字符串交换
func TestSwapString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		x, y     int
		expected string
	}{
		{
			name:     "交换中间字符",
			s:        "BV1e7NRemEwv",
			expected: "BV1E7NRemewv",
			x:        3,
			y:        9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := swapString(tt.s, tt.x, tt.y)
			fmt.Println(result)
			if result != tt.expected {
				t.Errorf("swapString(%s, %d, %d) = %s, expected %s",
					tt.s, tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

// TestGetMixinKey 测试mixin key生成
func TestGetMixinKey(t *testing.T) {
	tests := []struct {
		name    string
		orig    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "有效输入",
			orig:    "ce1f45942e44304d3b5f375c98f5cbe5cce1f45942e44304d3b5f375c98f5cbe5", // 64字符
			wantErr: false,
		},
		{
			name:    "短输入",
			orig:    "short",
			wantErr: true,
			errMsg:  "invalid mixin key length",
		},
		{
			name:    "空输入",
			orig:    "",
			wantErr: true,
			errMsg:  "invalid mixin key length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getMixinKey(tt.orig)

			if (err != nil) != tt.wantErr {
				t.Errorf("getMixinKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("错误消息不包含 %q, 实际: %v", tt.errMsg, err)
				}
			}

			if err == nil {
				// 验证结果长度
				if len(result) != 32 {
					t.Errorf("结果长度不是32: %d", len(result))
				}
				t.Logf("生成的mixin key: %s", result)
			}
		})
	}
}

// TestSignAndGenerateURL 测试URL签名（需要cookie）
func TestSignAndGenerateURL(t *testing.T) {
	// 尝试读取cookie文件
	cookiePaths := []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"}
	var cookie string
	var found bool

	for _, path := range cookiePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			cookie = string(data)
			found = true
			t.Logf("使用cookie文件: %s", path)
			break
		}
	}

	if !found {
		t.Skip("跳过: 未找到cookie文件")
	}

	// 重置缓存，确保测试时重新获取
	cache = sync.Map{}
	lastUpdateTime = time.Time{}

	tests := []struct {
		name    string
		urlStr  string
		wantErr bool
	}{
		{
			name:    "有效URL",
			urlStr:  "https://api.bilibili.com/x/v2/reply/wbi/main?oid=123&type=1",
			wantErr: false,
		},
		{
			name:    "带参数的URL",
			urlStr:  "https://api.bilibili.com/x/space/wbi/arc/search?mid=208259&pn=1&ps=25",
			wantErr: false,
		},
		{
			name:    "无效URL格式",
			urlStr:  "://invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SignAndGenerateURL(tt.urlStr, cookie)

			if (err != nil) != tt.wantErr {
				t.Errorf("SignAndGenerateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// 验证结果包含wts和w_rid参数
				if !strings.Contains(result, "wts=") {
					t.Error("签名后的URL不包含wts参数")
				}
				if !strings.Contains(result, "w_rid=") {
					t.Error("签名后的URL不包含w_rid参数")
				}
				t.Logf("签名后的URL: %s", result)
			}
		})
	}
}

// TestGetWbiKeys 测试获取WBI密钥（需要真实API调用）
func TestGetWbiKeys(t *testing.T) {
	// 尝试读取cookie文件
	cookiePaths := []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"}
	var cookie string
	var found bool

	for _, path := range cookiePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			cookie = string(data)
			found = true
			t.Logf("使用cookie文件: %s", path)
			break
		}
	}

	if !found {
		t.Skip("跳过: 未找到cookie文件")
	}

	// 重置缓存
	cache = sync.Map{}
	lastUpdateTime = time.Time{}

	imgKey, subKey, err := getWbiKeys(cookie)
	if err != nil {
		t.Errorf("getWbiKeys() error = %v", err)
		return
	}

	if imgKey == "" {
		t.Error("imgKey为空")
	}
	if subKey == "" {
		t.Error("subKey为空")
	}

	t.Logf("获取到WBI密钥: imgKey=%s, subKey=%s", imgKey, subKey)

	// 验证密钥格式（32字符十六进制）
	if len(imgKey) != 32 {
		t.Errorf("imgKey长度不是32: %d", len(imgKey))
	}
	if len(subKey) != 32 {
		t.Errorf("subKey长度不是32: %d", len(subKey))
	}
}

// TestCacheMechanism 测试缓存机制
func TestCacheMechanism(t *testing.T) {
	// 尝试读取cookie文件
	cookiePaths := []string{"../cookie.text", "../cookie.txt", "./cookie.text", "./cookie.txt"}
	var cookie string
	var found bool

	for _, path := range cookiePaths {
		data, err := os.ReadFile(path)
		if err == nil {
			cookie = string(data)
			found = true
			break
		}
	}

	if !found {
		t.Skip("跳过: 未找到cookie文件")
	}

	// 重置缓存
	cache = sync.Map{}
	lastUpdateTime = time.Time{}

	// 第一次调用，应该获取新密钥
	imgKey1, subKey1, err := getWbiKeysCached(cookie)
	if err != nil {
		t.Fatalf("第一次调用失败: %v", err)
	}

	// 第二次调用，应该使用缓存
	imgKey2, subKey2, err := getWbiKeysCached(cookie)
	if err != nil {
		t.Fatalf("第二次调用失败: %v", err)
	}

	// 验证缓存有效（两次结果相同）
	if imgKey1 != imgKey2 {
		t.Error("缓存无效: imgKey不一致")
	}
	if subKey1 != subKey2 {
		t.Error("缓存无效: subKey不一致")
	}

	t.Log("缓存机制工作正常")
}
