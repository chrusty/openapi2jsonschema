package main

import (
	"encoding/json"
	"fmt"
	"strings"

	openapi2proto "github.com/NYTimes/openapi2proto"
	jsonschema "github.com/alecthomas/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

// GenerateJSONSchemas takes an openAPI *APIDefinition and converts it into JSONSchemas:
func GenerateJSONSchemas(api *openapi2proto.APIDefinition) (err error) {

	// Store the output in here:
	var generatedJSONSchemas = make(map[string][]byte)

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

		// Report:
		logWithLevel(LOG_INFO, "Processing schema-definition: %s", definitionName)

		// Derive a jsonschema:
		definitionJSONSchema, err = convertItems(api, definitionName, definition)
		if err != nil {
			return err
		}
		definitionJSONSchema.Version = jsonschema.Version

		// Marshal the JSONSchema:
		generatedJSONSchemas[definitionName], err = json.MarshalIndent(definitionJSONSchema, "", "    ")
		if err != nil {
			return err
		}
	}

	// Generate a GoConstants file (if we've been asked to):
	if goConstants {
		writeAllJSONSchemasToGoConstants(generatedJSONSchemas)
	}

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
	}

	// Referenced models:
	if items.Ref != "" {
		nestedProperties, definitionJSONSchema.Type, requiredProperties, err = lookupReference(api, items.Ref)
		definitionJSONSchema.Properties, err = recurseNestedProperties(api, nestedProperties)
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {
		definitionJSONSchema.Required = buildRequiredPropertiesList(requiredProperties)
	}

	return
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
func lookupReference(api *openapi2proto.APIDefinition, referencePath string) (nestedProperties map[string]*openapi2proto.Items, definitionJSONSchemaType string, requiredProperties interface{}, err error) {

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
