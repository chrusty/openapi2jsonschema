package schemaconverter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// WriteJSONSchemasToFiles writes each JSONSchema to a file:
func (c *Converter) WriteJSONSchemasToFiles(generatedJSONSchemas []GeneratedJSONSchema) error {

	// Go through the JSONSchemas and write each one to a file:
	for _, generatedJSONSchema := range generatedJSONSchemas {

		// Generate a filename for the JSONSchema:
		jsonSchemaFileName := c.deriveJSONSchemaFilename(generatedJSONSchema.Name)

		// Write the schemaJSON out to a file:
		if err := c.writeToFile(jsonSchemaFileName, generatedJSONSchema.Bytes); err != nil {
			return err
		}

		c.logger.WithField("jsonschema_name", generatedJSONSchema.Name).WithField("filename", jsonSchemaFileName).Debug("Wrote schema-definition to a file")
	}

	return nil
}

// WriteGoConstantsToFile writes an importable go package containing constants for each JSONSchema:
func (c *Converter) WriteGoConstantsToFile(generatedJSONSchemas []GeneratedJSONSchema) error {

	goConstantsCode := []byte("package schema\n\n")

	// Prepare a filename:
	specFileName := c.deriveSpecPathFilename()
	goConstantsFilename := strings.Replace(fmt.Sprintf("%v/%v%v.go", c.config.OutPath, c.config.GoConstantsFilename, strings.Title(specFileName)), "-", "", 0)

	// Go through the JSONSchemas and write each one to a file:
	for _, generatedJSONSchema := range generatedJSONSchemas {
		definitionConstant := fmt.Sprintf("const Schema%s%s string = `%s`\n\n", strings.Title(specFileName), strings.Title(generatedJSONSchema.Name), generatedJSONSchema.Bytes)
		goConstantsCode = append(goConstantsCode, definitionConstant...)
	}

	// Write the schemaJSON out to a file:
	if err := c.writeToFile(goConstantsFilename, goConstantsCode); err != nil {
		return err
	}

	c.logger.WithField("go_constants_filename", goConstantsFilename).Debug("Wrote GoLang constants to a file")

	return nil
}

// deriveSpecPathFilename cleans up the name of the spec file:
func (c *Converter) deriveSpecPathFilename() string {
	_, sourceFileName := filepath.Split(c.config.SpecPath)
	return strings.TrimSuffix(sourceFileName, filepath.Ext(sourceFileName))
}

// deriveJSONSchemaFilename derives JSONSchema filenames:
func (c *Converter) deriveJSONSchemaFilename(outputFileNameWithoutExtention string) string {
	return fmt.Sprintf("%s/%s.%s", c.config.OutPath, outputFileNameWithoutExtention, c.config.JSONSchemaFileExtention)
}

// writeToFile handles writing files to disk:
func (c *Converter) writeToFile(fileName string, fileData []byte) error {

	// Open output file:
	outputFile, err := os.Create(fileName)
	if err != nil {
		return errors.Wrapf(err, "Can't open output file (%v)", fileName)
	}
	defer outputFile.Close()

	// Write to the file:
	if _, err := outputFile.Write(fileData); err != nil {
		return errors.Wrapf(err, "Can't write to file (%v)", fileName)
	}

	return nil
}
