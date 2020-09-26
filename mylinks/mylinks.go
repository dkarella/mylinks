package mylinks

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// T is the struct responsible for getting and creating links
type T struct {
	file  *os.File
	links map[string]string
}

// Load initialize the mylinks.T with the file with the given file name
func (t *T) Load(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}

	t.file = file
	t.links = make(map[string]string)

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ",")

		if len(tokens) != 2 {
			file.Close()
			return fmt.Errorf("invalid line: %s", line)
		}

		t.links[tokens[0]] = tokens[1]
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return err
	}

	return nil
}

// Get returns the value for the given key if it exists
func (t *T) Get(key string) (string, bool) {
	v, ok := t.links[key]
	return v, ok
}

// Set will create a new link and persist it in the file
func (t *T) Set(key, value string) error {
	if _, err := t.file.WriteString(fmt.Sprintf("\n%s,%s", key, value)); err != nil {
		return err
	}
	t.links[key] = value
	return nil
}

// Close will cleanup resources being used
func (t *T) Close() {
	t.file.Close()
}
