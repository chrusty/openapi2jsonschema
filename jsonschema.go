package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	openapi2proto "github.com/NYTimes/openapi2proto/openapi"
	jsonschema "github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type GeneratedJSONSchema struct {
	Name  string
	Bytes []byte
}

// GenerateJSONSchemas takes an OpenAPI "Spec":
// * converts each definition into JSONSchemas
// * writes them to files
// * optionally writes GoLang constants for convenience
func GenerateJSONSchemas(api *openapi2proto.Spec) (err error) {

	// Store the output in here:
	generatedJSONSchemas, err := MapOpenAPIDefinitionsToJSONSchema(api)
	if err != nil {
		return errors.Wrap(err, "could not map openapi definitions to jsonschema")
	}

	// Output the API name:
	logWithLevel(logDebug, "API: %v (%v)", api.Info.Title, api.Info.Description)

	// Generate a GoConstants file (if we've been asked to):
	if goConstants {
		writeAllJSONSchemasToGoConstants(generatedJSONSchemas)
	}

	// Also write them all out to jsonschema files:
	return writeAllJSONSchemasToFile(generatedJSONSchemas)
}

// MapOpenAPIDefinitionsToJSONSchema converts an OpenAPI "Spec" into a JSONSchema:
func MapOpenAPIDefinitionsToJSONSchema(openapi2protoSpec *openapi2proto.Spec) ([]GeneratedJSONSchema, error) {
	var generatedJSONSchemas []GeneratedJSONSchema

	// if we have no definitions then copy them from parameters:
	if openapi2protoSpec.Definitions == nil {
		logWithLevel(logDebug, "No definitions found - copying from parameters...")
		openapi2protoSpec.Definitions = map[string]*openapi2proto.Schema{}
	}

	// jam all the parameters into the normal 'definitions' for easier reference.
	for paramName, param := range openapi2protoSpec.Parameters {
		logWithLevel(logDebug, "Found a parameter: %s", paramName)
		openapi2protoSpec.Parameters[paramName] = param
	}

	// Iterate through the definitions, creating JSONSchemas for each:
	for definitionName, definition := range openapi2protoSpec.Definitions {

		var definitionJSONSchema jsonschema.Type
		var generatedJSONSchema GeneratedJSONSchema
		var err error

		// Report:
		logWithLevel(logInfo, "Processing schema-definition: %s", definitionName)

		// Derive a jsonschema:
		definitionJSONSchema, err = convertItems(openapi2protoSpec, definitionName, definition)
		if err != nil {
			return nil, errors.Wrap(err, "could not derive a json schema")

		}
		definitionJSONSchema.Version = jsonschema.Version

		// Marshal the JSONSchema:
		generatedJSONSchema.Name = definitionName
		generatedJSONSchema.Bytes, err = json.MarshalIndent(definitionJSONSchema, "", "    ")
		if err != nil {
			return nil, errors.Wrap(err, "could not marshall json schema")
		}

		// Append the new jsonschema to our list:
		generatedJSONSchemas = append(generatedJSONSchemas, generatedJSONSchema)
	}

	// Sort the results (so they come out in a consistent order):
	sort.Slice(generatedJSONSchemas, func(i, j int) bool { return generatedJSONSchemas[i].Name < generatedJSONSchemas[j].Name })

	return generatedJSONSchemas, nil

}

