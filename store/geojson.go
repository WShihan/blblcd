package store

import (
	"blblcd/model"
	"blblcd/utils"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/paulmach/orb/geojson"
)

func WriteGeoJSON(statMap map[string]model.Stat, filename string, output string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("写入GeoJSON错误:", err)
		}
	}()
	// 读取GeoJSON文件
	data, err := os.ReadFile(utils.ExcutePath() + "/geo-template.geojson")
	geojsonOutput := filepath.Join(output, filename+".geojson")
	if err != nil {
		slog.Error("读取文件错误" + err.Error())
		return
	}
	utils.PresetPath(output)
	// 解析GeoJSON数据
	fc := geojson.NewFeatureCollection()
	err = json.Unmarshal(data, &fc)
	if err != nil {
		slog.Error("解析GeoJSON错误:" + err.Error())
		return
	}

	// 遍历每个Feature并修改字段值
	for _, feat := range fc.Features {
		province := feat.Properties["name"]
		stat := statMap[province.(string)]
		feat.Properties["count"] = stat.Location
		feat.Properties["like"] = stat.Like
		feat.Properties["male"] = stat.Sex["男"]
		feat.Properties["female"] = stat.Sex["女"]
		feat.Properties["sexless"] = stat.Sex["保密"]
		feat.Properties["level0"] = stat.Level[0]
		feat.Properties["level1"] = stat.Level[1]
		feat.Properties["level2"] = stat.Level[2]
		feat.Properties["level3"] = stat.Level[3]
		feat.Properties["level4"] = stat.Level[4]
		feat.Properties["level5"] = stat.Level[5]
		feat.Properties["level6"] = stat.Level[6]
	}

	// 写入修改后的GeoJSON文件
	outputData, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		slog.Error("转换GeoJSON错误:" + err.Error())
		return
	}

	err = os.WriteFile(geojsonOutput, outputData, 0644)
	if err != nil {
		slog.Error("写入geojson错误:" + err.Error())
		return
	}
	slog.Info(fmt.Sprintf("-----写入geojson：%s成功-----", geojsonOutput))
	RenderHTML(filename, filepath.Join(output, filename+".geojson"), filepath.Join(output, filename+".html"))
}
