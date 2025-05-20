package utils

import (
	"bufio"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"golang.org/x/term"
	"os"
	"regexp"
	"strings"
	"time"
)

func RequestConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", message)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	return resp == "y" || resp == "yes", nil
}

func RequestOutputConfirmation(message, output string) (bool, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y: confirm, n: abort, e: edit]: ", message)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return false, "", fmt.Errorf("failed to read user input: %w", err)
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	switch resp {
	case "y", "yes":
		return true, output, nil
	case "n", "no":
		return false, output, nil
	case "e", "edit":
		edited, err := Edit(output)
		if err != nil {
			return false, output, err
		}
		return true, edited, nil
	default:
		return false, output, fmt.Errorf("invalid response: %s", resp)
	}
}

func RunREPL(processInput func(string) (response any, err error)) {
	reader := bufio.NewReader(os.Stdin)
	var renderer *glamour.TermRenderer
	r, renderErr := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.DraculaStyle),
		glamour.WithWordWrap(TerminalWidth()),
	)
	if renderErr == nil {
		renderer = r
	}
	fmt.Println("Interactive project-aware REPL started. Type 'exit' or 'quit' to quit.")
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "exit" || input == "quit" {
			break
		}
		resp, err := processInput(input)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if renderer != nil {
			respStr, ok := resp.(string)
			if ok {
				if out, err := renderer.Render(formatResponse(respStr)); err != nil {
					fmt.Println(respStr)
				} else {
					fmt.Println(out)
					//fmt.Println(formatResponse(respStr))
				}
				continue
			}
		}
		fmt.Println(resp)
	}
}

func formatResponse(input string) string {
	re := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
	return re.ReplaceAllStringFunc(input, func(block string) string {
		matches := re.FindStringSubmatch(block)
		if len(matches) < 2 {
			return block
		}
		inner := strings.TrimSpace(matches[1])
		quoted := "### Thought Process:\n> " + strings.ReplaceAll(inner, "\n", "\n> ")
		quoted += "\n---\n"
		return quoted
	})
}

func RunWithSpinner(msg string, process func()) {
	start := time.Now()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("⏳ %s ", msg)
	s.Suffix = "\n"
	s.Start()
	process()
	s.Stop()
	elapsed := time.Since(start)
	fmt.Printf("✅ %s done in %.f minutes\n", msg, elapsed.Minutes())
}

// TerminalWidth returns the terminal's current width in characters.
func TerminalWidth() int {
	// GetSize returns width, height, error
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	// Fallback width
	return 80
}
