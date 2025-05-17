package utils

import (
	"bufio"
	"fmt"
	"github.com/briandowns/spinner"
	"os"
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

func RunREPL(processInput func(string) (response string, err error)) {
	reader := bufio.NewReader(os.Stdin)
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
		fmt.Println(resp)
	}
}

func RunWithSpinner(msg string, process func()) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("⏳ %s ", msg)
	s.Suffix = "\n"
	s.Start()
	defer s.Stop()
	process()
}
