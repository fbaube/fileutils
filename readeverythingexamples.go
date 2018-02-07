package fileutils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func ExampleReadAll() {
	file, err := os.Open("input.txt")
	if err != nil {
		log.Panicf("failed reading file: %s", err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	fmt.Printf("\nLength: %d bytes", len(data))
	fmt.Printf("\nData: %s", data)
	fmt.Printf("\nError: %v", err)
}

func ExampleReadDir() {
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		log.Panicf("failed reading directory: %s", err)
	}
	fmt.Printf("\nNumber of files in current directory: %d", len(entries))
	fmt.Printf("\nError: %v", err)
}

func ExampleReadFile() {
	data, err := ioutil.ReadFile("input.txt")
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	fmt.Printf("\nLength: %d bytes", len(data))
	fmt.Printf("\nData: %s", data)
	fmt.Printf("\nError: %v", err)
}

/*
func main() {
	fmt.Printf("########Demo of ReadAll function#########\n")
	exampleReadAll()

	fmt.Printf("\n\n########Demo of ReadDir function#########\n")
	exampleReadDir()

	fmt.Printf("\n\n########Demo of ReadFile function#########\n")
	exampleReadFile()
}
*/
