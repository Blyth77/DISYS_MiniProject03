package main

import (
	"os"
	"runtime"
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
) 

func main() {
	defer printStack() // The defer makes the program wait for the method to return before terminating TGPL-book p. 143

	logger.ClearLog()
	logger.LogFileInit("main")
}


// Function for printing the stack when a goroutine panics TGPL-book p. 151
func printStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	os.Stdout.Write(buf[:n])
}