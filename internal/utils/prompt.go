package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Prompt(message string) (string, error) {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func PromptWithDefault(message, defaultValue string) (string, error) {
	promptMsg := fmt.Sprintf("%s [%s]: ", message, defaultValue)
	input, err := Prompt(promptMsg)
	if err != nil {
		return "", err
	}
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func PromptYesNo(message string) (bool, error) {
	input, err := Prompt(message + " (y/n): ")
	if err != nil {
		return false, err
	}
	input = strings.ToLower(input)
	return input == "y" || input == "yes", nil
}

func PromptSelect(message string, options []string) (string, error) {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	input, err := Prompt("Enter selection: ")
	if err != nil {
		return "", err
	}

	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(options) {
		return "", fmt.Errorf("invalid selection")
	}

	return options[selection-1], nil
}
