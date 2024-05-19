package process

import (
	"bufio"
	"fmt"
	"github.com/harshit22394/2million/defaults"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var domain = "www.apple.com"

// ProcessingPathsWithErrors process the paths but also uses a dedicated channel to handle errors
// Creates results.txt  and errors.txt
func ProcessingPathsWithErrors(fileName string) error {

	noOfGoRoutines := runtime.GOMAXPROCS(runtime.NumCPU())
	inputChannel := make(chan string, len(defaults.CommonPaths))
	outputChannel := make(chan string)
	errorsChannel := make(chan string)

	var wg sync.WaitGroup
	wg.Add(noOfGoRoutines)
	for range noOfGoRoutines {
		go processWithErrorChannel(inputChannel, &wg, outputChannel, errorsChannel)
	}

	var anotherWaitGroup sync.WaitGroup
	anotherWaitGroup.Add(2)
	err := results(&anotherWaitGroup, outputChannel)
	if err != nil {
		return err
	}

	err = errors(&anotherWaitGroup, errorsChannel)
	if err != nil {
		return err
	}

	err = input(fileName, inputChannel)
	if err != nil {
		return err
	}

	close(inputChannel)
	wg.Wait()
	close(outputChannel)
	close(errorsChannel)
	anotherWaitGroup.Wait()

	return nil
}

// ProcessingPaths process the paths but does not use a dedicated channel to handle errors
// Creates results.txt
func ProcessingPaths(fileName string) error {

	noOfGoRoutines := runtime.GOMAXPROCS(runtime.NumCPU())
	inputChannel := make(chan string, len(defaults.CommonPaths))
	outputChannel := make(chan string)
	var wg sync.WaitGroup
	wg.Add(noOfGoRoutines)

	log.Printf("Using %d goroutines\n", noOfGoRoutines)
	for range noOfGoRoutines {
		go process(inputChannel, &wg, outputChannel)
	}

	var anotherWaitGroup sync.WaitGroup
	anotherWaitGroup.Add(1)
	err := results(&anotherWaitGroup, outputChannel)
	if err != nil {
		return err
	}

	err = input(fileName, inputChannel)
	if err != nil {
		return err
	}

	close(inputChannel)
	wg.Wait()
	close(outputChannel)
	anotherWaitGroup.Wait()

	return nil
}

func processWithErrorChannel(inputChannel chan string, wg *sync.WaitGroup, outputChannel chan string, errorsChannel chan string) {
	defer wg.Done()
	for path := range inputChannel {
		startTime := time.Now()
		resp, err := http.Get(path)
		if err != nil {
			errorsChannel <- fmt.Sprintf("HTTP GET FAILURE %s %v\n", path, err)
			continue
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				errorsChannel <- fmt.Sprintf("READ BODY FAILURE %s %v %s\n", path, err, time.Since(startTime))
				continue
			}
			if strings.Contains(string(body), "iPhone") {
				outputChannel <- fmt.Sprintf("iPhone Success %s %s %s\n", resp.Status, path, time.Since(startTime))
				continue
			}
			outputChannel <- fmt.Sprintf("Success %s %s %s\n", resp.Status, path, time.Since(startTime))
			continue
		}
		outputChannel <- fmt.Sprintf("Failure %s %s %s\n", resp.Status, path, time.Since(startTime))
	}
}

func process(inputChannel chan string, wg *sync.WaitGroup, outputChannel chan string) {
	defer wg.Done()
	for path := range inputChannel {
		startTime := time.Now()
		resp, err := http.Get(path)
		if err != nil {
			outputChannel <- fmt.Sprintf("HTTP GET FAILURE %s %v %s\n", path, err, time.Since(startTime))
			continue
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				outputChannel <- fmt.Sprintf("READ BODY FAILURE %s %v %s\n", path, err, time.Since(startTime))
				continue
			}
			if strings.Contains(string(body), "iPhone") {
				outputChannel <- fmt.Sprintf("iPhone Success %s %s %s\n", resp.Status, path, time.Since(startTime))
				continue
			}
			outputChannel <- fmt.Sprintf("Success %s %s %s\n", resp.Status, path, time.Since(startTime))
			continue
		}
		outputChannel <- fmt.Sprintf("Failure %s %s %s\n", resp.Status, path, time.Since(startTime))
	}
}

func input(fileName string, inputChannel chan string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// reading paths from the given file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := scanner.Text()
		inputChannel <- fmt.Sprintf("https://%s/%s", domain, path)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func errors(anotherWaitGroup *sync.WaitGroup, errorsChannel chan string) error {
	errorsFile, err := os.Create("errors.txt")
	if err != nil {
		return err
	}
	errorWriter := bufio.NewWriter(errorsFile)
	go func() {
		defer func() {
			anotherWaitGroup.Done()
			errorWriter.Flush()
			errorsFile.Close()
		}()
		for error := range errorsChannel {
			errorWriter.Write([]byte(error))
		}
	}()
	return nil
}

func results(anotherWaitGroup *sync.WaitGroup, outputChannel chan string) error {
	outputFile, err := os.Create("results.txt")
	if err != nil {
		return err
	}
	outputWriter := bufio.NewWriter(outputFile)
	go func() {
		defer func() {
			anotherWaitGroup.Done()
			outputWriter.Flush()
			outputFile.Close()
		}()
		for output := range outputChannel {
			outputWriter.Write([]byte(output))
		}
	}()
	return err
}