// Converts an OpenAPI "Items" into a JSON-Schema:
func convertItems(openapi2protoSpec *openapi2proto.Spec, itemName string, openapi2protoSchema *openapi2proto.Schema) (definitionJSONSchema jsonschema.Type, err error) {
	var nestedProperties map[string]*openapi2proto.Schema
	var requiredProperties interface{}

	// Prepare a new jsonschema:
	definitionJSONSchema = jsonschema.Type{
		AdditionalProperties: generateAdditionalProperties(blockAdditionalProperties),
		Description:          strings.Replace(openapi2protoSchema.Description, "`", "'", -1),
		MaxLength:            openapi2protoSchema.MaxLength,
		MinLength:            openapi2protoSchema.MinLength,
		Pattern:              openapi2protoSchema.Pattern,
		Properties:           make(map[string]*jsonschema.Type),
		Title:                openapi2protoSchema.ProtoName,
		Minimum:              openapi2protoSchema.Minimum,
		Maximum:              openapi2protoSchema.Maximum,
	}

	// Self-contained schemas:
	if openapi2protoSchema.Items != nil {
		itemsMap, recurseError := recurseNestedSchemas(openapi2protoSpec, map[string]*openapi2proto.Schema{"items": openapi2protoSchema.Items})
		err = recurseError
		definitionJSONSchema = *itemsMap["items"]
		return
	}

	// Arrays of self-defined parameters:
	if openapi2protoSchema.Ref == "" && openapi2protoSchema.Type.Contains(gojsonschema.TYPE_ARRAY) {
		itemsMap, recurseError := recurseNestedSchemas(openapi2protoSpec, map[string]*openapi2proto.Schema{"items": openapi2protoSchema.Items})
		err = recurseError
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if openapi2protoSchema.Ref == "" && !openapi2protoSchema.Type.Contains(gojsonschema.TYPE_ARRAY) && openapi2protoSchema.Items == nil {
		definitionJSONSchema.Type = mapOpenAPITypeToJSONSchemaType(openapi2protoSchema.Type)
		requiredProperties = openapi2protoSchema.Required
		definitionJSONSchema.Properties, err = recurseNestedSchemas(openapi2protoSpec, openapi2protoSchema.Items.Properties)
		definitionJSONSchema.Enum = mapEnums(openapi2protoSchema.Enum, openapi2protoSchema.Type)

		if openapi2protoSchema.Format != "" {
			definitionJSONSchema.Format = openapi2protoSchema.Format
		}
	}

	// Referenced models:
	if openapi2protoSchema.Ref != "" {
		var enum []string
		nestedProperties, definitionJSONSchema.Type, requiredProperties, enum, err = lookupReference(openapi2protoSpec, openapi2protoSchema.Ref)
		definitionJSONSchema.Properties, err = recurseNestedSchemas(openapi2protoSpec, nestedProperties)
		definitionJSONSchema.Enum = mapEnums(enum, definitionJSONSchema.Type)
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {

		// if we have any nested items in the object then we should process them
		if additionalPropertiesSchema := openapi2protoSchema.AdditionalProperties; additionalPropertiesSchema != nil {
			var schema jsonschema.Type
			var raw json.RawMessage
			schema, err = convertItems(openapi2protoSpec, additionalPropertiesSchema.ProtoName, additionalPropertiesSchema)

			// Annoyingly since "additionalProperties" can actually be a
			// boolean or an object we have to marshal the resulting schema
			// so we can assign the raw bytes to back
			raw, err = json.Marshal(schema)
			definitionJSONSchema.AdditionalProperties = raw
			return
		}

		definitionJSONSchema.Required = buildRequiredPropertiesList(requiredProperties)
	}

	return
}

func mapEnums(items []string, itemsType interface{}) []interface{} {
	var result []interface{}

	for _, item := range items {
		var value interface{}
		if itemsType == gojsonschema.TYPE_NUMBER {
			value, _ = strconv.Atoi(item)
		} else {
			value = item
		}
		result = append(result, value)
	}

	return result
}

// Map OpenAPI types to JSONSchema types:
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
		logWithLevel(logWarn, "Can't determine JSONSchema type (%v)", openAPIType)
		return gojsonschema.TYPE_NULL
	}
}

// Break up a reference path into its components:
func splitReferencePath(ref string) (string, string, error) {

	// split on '/'
	refDatas := strings.Split(ref, "/")

	// Return the 2nd and 3rd components (source and definition name):
	if len(refDatas) > 1 {
		return refDatas[1], refDatas[2], nil
	}
	return ref, "", fmt.Errorf("Unable to split this reference (%s)", ref)
}

// Look up a reference and return its schema and metadata:
func lookupReference(api *openapi2proto.Spec, referencePath string) (nestedProperties map[string]*openapi2proto.Schema, definitionJSONSchemaType string, requiredProperties interface{}, enum []string, err error) {

	// Break up the path:
	_, reference, err := splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	logWithLevel(logDebug, "Found a referenced model (%s)", reference)
	referencedDefinition, ok := api.Definitions[reference]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", reference)
		return
	}

	// Use the model's items, type, and required-properties:
	nestedProperties = referencedDefinition.Items.Properties
	definitionJSONSchemaType = mapOpenAPITypeToJSONSchemaType(referencedDefinition.Type)
	requiredProperties = referencedDefinition.Required
	enum = referencedDefinition.Enum

	return
}

// Build a list of required-properties:
func buildRequiredPropertiesList(requiredPropertiesInterface interface{}) (requiredProperties []string) {

	// Ugly type-assertion to get the list of required properties:
	if requiredPropertiesList, ok := requiredPropertiesInterface.([]interface{}); ok {

		// Iterate through the required-properties list, and add them to the JSONSchema:
		for _, requiredProperty := range requiredPropertiesList {
			logWithLevel(logDebug, "Adding required property (%s)", requiredProperty)
			requiredProperties = append(requiredProperties, requiredProperty.(string))
		}
	} else {
		logWithLevel(logDebug, "Failed to type-assert required-properties list")
	}

	return
}

func recurseNestedSchemas(openapi2protoSpec *openapi2proto.Spec, nestedSchemas map[string]*openapi2proto.Schema) (properties map[string]*jsonschema.Type, err error) {
	properties = make(map[string]*jsonschema.Type)

	// Recurse nested items:
	for nestedSchemaName, nestedSchema := range nestedSchemas {
		logWithLevel(logDebug, "Processing nested-items: %s", nestedSchemaName)
		recursedJSONSchema, err := convertItems(openapi2protoSpec, nestedSchemaName, nestedSchema)
		if err != nil {
			return properties, fmt.Errorf("Failed to convert items %s: %v", nestedSchemaName, err)
		}
		properties[nestedSchemaName] = &recursedJSONSchema
	}

	return
}

// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
func generateAdditionalProperties(disallowAdditionalProperties bool) []byte {
	if disallowAdditionalProperties {
		return []byte("false")
	}

	return []byte("true")
}
