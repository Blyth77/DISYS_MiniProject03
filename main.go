package main

import (
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
) 

func main() {
	logger.ClearLog()
	logger.LogFileInit("main", 0)
}

