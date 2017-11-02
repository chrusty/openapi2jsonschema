package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openapi2proto "github.com/NYTimes/openapi2proto"
	jsonschema "github.com/alecthomas/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

// GenerateJSONSchemas takes an openAPI *APIDefinition and converts it into JSONSchema files (in the "output" directory):
func GenerateJSONSchemas(api *openapi2proto.APIDefinition) (err error) {

	// Output the API name:
	logWithLevel(LOG_DEBUG, "API: %v (%v)", api.Info.Title, api.Info.Description)

	// if we have no definitions then copy them from parameters:
	if api.Definitions == nil {
		logWithLevel(LOG_DEBUG, "No definitions found - copying from parameters...")
		api.Definitions = map[string]*openapi2proto.Items{}
	}

	// jam all the parameters into the normal 'definitions' for easier reference.
	for paramName, param := range api.Parameters {
		logWithLevel(LOG_DEBUG, "Found a parameter: %s", paramName)
		api.Definitions[paramName] = param
	}

	// Iterate through the definitions, creating JSONSchemas for each:
	for definitionName, definition := range api.Definitions {

		var definitionJSONSchema jsonschema.Type
		var err error

		// Generate a filename to store the JSONSchema in:
		jsonSchemaFileName := generateFileName(definitionName)

		// Report:
		logWithLevel(LOG_INFO, "Processing schema-definition: %s => %s", definitionName, jsonSchemaFileName)

		// Derive a jsonschema:
		if definition.Ref != "" {
			logWithLevel(LOG_DEBUG, "Converting nested schema-definition reference: %s (%s)", definitionName, definition.Ref)
			definitionJSONSchema, err = convertItems(definitionName, definition, false)
		} else {
			logWithLevel(LOG_DEBUG, "Converting nested schema-definition: %s", definitionName)
			definitionJSONSchema, err = convertItems(definitionName, definition, false)
		}
		if err != nil {
			return err
		}

		definitionJSONSchemaJSON, err := json.MarshalIndent(definitionJSONSchema, "", "    ")
		if err != nil {
			return err
		}

		// Write the schemaJson out to a file:
		if err := writeToFile(jsonSchemaFileName, definitionJSONSchemaJSON); err != nil {
			return err
		}
	}

	return

}

// Converts an OpenAPI "Items" into a JSON-Schema:
func convertItems(itemName string, items *openapi2proto.Items, nested bool) (jsonschema.Type, error) {

	// Prepare a new jsonschema:
	definitionJSONSchema := jsonschema.Type{
		Title:       items.Name,
		Properties:  make(map[string]*jsonschema.Type),
		Type:        mapOpenAPITypeToJSONSchemaType(items.Type),
		Description: items.Description,
	}

	// Set the schema version (but only at the base level):
	if !nested {
		definitionJSONSchema.Version = jsonschema.Version
	}

	// blockAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if blockAdditionalProperties {
		definitionJSONSchema.AdditionalProperties = []byte("false")
	} else {
		definitionJSONSchema.AdditionalProperties = []byte("true")
	}

	// Recurse nested items:
	for nestedItemsName, nestedItems := range items.Model.Properties {
		logWithLevel(LOG_DEBUG, "Processing nested-items: %s", nestedItemsName)
		recurseddefinitionJSONSchema, err := convertItems(nestedItemsName, nestedItems, true)
		if err != nil {
			logWithLevel(LOG_ERROR, "Failed to convert items %s in %s: %v", nestedItemsName, itemName, err)
			return definitionJSONSchema, err
		}
		definitionJSONSchema.Properties[nestedItemsName] = &recurseddefinitionJSONSchema
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {

		// Ugly type-assertion to get the list of required properties:
		if requiredPropertiesList, ok := items.Required.([]interface{}); ok {

			// Iterate through the required-properties list, and add them to the JSONSchema:
			for _, requiredProperty := range requiredPropertiesList {
				logWithLevel(LOG_DEBUG, "Adding required property (%s)", requiredProperty)
				definitionJSONSchema.Required = append(definitionJSONSchema.Required, requiredProperty.(string))
			}
		} else {
			logWithLevel(LOG_ERROR, "Failed to type-assert required-properties list")
		}
	}

	return definitionJSONSchema, nil
}

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

func mapOpenAPITypeToJSONSchemaType(openAPIType interface{}) string {
	switch openAPIType {
	case "array":
		return gojsonschema.TYPE_ARRAY
	case "boolean":
		return gojsonschema.TYPE_BOOLEAN
	case "integer":
		return gojsonschema.TYPE_INTEGER
	case "number":
		return gojsonschema.TYPE_NUMBER
	case "object":
		return gojsonschema.TYPE_OBJECT
	case "string":
		return gojsonschema.TYPE_STRING
	case "":
		return gojsonschema.TYPE_NULL
	default:
		logWithLevel(LOG_WARN, "Can't determine JSONSchema type (%v)", openAPIType)
		return gojsonschema.TYPE_OBJECT
	}
}
