package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type MenuItem struct {
	Label  string
	Action func()
}

func RenderMenu(title string, items []MenuItem) {
	reader := bufio.NewReader(os.Stdin)
	for {
		Heading(title)
		
		fmt.Printf("\n  %s%s%s\n", Dim, "Select an action:", Reset)
		for i, item := range items {
			fmt.Printf("    %s%d.%s %s\n", Cyan, i+1, Reset, item.Label)
		}
		fmt.Printf("    %s0.%s %sExit%s\n", Red, Reset, Dim, Reset)
		
		fmt.Printf("\n%s %s➜%s ", Cyan, Bold+"GCAM", Reset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "0" || strings.ToLower(input) == "q" || strings.ToLower(input) == "exit" {
			fmt.Printf("\n%s  Goodbye!%s\n\n", Italic+Dim, Reset)
			return
		}

		var choice int
		fmt.Sscanf(input, "%d", &choice)

		if choice > 0 && choice <= len(items) {
			items[choice-1].Action()
			fmt.Printf("\n%sPress Enter to continue...%s", Dim, Reset)
			reader.ReadString('\n')
		} else if input != "" {
			Warn("Invalid choice: %s", input)
		}
	}
}
