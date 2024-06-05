package store

import (
	"blblcd/assets"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/paulmach/orb/geojson"
)

func WriteGeoJSON(statMap map[string]int, filename string, output string) (ok bool) {
	// 读取GeoJSON文件
	data, err := assets.Assets.ReadFile("China_3857.geojson")
	geojsonOutput := fmt.Sprintf("%s/data_%s.geojson", output, filename)
	if err != nil {
		slog.Error("读取文件错误", err)
		return
	}

	// 解析GeoJSON数据
	fc := geojson.NewFeatureCollection()
	err = json.Unmarshal(data, &fc)
	if err != nil {
		slog.Error("解析GeoJSON错误:", err)
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
		slog.Error("转换GeoJSON错误:", err)
		return
	}

	err = os.WriteFile(geojsonOutput, outputData, 0644)
	if err != nil {
		slog.Error("写入geojson错误:", err)
		return
	}

	slog.Info(fmt.Sprintf("-----写入geojson：%s成功-----", geojsonOutput))
	ok = true
	return
}
