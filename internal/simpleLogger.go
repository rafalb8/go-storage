package internal

import "fmt"

type SimpleLogger struct{}

func (s *SimpleLogger) Debug(args ...any) {
	fmt.Println(args...)
}

func (s *SimpleLogger) Warn(args ...any) {
	fmt.Println(args...)
}

func (s *SimpleLogger) Info(args ...any) {
	fmt.Println(args...)
}

func (s *SimpleLogger) Error(args ...any) {
	fmt.Println(args...)
}

func (s *SimpleLogger) Fatal(args ...any) {
	panic(fmt.Sprintln(args...))
}
