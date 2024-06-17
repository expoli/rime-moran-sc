package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SensitiveWordMap 创建一个数据结构来高效存储敏感词以便快速查找。哈希表或字典树都是合适的选择。
type SensitiveWordMap map[string]bool

type SensitiveWordResult struct {
	sensitiveWord string
	line          int
	lineText      string // Optional: Include the entire line for context
	err           error
}

// 逐行读取敏感词列表文件并填充敏感词数据结构。
func loadSensitiveWords(filePath string) (SensitiveWordMap, error) {
	sensitiveWords := make(SensitiveWordMap)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sensitiveWord := scanner.Text()
		sensitiveWords[sensitiveWord] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return sensitiveWords, nil
}

// 使用协程并发处理文本文件逐行检查敏感词并输出结果。
func concurrentProcessText(textFilePath string, sensitiveWords SensitiveWordMap, concurrency int, ch chan<- SensitiveWordResult) {
	file, err := os.Open(textFilePath)
	if err != nil {
		ch <- SensitiveWordResult{err: err}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		for sensitiveWord, _ := range sensitiveWords {
			if strings.Contains(line, sensitiveWord) {
				ch <- SensitiveWordResult{
					sensitiveWord: sensitiveWord,
					line:          lineNum,
					lineText:      line,
				}
			}
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		ch <- SensitiveWordResult{err: err}
	}

	close(ch)
}

func main() {
	sensitiveWords, err := loadSensitiveWords("色情词.txt")
	if err != nil {
		fmt.Println("Error loading sensitive words:", err)
		return
	}

	concurrency := 4 // Adjust concurrency as needed
	ch := make(chan SensitiveWordResult, concurrency)

	go concurrentProcessText("base.dict.yaml", sensitiveWords, concurrency, ch)

	for result := range ch {
		if result.err != nil {
			fmt.Println("Error processing text:", result.err)
			return
		}
		if result.sensitiveWord != "" {
			fmt.Printf("Sensitive word '%s' found on line %d: %s\n", result.sensitiveWord, result.line, result.lineText)
		}
	}
}
