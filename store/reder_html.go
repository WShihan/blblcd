package store

import (
	"blblcd/assets"
	"fmt"
	"html/template"
	"log/slog"
	"os"
)

type Data struct {
	Title string
	Name  string
	China template.HTML
}

func RenderHTML(geofile string, htmlname string) (ok bool, err error) {
	geojson, err := os.ReadFile(geofile)
	HtmlData := Data{
		Title: "Bvidfdodfd",
		Name:  "John",
		China: template.HTML(string(geojson)),
	}

	tmpl, err := template.ParseFS(assets.Assets, "template.html")
	if err != nil {
		panic(err)
	}

	//  写入文件
	file, err := os.Create(htmlname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, HtmlData)
	if err != nil {
		panic(err)
	}
	slog.Info(fmt.Sprintf("-------渲染html文件成功:%s--------", htmlname))

	return
}
