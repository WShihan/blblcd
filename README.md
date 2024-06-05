Blblcd
====

Blblcd（Bilibili-comment-dowloader）——基于[bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect)的Bilibili视频评论下载工具。

* 下载单个视频评论，可按热评/时间顺序
* 按投稿时间/收藏/播放顺序下载某up主投稿的多个视频评论



下载评论信息如下：

```
	Uname         姓名
	Sex           性别
	Content       评论内容
	Rpid          评论id
	Oid           评论区id
	Bvid          视频bv
	Mid           发送者id
	Parent        父级评论
	Fansgrade     是否粉丝标签
	Ctime         评论时间戳
	Like          喜欢数
	Following     是否关注
	Current_level 当前等级
	Location      位置
```

下载后单个视频的所有评论保存为一份csv。

使用
====

## 基础信息

* cookie ：必须
* mid ：up主代码，当查找up视频时必须
* bvid：稿件视频id，下载单个视频评论时必须



#### 获取cookie

登录[哔哩哔哩 ](https://www.bilibili.com/)，按住`F12`进入开发者工具页面，选择`网络`，点击其中任意一条请求，查看请求头，将cookie值复制后，在本地保存为text文件（最好是与blblcd放置于同一个目录下，命名为cookie.text）。

![image-20240603170934244](https://md-1301600412.cos.ap-nanjing.myqcloud.com/pic/typora/image-20240603170934244.png)

#### 获取up主mid

进入up主页，浏览器地址栏上将会显示mid，复制它，例如下面链接里的mid为`112233445`。

![image-20240605170502785](https://md-1301600412.cos.ap-nanjing.myqcloud.com/pic/typora/image-20240605170502785.png)

## 使用命令

在终端输入-help查看所有指令

```bash
./blblcd -help
```

![image-20240605164108092](https://md-1301600412.cos.ap-nanjing.myqcloud.com/pic/typora/image-20240605164108092.png)

### 示例

#### 单一视频

基础

```bash
blblcd -bvid BV1VJ4m1jk34K
```

指定评论按`回复`顺序

```bash
blblcd -bvid BV1VJ4m1jk34K -corder 2
```

指定`cookie`文件

```bash
blblcd -cookie /path/to/cookiefile.text -bvid BV1VJ4m1jk34K -corder 2
```

输出位置

```bash
blblcd -bvid BV1VJ4m1jk34K -corder 2 -output path/to/output
```



#### up视频列表

基础（默认获取前三页，一页30条视频）

```bash
blblcd -mid 123344555
```

指定`cookie`

```bash
blblcd -mid 123344555 -cookie /path/to/cookiefile.text
```

视频列表顺序，按`最多收藏`

```bash
blblcd -mid 123344555 -skip 3 -pages 5 -vorder stow
```

固定页数

```bash
blblcd -mid 123344555 -pages 5
```

忽略页数，跳过前三页后获取5页，即4-8页

```bash
blblcd -mid 123344555 -skip 3 -pages 5
```



输出位置

```bash
blblcd -mid 123344555  -output output/path
```



并发数量

```bash
blblcd -mid 123344555  -goroutines 10
```







声明
====

* 源代码仅供交流学习使用，切勿用于违法犯罪。
* blblcd不会保存或泄露cookie，请放心食用。
