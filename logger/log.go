package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)


// This is a variadic func, the "..." means sthe int32 is optional TGPL-book p. 142
func LogFileInit(typeOf string, id ...int32) {
    if(typeOf == "main") {
        ClearLog("log")

        MakeLogFolder("log")
        
        file, err := os.Create("log/log.txt")
        if err != nil {
            log.Fatal(err)
        }

        initLoggerMessages(file)
    } else {
        folder := fmt.Sprintf("log/%vlog", typeOf)

        if !checkIfExist(folder) {
            MakeLogFolder(folder)
        }

        file, err := os.Create(fmt.Sprintf("log/%vlog/%v%d.txt", typeOf, typeOf, id))
        if err != nil {
            log.Fatal(err)
        }
        initLoggerMessages(file)
    }
}

func initLoggerMessages(file *os.File) {
    InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func ClearLog(name string) {
    if checkIfExist(name) {
        os.RemoveAll(name)
    }        
}

func checkIfExist(name string) bool {
    _, err := os.Stat(name); 
    return !os.IsNotExist(err)
}

func MakeLogFolder(name string) {
    err := os.Mkdir(fmt.Sprintf("%v", name), os.ModePerm)
    if err != nil {
        log.Fatal(err)
    }
}
