package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Get the source and destination file paths from command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: \t\trime-value-sync <source_file> <word_frequency_optimization_file> <file_overwrite_flag>")
		fmt.Println("Example: \trime-value-sync file1.yaml file2.yaml 1")
		return
	}

	file1path := os.Args[1]
	file2path := os.Args[2]
	overWrite := os.Args[3]

	// 打开文件并创建缓冲区
	file1, err := os.Open(file1path)
	if err != nil {
		fmt.Println("Error opening file1:", err)
		return
	}
	defer file1.Close()

	file2, err := os.Open(file2path)
	if err != nil {
		fmt.Println("Error opening file2:", err)
		return
	}
	defer file2.Close()

	scanner1 := bufio.NewScanner(file1)
	scanner2 := bufio.NewScanner(file2)

	// 创建一个map来存储文件2中的第一列和第三列
	file2Map := make(map[string]string)

	// 将文件2中的第一列和第三列映射到map中
	scanner2.Split(bufio.ScanLines)
	for scanner2.Scan() {
		line2 := scanner2.Text()
		fields2 := splitLine(line2)
		if len(fields2) >= 3 && len(fields2[2]) > 0 {
			file2Map[fields2[0]] = fields2[2]
		}
	}

	// 检查文件2扫描错误
	if err := scanner2.Err(); err != nil {
		fmt.Println("Error scanning file2:", err)
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("/tmp", "update_result.txt")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}

	// 处理文件1的每一行
	scanner1.Split(bufio.ScanLines)
	for scanner1.Scan() {
		line1 := scanner1.Text()
		fields1 := splitLine(line1)

		fields2, ok := file2Map[fields1[0]]

		// 如果在文件2中找到匹配项，则更新文件1
		if ok && fields1[2] != fields2 {
			fields1[2] = fields2

			updatedLine1 := joinLine(fields1)
			// 将更新后的行写入临时文件
			fmt.Fprintf(tempFile, "%s\n", updatedLine1)
			// Flush the file buffer to ensure immediate writing
			err := tempFile.Sync()
			if err != nil {
				fmt.Println("Error syncing temporary file:", err)
				return
			}
		} else {
			// 将原始行写入临时文件
			fmt.Fprintf(tempFile, "%s\n", line1)
			// Flush the file buffer to ensure immediate writing
			err := tempFile.Sync()
			if err != nil {
				fmt.Println("Error syncing temporary file:", err)
				return
			}
		}
	}

	// 检查文件1扫描错误
	if err := scanner1.Err(); err != nil {
		fmt.Println("Error scanning file1:", err)
		return
	}
	tempFile.Close()

	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}
	defer tempFile.Close()

	if overWrite == "1" {
		file1.Close()
		saveFile, _ := os.Create(file1path)
		io.Copy(saveFile, tempFile)
	} else {
		saveFile, _ := os.Create(filepath.Dir(file1path) + "/" + filepath.Base(tempFile.Name()))
		io.Copy(saveFile, tempFile)
	}
	//defer os.Remove(tempFile.Name())
	fmt.Println("文件匹配并更新完成")
}

// 将一行文本拆分为数组
// 将一行字符串以 \t 为分隔符，分成三个字符串
func splitLine(line string) []string {
	// 创建一个切片来存储分割后的字符串
	fields := make([]string, 3)

	// 使用切片索引来跟踪当前字段
	index := 0

	// 使用 bufio.Scanner 扫描字符串
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		// 跳过制表符
		if scanner.Text() == "\t" {
			if index < 2 {
				index++ // 移动到下一个字段
			}
			continue
		}
		// 将文本添加到当前字段
		fields[index] += scanner.Text()
	}

	// 检查扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning line:", err)
		return nil
	}

	// 返回分割后的字符串
	return fields
}

// 将数组拼接为一行文本
func joinLine(fields []string) string {
	line := ""
	for i, field := range fields {
		if i == len(fields)-1 {
			line += field
		} else {
			line += field + "\t"
		}
	}
	return line
}
