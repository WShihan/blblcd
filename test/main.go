package main

import (
	"blblcd/assets"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Data struct {
	Title string        `json:"title"`
	China template.HTML `json:"china"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		geojson, err := assets.Assets.ReadFile("China_3857.geojson")
		if err != nil {
			fmt.Println(err.Error())
		}
		// 创建数据
		data := Data{Title: "Bvdfdf", China: template.HTML(string(geojson))}

		tmpl, err := template.ParseFS(assets.Assets, "template.html")
		if err != nil {
			panic(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatalf("Error executing template: %v", err)
		}

		//  写入文件
		file, err := os.Create("output.html")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		err = tmpl.Execute(file, data)
		if err != nil {
			panic(err)
		}
	})

	log.Println("Listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
