package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	openapi2proto "github.com/NYTimes/openapi2proto"
)

const (
	LOG_DEBUG = 0
	LOG_INFO  = 1
	LOG_WARN  = 2
	LOG_ERROR = 3
	LOG_FATAL = 4
	LOG_PANIC = 5

	JSONSCHEMA_FILE_EXTENTION = "jsonschema"
	GO_CONSTANTS_FILENAME     = "jsonschemas"
)

var (
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
	if logLevel < LOG_INFO && !debugLogging {
		return
	}

	// Otherwise log:
	logMessage := fmt.Sprintf(logFormat, logParams...)
	log.Printf(fmt.Sprintf("[%v] %v", logLevels[logLevel], logMessage))

	// If we're handling a fatal error:
	if logLevel >= LOG_FATAL {
		os.Exit(2)
	}
}

func init() {
	flag.BoolVar(&blockAdditionalProperties, "block_additional_properties", false, "Block additional properties")
	flag.BoolVar(&debugLogging, "debug", false, "Log debug messages")
	flag.BoolVar(&goConstants, "go_constants", false, "Output GoLang constants (in addition to JSONSchemas)")
	flag.StringVar(&specPath, "spec", "../../spec.yaml", "location of the swagger spec file")
	flag.StringVar(&outPath, "out", "./out", "where to write jsonschema output files to")
}

func main() {

	flag.Parse()

	// Load the OpenAPI spec:
	api, err := openapi2proto.LoadDefinition(specPath)
	if err != nil {
		logWithLevel(LOG_FATAL, "Unable to load spec: %v", err)
	}

	// Generate JSONSchemas:
	if err := GenerateJSONSchemas(api); err != nil {
		logWithLevel(LOG_FATAL, "Unable to generate json-schema: %v", err)
	}

}
