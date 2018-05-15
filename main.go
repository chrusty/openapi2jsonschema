package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	openapi2proto "github.com/NYTimes/openapi2proto/openapi"
)

const (
	logDebug = 0
	logInfo  = 1
	logWarn  = 2
	logError = 3
	logFatal = 4
	logPanic = 5

	jsonschemaFileExtention = "jsonschema"
	goConstantsFilename     = "jsonschemas"
)

var (
	allowNullValues           bool
	blockAdditionalProperties bool
	debugLogging              bool
	goConstants               bool
	outPath, specPath         string
	logLevels                 = map[LogLevel]string{
		0: "DEBUG",
		1: "INFO",
		2: "WARN",
		3: "ERROR",
		4: "FATAL",
		5: "PANIC",
	}
)

// LogLevel defines constants for logging levels:
type LogLevel int

func logWithLevel(logLevel LogLevel, logFormat string, logParams ...interface{}) {
	// If we're not doing debug logging then just return:
	if logLevel < logInfo && !debugLogging {
		return
	}

	// Otherwise log:
	logMessage := fmt.Sprintf(logFormat, logParams...)
	log.Printf(fmt.Sprintf("[%v] %v", logLevels[logLevel], logMessage))

	// If we're handling a fatal error:
	if logLevel >= logFatal {
		os.Exit(2)
	}
}

func init() {
	flag.BoolVar(&allowNullValues, "allow_null_values", false, "Allow NULL values as well as the defined types?")
	flag.BoolVar(&blockAdditionalProperties, "block_additional_properties", false, "Block additional properties?")
	flag.BoolVar(&debugLogging, "debug", false, "Log debug messages?")
	flag.BoolVar(&goConstants, "go_constants", false, "Output GoLang constants (in addition to JSONSchemas)?")
	flag.StringVar(&specPath, "spec", "../../spec.yaml", "Location of the swagger spec file")
	flag.StringVar(&outPath, "out", "./out", "Where to write jsonschema output files to")
}

func main() {

	flag.Parse()

	// Load the OpenAPI spec:
	api, err := openapi2proto.LoadFile(specPath)
	if err != nil {
		logWithLevel(logFatal, "Unable to load spec: %v", err)
	}

	// Generate JSONSchemas:
	if err := GenerateJSONSchemas(api); err != nil {
		logWithLevel(logFatal, "Unable to generate json-schema: %v", err)
	}

}
