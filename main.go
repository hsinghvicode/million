package main

import (
	"github.com/harshit22394/2million/process"
	"github.com/harshit22394/2million/records"
	"log"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	numberOfRecords := 1000
	log.Printf("Creating records with %d records\n", numberOfRecords)
	fileName, err := records.CreateFile(numberOfRecords)
	if err != nil {
		log.Printf("error while creating records\n%v", err)
		os.Exit(1)
	}
	log.Printf("Time Taken to create paths %v", time.Since(startTime))

	log.Println("Starting Paths Processing")
	startTime = time.Now()
	//err = process.ProcessingPathsWithErrors(fileName)
	err = process.ProcessingPaths(fileName)
	if err != nil {
		log.Printf("error while processing paths\n%v", err)
		os.Exit(1)
	}
	log.Printf("Time Taken to process paths %v", time.Since(startTime))

}
