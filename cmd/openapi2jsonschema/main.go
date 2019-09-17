package main

import (
	"flag"

	"github.com/chrusty/openapi2jsonschema/pkg/schemaconverter"
	"github.com/chrusty/openapi2jsonschema/pkg/schemaconverter/types"

	"github.com/sirupsen/logrus"
)

var (
	config = &types.Config{
		GoConstantsFilename:     "jsonschemas",
		JSONSchemaFileExtention: "jsonschema",
	}
	logLevel string
)

func init() {
	flag.BoolVar(&config.AllowNullValues, "allow_null_values", false, "Allow NULL values as well as the defined types?")
	flag.BoolVar(&config.BlockAdditionalProperties, "block_additional_properties", false, "Block additional properties?")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level [trace, debug, info, warn, error]")
	flag.BoolVar(&config.GoConstants, "go_constants", false, "Output GoLang constants (in addition to JSONSchemas)?")
	flag.StringVar(&config.OutPath, "out", "./out", "Where to write jsonschema output files to")
	flag.StringVar(&config.SpecPath, "spec", "spec.yaml", "Location of the swagger spec file")
	flag.BoolVar(&config.V3, "v3", false, "Use OpenAPI3 (instead of Swagger 2)?")
	flag.Parse()
}

func main() {

	// Prepare a new logger:
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	// Parse the log-level:
	parsedLogLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Unable to parse loglevel")
	}
	logger.SetLevel(parsedLogLevel)

	// Prepare a new schema converter and writer:
	schemaConverter, schemaWriter, err := schemaconverter.New(config, logger)
	if err != nil {
		logger.WithError(err).Fatal("Unable to prepare a schema converter")
	}

	// Generate JSONSchemas:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	if err != nil {
		logger.WithError(err).Fatal("Unable to generate json-schema")
	}

	// Write the generated JSONSchemas to files:
	if err := schemaWriter.WriteJSONSchemasToFiles(generatedJSONSchemas); err != nil {
		logger.WithError(err).Fatal("Unable to write JSONSchemas")
	}

	// Write a file containing go-constants for the generated JSON schemas:
	if config.GoConstants {
		if err := schemaWriter.WriteGoConstantsToFile(generatedJSONSchemas); err != nil {
			logger.WithError(err).Fatal("Unable to write go-constants")
		}
	}
}
