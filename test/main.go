package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/paulmach/orb/geojson"
)

func WriteGeoJSON(filePath string, statMap map[string]int) {
	// 读取GeoJSON文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("读取文件错误", err)
		return
	}

	// 解析GeoJSON数据
	fc := geojson.NewFeatureCollection()
	err = json.Unmarshal(data, &fc)
	if err != nil {
		fmt.Println("解析GeoJSON错误:", err)
		return
	}

	// 遍历每个Feature并修改字段值
	for _, feat := range fc.Features {
		province := feat.Properties["name"]
		value := statMap[province.(string)]
		feat.Properties["count"] = value
	}

	// 写入修改后的GeoJSON文件
	outputData, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		fmt.Println("转换GeoJSON错误:", err)
		return
	}

	outputFilePath := "output.geojson"
	err = os.WriteFile(outputFilePath, outputData, 0644)
	if err != nil {
		fmt.Println("写入geojson错误:", err)
		return
	}

	fmt.Printf("-----写入geojson：%s成功----- output.geojson", outputFilePath)
}

func main() {
	statMap := map[string]int{
		"广东": 100,
		"北京": 10,
	}
	WriteGeoJSON("../assets/China_3857.geojson", statMap)
}
