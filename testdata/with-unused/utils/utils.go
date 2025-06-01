package utils

import "fmt"

// Used functions
func PrintMessage(msg string) {
	fmt.Println("Message:", msg)
}

type Logger struct {
	prefix string
}

func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

func (l *Logger) Log(message string) {
	fmt.Printf("[%s] %s\n", l.prefix, message)
}

// UNUSED - these should be eliminated
func FormatNumber(n int) string {
	return fmt.Sprintf("Number: %d", n)
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

type FileManager struct {
	basePath string
}

func (f *FileManager) CreateFile(name string) error {
	// mock implementation
	return nil
}

func (f *FileManager) DeleteFile(name string) error {
	// mock implementation
	return nil
}

const DefaultPath = "/tmp"

var GlobalCounter int = 0