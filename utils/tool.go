package utils

import (
	"blblcd/model"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
)

func FileOrPathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func ExecutePath() string {
	executePath, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return filepath.Dir(executePath)
}

func ReadTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func PresetPath(path string) {
	if !FileOrPathExists(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			slog.Error("failed to create directory", "path", path, "err", err)
		}
	}
}

func EncodePaginationOffset(pagination model.PaginationOffset) string {
	paginationJSON, err := json.Marshal(pagination)
	if err != nil {
		fmt.Println("Error marshaling pagination:", err)
		return ""
	}

	return string(paginationJSON)
}

func DecodePaginationOffset(paginationStr string) (*model.PaginationOffset, error) {
	var paginationOffset model.PaginationOffset
	err := json.Unmarshal([]byte(paginationStr), &paginationOffset)
	if err != nil {
		fmt.Println("Error unmarshaling inner JSON:", err)
		return nil, err
	}
	return &paginationOffset, nil
}

func ProgressBar(current int64, total int64) {
	var percent float64
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	} else {
		percent = 0
	}

	percent = math.Min(math.Max(percent, 0), 100)

	barLength := 100
	filled := int(float64(barLength) * percent / 100)
	var bar string
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}

	fmt.Printf("\r[%s]进度: %.0f%%\n", bar, percent)
}

func PrintLogo() {
	logo := `
                         +++++                 +++++            
                         +====++             ++====++           
                         +======++          +======+            
                           +======++      +======+              
                    +++++++++=======++++++======+++++++++       
                  ++====================================++      
                  +======================================+     
                  ========================================     
                  ========================================     
                  ========================================     
                  ========*==========+==========*=========     
                  ======*   *======+   +======*   *=======     
                  ======*   *======+   +======*   *=======     
                  ========*==========+==========*=========     
                  ========================================     
                  ========================================     
                  ========================================     
                  +======================================+     
                  ++====================================++      
                    +++++++++++++==========+++++++++++++        
                                 ++======++                     
                                   ++===+*                      
                                     +*+                        
                                                                                              
                 ==        ==  ==        ==                ==                                 
                 ==        ==  ==        ==                ==                                 
                 =======   ==  =======   ==   ======  =======                                 
                 ==   ===  ==  ===   ==  ==  ==      ===   ==                                 
                 ==    ==  ==  ==    ==  == ===      ==    ==                                 
                 ==   ===  ==  ===   ==  ==  ==      ===   ==                                 
                 =======   ==  =======   ==   ======  =======                                 
                                                                                                                      
                                                                                                              `
	fmt.Println(logo)
}
