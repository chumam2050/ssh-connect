package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ask(reader *bufio.Reader, label string, current string) (string, error) {
	if current != "" {
		fmt.Printf("%s [%s]: ", label, current)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return current, nil
	}
	return trimmed, nil
}

func askInt(reader *bufio.Reader, label string, current int) (int, error) {
	val, err := ask(reader, label, strconv.Itoa(current))
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s", label)
	}
	return parsed, nil
}

func selectFromList(title string, values []string) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("%s is empty", title)
	}
	fmt.Println(title)
	for i, v := range values {
		fmt.Printf("  %d) %s\n", i+1, v)
	}
	fmt.Print("Choose number: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(values) {
		return "", fmt.Errorf("invalid selection")
	}
	return values[idx-1], nil
}
