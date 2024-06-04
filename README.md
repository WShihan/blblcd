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

登录[哔哩哔哩 ](https://www.bilibili.com/)，打开开发者工具，点击其中任意一条请求，查看请求头，将cookie值复制后，在本地保存为text文件。

![image-20240603170934244](https://md-1301600412.cos.ap-nanjing.myqcloud.com/pic/typora/image-20240603170934244.png)

#### 获取up主mid

进入up主页，浏览器地址栏上将会显示mid

![image-20240603171216624](https://md-1301600412.cos.ap-nanjing.myqcloud.com/pic/typora/image-20240603171216624.png)



## 使用命令

在终端输入-help查看所有指令

```bash
blblcd -help
```




声明
====

* 源代码仅供交流学习使用，切勿用于违法犯罪。
