package main

import (
	"encoding/json"
	"fmt"
	"strings"

	openapi2proto "github.com/NYTimes/openapi2proto"
	jsonschema "github.com/alecthomas/jsonschema"
	"github.com/davecgh/go-spew/spew"
	"github.com/xeipuuv/gojsonschema"
)

// GenerateJSONSchemas takes an openAPI *APIDefinition and converts it into JSONSchema files (in the "output" directory):
func GenerateJSONSchemas(api *openapi2proto.APIDefinition) (err error) {

	// Output the API name:
	logWithLevel(LOG_DEBUG, "API: %v (%v)", api.Info.Title, api.Info.Description)
	spew.Dump(api)

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
		definitionJSONSchema, err = convertItems(api, definitionName, definition, false)
		if err != nil {
			return err
		}

		// Marshal the JSONSchema:
		definitionJSONSchemaJSON, err := json.MarshalIndent(definitionJSONSchema, "", "    ")
		if err != nil {
			return err
		}

		// Write the schemaJSON out to a file:
		if err := writeToFile(jsonSchemaFileName, definitionJSONSchemaJSON); err != nil {
			return err
		}
	}

	return

}

// Converts an OpenAPI "Items" into a JSON-Schema:
func convertItems(api *openapi2proto.APIDefinition, itemName string, items *openapi2proto.Items, nested bool) (jsonschema.Type, error) {

	var nestedModel openapi2proto.Model
	var requiredProperties interface{}

	// Prepare a new jsonschema:
	definitionJSONSchema := jsonschema.Type{
		Title:       items.Name,
		Properties:  make(map[string]*jsonschema.Type),
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

	// Either use the items we have, or lookup referenced models if required:
	if items.Ref == "" {
		// Regular old model:
		nestedModel = items.Model
		definitionJSONSchema.Type = mapOpenAPITypeToJSONSchemaType(items.Type)
		requiredProperties = items.Required
	} else {
		// Referenced model:
		_, reference, err := splitReference(items.Ref)
		if err != nil {
			logWithLevel(LOG_ERROR, "Failed to split a reference (%s)", items.Ref)
			return definitionJSONSchema, err
		}

		// Look up the referenced model:
		logWithLevel(LOG_DEBUG, "Found a referenced model (%s)", reference)
		referencedDefinition, ok := api.Definitions[reference]
		if !ok {
			logWithLevel(LOG_ERROR, "Unable to find a referenced model (%s)", reference)
			return definitionJSONSchema, fmt.Errorf("Unable to find a referenced model (%s)", reference)
		}

		// Use the model's items, type, and required-properties:
		nestedModel = referencedDefinition.Model
		definitionJSONSchema.Type = mapOpenAPITypeToJSONSchemaType(referencedDefinition.Type)
		requiredProperties = referencedDefinition.Required
	}

	// Recurse nested items:
	for nestedItemsName, nestedItems := range nestedModel.Properties {
		logWithLevel(LOG_DEBUG, "Processing nested-items: %s", nestedItemsName)
		recurseddefinitionJSONSchema, err := convertItems(api, nestedItemsName, nestedItems, true)
		if err != nil {
			logWithLevel(LOG_ERROR, "Failed to convert items %s in %s: %v", nestedItemsName, itemName, err)
			return definitionJSONSchema, err
		}
		definitionJSONSchema.Properties[nestedItemsName] = &recurseddefinitionJSONSchema
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {

		// Ugly type-assertion to get the list of required properties:
		if requiredPropertiesList, ok := requiredProperties.([]interface{}); ok {

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
	case nil, "":
		return gojsonschema.TYPE_NULL
	default:
		logWithLevel(LOG_WARN, "Can't determine JSONSchema type (%v)", openAPIType)
		return gojsonschema.TYPE_NULL
	}
}

func splitReference(ref string) (string, string, error) {

	// split on '/'
	refDatas := strings.Split(ref, "/")

	// Return the 2nd and 3rd components (source and definition name):
	if len(refDatas) > 1 {
		return refDatas[1], refDatas[2], nil
	}
	return ref, "", fmt.Errorf("Unable to split this reference (%s)", ref)
}
