package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func deriveSpecPathFileName() string {
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

func writeAllJSONSchemasToFile(generatedJSONSchemas map[string][]byte) error {

	// Go through the JSONSchemas and write each one to a file:
	for definitionName, definitionJSONSchemaJSON := range generatedJSONSchemas {

		// Generate a filename to store the JSONSchema in:
		jsonSchemaFileName := generateFileName(definitionName)

		// Write the schemaJSON out to a file:
		if err := writeToFile(jsonSchemaFileName, definitionJSONSchemaJSON); err != nil {
			return err
		}

		logWithLevel(LOG_DEBUG, "Wrote schema-definition (%s) to a file: %v", definitionName, jsonSchemaFileName)
	}

	return nil
}

func writeAllJSONSchemasToGoConstants(generatedJSONSchemas map[string][]byte) error {

	goConstantsCode := []byte("package schema\n\n")

	// Prepare a filename:
	specFileName := deriveSpecPathFileName()
	goConstantsFilename := fmt.Sprintf("%v/%v_%v.go", outPath, GO_CONSTANTS_FILENAME, specFileName)

	// Go through the JSONSchemas and write each one to a file:
	for definitionName, definitionJSONSchemaJSON := range generatedJSONSchemas {
		definitionConstant := fmt.Sprintf("const %s_%s string = `%s`\n\n", specFileName, definitionName, definitionJSONSchemaJSON)
		goConstantsCode = append(goConstantsCode, definitionConstant...)
	}

	// Write the schemaJSON out to a file:
	if err := writeToFile(goConstantsFilename, goConstantsCode); err != nil {
		return err
	}

	logWithLevel(LOG_DEBUG, "Wrote GoLang constants to a file: %v", goConstantsFilename)

	return nil
}
