package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Writer handles writing JSONSchemas and Go constants to files:
type Writer struct {
	config *types.Config
	logger *logrus.Logger
}

// New takes a config and returns a new Writer:
func New(config *types.Config, logger *logrus.Logger) *Writer {
	return &Writer{
		config: config,
		logger: logger,
	}
}

// WriteJSONSchemasToFiles writes each JSONSchema to a file:
func (w *Writer) WriteJSONSchemasToFiles(generatedJSONSchemas []types.GeneratedJSONSchema) error {

	// Go through the JSONSchemas and write each one to a file:
	for _, generatedJSONSchema := range generatedJSONSchemas {

		// Generate a filename for the JSONSchema:
		jsonSchemaFileName := w.deriveJSONSchemaFilename(generatedJSONSchema.Name)

		// Write the schemaJSON out to a file:
		if err := w.writeToFile(jsonSchemaFileName, generatedJSONSchema.Bytes); err != nil {
			return err
		}

		w.logger.WithField("jsonschema_name", generatedJSONSchema.Name).WithField("filename", jsonSchemaFileName).Debug("Wrote schema-definition to a file")
	}

	return nil
}

// WriteGoConstantsToFile writes an importable go package containing constants for each JSONSchema:
func (w *Writer) WriteGoConstantsToFile(generatedJSONSchemas []types.GeneratedJSONSchema) error {

	goConstantsCode := []byte("package schema\n\n")

	// Prepare a filename:
	specFileName := w.deriveSpecPathFilename()
	goConstantsFilename := w.deriveGoConstantsFilename(specFileName)

	// Go through the JSONSchemas and write each one to a file:
	for _, generatedJSONSchema := range generatedJSONSchemas {
		definitionConstant := fmt.Sprintf("const Schema%s%s string = `%s`\n\n", strings.ReplaceAll(strings.Title(specFileName), "-", ""), strings.ReplaceAll(strings.Title(generatedJSONSchema.Name), "-", ""), generatedJSONSchema.Bytes)
		goConstantsCode = append(goConstantsCode, definitionConstant...)
	}

	// Write the schemaJSON out to a file:
	if err := w.writeToFile(goConstantsFilename, goConstantsCode); err != nil {
		return err
	}

	w.logger.WithField("go_constants_filename", goConstantsFilename).Debug("Wrote GoLang constants to a file")

	return nil
}

// deriveGoConstantsFilename derives the go-constants filename:
func (w *Writer) deriveGoConstantsFilename(specFileName string) string {
	return strings.Replace(fmt.Sprintf("%v/%v%v.go", w.config.OutPath, w.config.GoConstantsFilename, strings.Title(specFileName)), "-", "", 0)
}

// deriveJSONSchemaFilename derives JSONSchema filenames:
func (w *Writer) deriveJSONSchemaFilename(outputFileNameWithoutExtention string) string {
	return fmt.Sprintf("%s/%s.%s", w.config.OutPath, outputFileNameWithoutExtention, w.config.JSONSchemaFileExtention)
}

// deriveSpecPathFilename cleans up the name of the spec file:
func (w *Writer) deriveSpecPathFilename() string {
	_, sourceFileName := filepath.Split(w.config.SpecPath)
	return strings.TrimSuffix(sourceFileName, filepath.Ext(sourceFileName))
}

// writeToFile handles writing files to disk:
func (w *Writer) writeToFile(fileName string, fileData []byte) error {

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
