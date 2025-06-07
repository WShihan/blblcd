# 开发测试用例
# 爬取单个评论
go run main.go video BV1KoTyzBErV --cookie=./cookie.text  --output=./tests/one_video
# 输出地图
go run main.go video BV1KoTyzBErV --cookie=./cookie.text --mapping  --output=./tests/one_video_mapping
# 下载图片
go run main.go video BV1KoTyzBErV --cookie=./cookie.text --img-download  --output=./tests/one_video_download_pic
# 多个视频
go run main.go video BV1KoTyzBErV BV17VTpztExx --cookie=./cookie.text --output=./tests/multi_video


go run main.go up 492366270 --cookie=./cookie.text --pages=1 --output=./tests/up
go run main.go up 492366270 --cookie=./cookie.text --pages=1 --output=./tests/up_mapping --mapping