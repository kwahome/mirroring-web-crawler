package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func writeToFile(data []byte, filePath string) error {
	fmt.Printf("writing to '%s'\n", filePath)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return err
	}

	defer closeFile(file)

	_, err = file.Write(data)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return err
	}

	return nil
}

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Fatalf("an error has occurred while closing file: %v\n", err)
	}
}
