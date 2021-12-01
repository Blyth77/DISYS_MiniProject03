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
	logger.InfoLogger.Println("Setting up database")

	numberOfReplicas, _ := strconv.Atoi(os.Args[1])
	makePortListForFrontEnd(numberOfReplicas)
	logger.InfoLogger.Printf("Setting up replicas. Number of replicas in database: %v\n", numberOfReplicas)

	for i := 1; i <= numberOfReplicas; i++ {
		number := int32(i)
		port := 6000 + number
		go replica.Start(number, port)
		logger.InfoLogger.Printf("Replicas number: %v Port: %v started.\n", number, port)
	}
	logger.InfoLogger.Println("Replicas setup complete.")
	logger.InfoLogger.Println("Database setup complete. Ready for clients!")

	bl := make(chan bool)
	<-bl
}

func makePortListForFrontEnd(numberOfReplicas int) {
	logger.InfoLogger.Println("Setting up listOfReplicaPorts.txt")
	logger.ClearLog("replicamanager/portlist")
	logger.MakeLogFolder("replicamanager/portlist")

	f, err := os.OpenFile("replicamanager/portlist/listOfReplicaPorts.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to create listOfReplicaPorts.txt. Error: %v", err)
	}
	defer f.Close()

	for i := 1; i <= numberOfReplicas; i++ {
		number := 6000 + i
		if _, err := f.WriteString(fmt.Sprintf("%v\n", number)); err != nil {
			logger.ErrorLogger.Fatalf("Failed to add %v listOfReplicaPorts.txt. Error: %v", number, err)
		}
	}
	logger.InfoLogger.Println("listOfReplicaPorts.txt setup complete.")
}
