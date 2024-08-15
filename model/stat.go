package model

// 统计类型
type Stat struct {
	Name     string
	Location int
	Sex      map[string]int
	Level    [7]int // 登记 1-7
	Like     int
	// Following int
}
