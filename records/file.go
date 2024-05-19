package records

import (
	"fmt"
	"github.com/harshit22394/2million/defaults"
	"log"
	"os"
	"runtime"
	"sync"
)

func CreateFile(numOfRecords int) (string, error) {
	noOfGoRoutines := runtime.GOMAXPROCS(runtime.NumCPU())
	commonPathsLength := len(defaults.CommonPaths)
	var wg sync.WaitGroup
	inputChannel := make(chan string, commonPathsLength)

	file, err := os.Create("paths.txt")
	if err != nil {
		return "", err
	}

	log.Printf("Using %d goroutines\n", noOfGoRoutines)
	for range noOfGoRoutines {
		wg.Add(1)
		go addPathToFile(inputChannel, &wg, file)
	}

	for i := 0; i < numOfRecords; i++ {
		inputChannel <- defaults.CommonPaths[i%commonPathsLength]
	}
	close(inputChannel)
	wg.Wait()

	err = file.Close()
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func addPathToFile(inputChannel chan string, wg *sync.WaitGroup, millionPathFile *os.File) {
	defer wg.Done()
	for path := range inputChannel {

		// adding a newline to end of each path
		path = fmt.Sprintf("%s\n", path)

		// not handling error here to keep this simple for now
		_, _ = millionPathFile.Write([]byte(path))

	}
}
