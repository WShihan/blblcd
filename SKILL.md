---
name: blblcd
description: 通过命令行下载哔哩哔哩（B站）视频的评论。支持下载单个/多个视频；按照up主视频列表下载；最新视频下载；可下载评论图片；可统计评论地区输出地图报告。
---

# 技能用途

* 获取单个或多个 B 站视频的评论数据（含楼中楼）
* 批量获取 UP 主所有/部分视频的评论
* 下载评论中的图片
* 生成评论时间/热度分布地图

## 触发条件
* 用户提到「B站评论」「哔哩哔哩评论」「下载视频评论」「获取弹幕/评论数据」等关键词
* 用户提供 B 站视频链接或 BV 号
* 用户需要分析视频评论数据


# 工作流

对于所有请求，按照以下步骤执行。

## Step 1：检查工具是否安装

执行：

```bash
blblcd version
```

如果返回版本号：

继续执行。

如果返回：

```text
command not found
```

进入【安装流程】。

---

## Step 2：检查 Cookie 配置

读取环境变量：

```bash
BLBLCD_COOKIE_PATH
```

检查：

* 环境变量存在
* 文件可读

如果 Cookie 不存在：

向用户说明：

```text
需要先导出 B 站登录 Cookie，并设置环境变量：

BLBLCD_COOKIE_PATH=/path/to/cookie.txt
```

停止执行。

---

## Step 3：解析用户输入

支持以下输入形式。

### BV号

例如：

```text
BV1xx411c7mD
```

直接作为 bvid 使用。

---

### 视频链接

例如：

```text
https://www.bilibili.com/video/BV1sULg65Evk
```

自动提取：

```text
BV1sULg65Evk
```

作为 bvid。

---

### UP主页链接

例如：

```text
https://space.bilibili.com/12345678
```

自动提取：

```text
12345678
```

作为 uid。

---

## Step 4：识别用户需求

将请求归类为以下类型。

### 类型 A：单视频评论

例如：

```text
下载这个视频的评论
```

执行：

```bash
blblcd video <bvid>
```

---

### 类型 B：多个视频评论

例如：

```text
下载这几个视频的评论
```

逐个执行：

```bash
blblcd video <bvid>
```

---

### 类型 C：UP主评论批量下载

例如：

```text
下载这个 UP 主最近的视频评论
```

执行：

```bash
blblcd up <uid>
```

---

### 类型 D：评论分析

例如：

```text
分析评论地域分布
分析热门评论
分析评论情绪
```

先下载评论数据，再进行分析。

---

## Step 5：补充需求确认

如果用户没有明确指定以下选项，主动告知可选功能。

### 评论图片下载

支持下载评论中的：

* 图片
* 表情包
* 梗图

询问：

```text
是否同时下载评论中的图片？
```

---

### 评论地域分布地图

支持统计：

* 评论用户地域
* 地域分布地图

询问：

```text
是否生成评论地域分布地图？
```

---

## Step 6：执行任务

根据需求选择对应命令。

---

## Step 7：汇总结果

完成后返回：

* 视频标题
* 评论数量
* 保存位置
* 图片下载状态
* 地图生成状态

---


# 安装流程
## 下载
访问`https://codeberg.org/wsh233/blblcd/releases#latest`,根据系统类型下载预编译好的最新的二进制文件（二进制文件名称示例”blblcd_0.2.1_darwin_amd64.zip“）。

## 安装
解压二进制文件压缩包，存放在合适的目录，如”/usr/local/bin“或系统 PATH 目录。

## 验证
执行
```bash
blblcd version
```


# 快速参考

通过执行命令 `blblcd version` 验证工具是否正常安装。

```bash
# 查看所有命令及帮助描述
blblcd -h
```
## 常用命令
* version
* video
* up
* latest


# 示例
## 示例 1

用户：

```text
下载这个视频的评论

https://www.bilibili.com/video/BV1sULg65Evk
```

处理：

1. 提取 BV 号
2. 检查 Cookie
3. 检查工具
4. 执行

```bash
blblcd video BV1sULg65Evk
```

5. 返回结果

---
## 示例 2

用户：

```text
下载评论中的表情包
```

执行：

```bash
blblcd video BVxxxx --img-download
```

---

# 输出格式

执行成功后输出：

```text
✓ 视频标题：<title>

✓ 评论数量：<count>

✓ 图片下载：已启用/未启用

✓ 地图生成：已启用/未启用

✓ 输出目录：

<output_path>
```

如果执行失败：

```text
✗ 失败原因：

<error>

建议：

<solution>
```