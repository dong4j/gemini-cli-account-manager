package ui

import (
	"fmt"
	"strings"
)

const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Italic  = "\033[3m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlue = "\033[44m"
)

func Heading(text string) {
	width := 50
	line := strings.Repeat("━", width)
	fmt.Printf("\n%s%s%s\n", Blue, line, Reset)
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	fmt.Printf("%s%s%s%s%s\n", Blue, strings.Repeat(" ", padding), Bold+White, strings.ToUpper(text), Reset)
	fmt.Printf("%s%s%s\n", Blue, line, Reset)
}

func Box(title string, lines []string) {
	width := 60
	fmt.Printf("\n%s┏━ %s %s%s\n", Blue, Bold+White+title+Reset+Blue, strings.Repeat("━", width-len(title)-4), Reset)
	for _, l := range lines {
		fmt.Printf("%s┃%s %s\n", Blue, Reset, l)
	}
	fmt.Printf("%s┗%s%s\n", Blue, strings.Repeat("━", width-1), Reset)
}

func Success(format string, a ...interface{}) {
	fmt.Printf("%s[OK] %s%s\n", Green, fmt.Sprintf(format, a...), Reset)
}

func Error(format string, a ...interface{}) {
	fmt.Printf("%s[Error] %s%s\n", Red, fmt.Sprintf(format, a...), Reset)
}

func Info(format string, a ...interface{}) {
	fmt.Printf("%s[Info] %s%s\n", Cyan, fmt.Sprintf(format, a...), Reset)
}

func Warn(format string, a ...interface{}) {
	fmt.Printf("%s[Warning] %s%s\n", Yellow, fmt.Sprintf(format, a...), Reset)
}
