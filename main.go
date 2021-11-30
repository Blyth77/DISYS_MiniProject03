package main

import (
	"fmt"
	"os"
	"strconv"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	replica "github.com/Blyth77/DISYS_MiniProject03/replicamanager"
)

func main() {
	logger.ClearLog("log")
	logger.LogFileInit("main")


	numberOfReplicas, _ := strconv.Atoi(os.Args[1])
	makePortListForFrontEnd(numberOfReplicas)

	for i := 1; i <= numberOfReplicas; i++ { 
		number := int32(i)
		go replica.Start(number, (3000 + number))
	}

	bl := make(chan bool)
	<-bl
}

func makePortListForFrontEnd(numberOfReplicas int) {
	
	logger.ClearLog("replicamanager/portlist")
	logger.MakeLogFolder("replicamanager/portlist")

	f, err := os.OpenFile("replicamanager/portlist/listOfReplicaPorts.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//
	}
	
	defer f.Close()
	for i := 1; i <= numberOfReplicas; i++ {
		if _, err := f.WriteString(fmt.Sprintf("%v\n", 6000+i)); err != nil {
			//
		}
	}
}
