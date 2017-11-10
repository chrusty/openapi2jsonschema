package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func deriveSpecPathFileName(specPath string) string {
	_, sourceFileName := filepath.Split(specPath)
	return strings.TrimSuffix(sourceFileName, filepath.Ext(sourceFileName))
}

func generateFileName(outputFileNameWithoutExtention string) string {
	return fmt.Sprintf("%s/%s.%s", outPath, outputFileNameWithoutExtention, JSONSCHEMA_FILE_EXTENTION)
}

func writeToFile(fileName string, fileData []byte) error {

	// Open output file:
	outputFile, err := os.Create(fileName)
	if err != nil {
		logWithLevel(LOG_FATAL, "Can't open output file (%v): %v", fileName, err)
		return err
	}
	defer outputFile.Close()

	// Write to the file:
	if _, err := outputFile.Write(fileData); err != nil {
		logWithLevel(LOG_FATAL, "Can't write to file (%v): %v", fileName, err)
		return err
	}

	return nil
}
