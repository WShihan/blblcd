go test -coverprofile=coverage.out ./...

# 输出 HTML 文件
go tool cover -html=coverage.out -o coverage.html 