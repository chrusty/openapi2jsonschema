package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	openapi2proto "github.com/NYTimes/openapi2proto"
	jsonschema "github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type GeneratedJSONSchema struct {
	Name  string
	Bytes []byte
}

func MapOpenAPIDefinitionsToJSONSchema(api *openapi2proto.APIDefinition) ([]GeneratedJSONSchema, error) {
	var generatedJSONSchemas []GeneratedJSONSchema

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
		var generatedJSONSchema GeneratedJSONSchema
		var err error

		// Report:
		logWithLevel(LOG_INFO, "Processing schema-definition: %s", definitionName)

		// Derive a jsonschema:
		definitionJSONSchema, err = convertItems(api, definitionName, definition)
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

// GenerateJSONSchemas takes an openAPI *APIDefinition and converts it into JSONSchemas:
func GenerateJSONSchemas(api *openapi2proto.APIDefinition) (err error) {

	// Store the output in here:
	generatedJSONSchemas, err := MapOpenAPIDefinitionsToJSONSchema(api)
	if err != nil {
		return errors.Wrap(err, "could not map openapi definitions to jsonschema")
	}

	// Output the API name:
	logWithLevel(LOG_DEBUG, "API: %v (%v)", api.Info.Title, api.Info.Description)

	// Generate a GoConstants file (if we've been asked to):
	if goConstants {
		writeAllJSONSchemasToGoConstants(generatedJSONSchemas)
	}

	// Also write them all out to jsonschema files:
	return writeAllJSONSchemasToFile(generatedJSONSchemas)
}

// Converts an OpenAPI "Items" into a JSON-Schema:
func convertItems(api *openapi2proto.APIDefinition, itemName string, items *openapi2proto.Items) (definitionJSONSchema jsonschema.Type, err error) {
	var nestedProperties map[string]*openapi2proto.Items
	var requiredProperties interface{}

	// Prepare a new jsonschema:
	definitionJSONSchema = jsonschema.Type{
		AdditionalProperties: generateAdditionalProperties(blockAdditionalProperties),
		Description:          strings.Replace(items.Description, "`", "'", -1),
		MaxLength:            items.MaxLength,
		MinLength:            items.MinLength,
		Pattern:              items.Pattern,
		Properties:           make(map[string]*jsonschema.Type),
		Title:                items.Name,
	}

	// Self-contained schemas:
	if items.Schema != nil {
		itemsMap, recurseError := recurseNestedProperties(api, map[string]*openapi2proto.Items{"schema": items.Schema})
		err = recurseError
		definitionJSONSchema = *itemsMap["schema"]
		return
	}

	// Arrays of self-defined parameters:
	if items.Ref == "" && items.Type == gojsonschema.TYPE_ARRAY {
		itemsMap, recurseError := recurseNestedProperties(api, map[string]*openapi2proto.Items{"items": items.Items})
		err = recurseError
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if items.Ref == "" && items.Type != gojsonschema.TYPE_ARRAY && items.Schema == nil {
		definitionJSONSchema.Type = mapOpenAPITypeToJSONSchemaType(items.Type)
		requiredProperties = items.Required
		definitionJSONSchema.Properties, err = recurseNestedProperties(api, items.Model.Properties)
		definitionJSONSchema.Enum = mapEnums(items.Enum, items.Type)

		if items.Format != nil {
			definitionJSONSchema.Format = items.Format.(string)
		}
	}

	// Referenced models:
	if items.Ref != "" {
		var enum []string
		nestedProperties, definitionJSONSchema.Type, requiredProperties, enum, err = lookupReference(api, items.Ref)
		definitionJSONSchema.Properties, err = recurseNestedProperties(api, nestedProperties)
		definitionJSONSchema.Enum = mapEnums(enum, definitionJSONSchema.Type)
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {
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
		logWithLevel(LOG_WARN, "Can't determine JSONSchema type (%v)", openAPIType)
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
func lookupReference(api *openapi2proto.APIDefinition, referencePath string) (nestedProperties map[string]*openapi2proto.Items, definitionJSONSchemaType string, requiredProperties interface{}, enum []string, err error) {

	// Break up the path:
	_, reference, err := splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	logWithLevel(LOG_DEBUG, "Found a referenced model (%s)", reference)
	referencedDefinition, ok := api.Definitions[reference]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", reference)
		return
	}

	// Use the model's items, type, and required-properties:
	nestedProperties = referencedDefinition.Model.Properties
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
			logWithLevel(LOG_DEBUG, "Adding required property (%s)", requiredProperty)
			requiredProperties = append(requiredProperties, requiredProperty.(string))
		}
	} else {
		logWithLevel(LOG_DEBUG, "Failed to type-assert required-properties list")
	}

	return
}

func recurseNestedProperties(api *openapi2proto.APIDefinition, nestedProperties map[string]*openapi2proto.Items) (properties map[string]*jsonschema.Type, err error) {
	properties = make(map[string]*jsonschema.Type)

	// Recurse nested items:
	for nestedItemsName, nestedItems := range nestedProperties {
		logWithLevel(LOG_DEBUG, "Processing nested-items: %s", nestedItemsName)
		recurseddefinitionJSONSchema, err := convertItems(api, nestedItemsName, nestedItems)
		if err != nil {
			return properties, fmt.Errorf("Failed to convert items %s: %v", nestedItemsName, err)
		}
		properties[nestedItemsName] = &recurseddefinitionJSONSchema
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
