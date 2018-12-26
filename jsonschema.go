package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	openAPI "github.com/NYTimes/openapi2proto/openapi"
	jsonSchema "github.com/alecthomas/jsonschema"
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
func GenerateJSONSchemas(api *openAPI.Spec) (err error) {

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
func MapOpenAPIDefinitionsToJSONSchema(openAPISpec *openAPI.Spec) ([]GeneratedJSONSchema, error) {
	var generatedJSONSchemas []GeneratedJSONSchema

	// if we have no definitions then copy them from parameters:
	if openAPISpec.Definitions == nil {
		logWithLevel(logDebug, "No definitions found - copying from parameters...")
		openAPISpec.Definitions = map[string]*openAPI.Schema{}
	}

	// jam all the parameters into the normal 'definitions' for easier reference.
	for paramName, param := range openAPISpec.Parameters {
		logWithLevel(logDebug, "Found a parameter: %s", paramName)
		openAPISpec.Parameters[paramName] = param
	}

	// Iterate through the definitions, creating JSONSchemas for each:
	for definitionName, definition := range openAPISpec.Definitions {

		var definitionJSONSchema jsonSchema.Type
		var generatedJSONSchema GeneratedJSONSchema
		var err error

		// Report:
		logWithLevel(logInfo, "Processing schema-definition: %s", definitionName)

		// Derive a jsonschema:
		definitionJSONSchema, err = convertItems(openAPISpec, definitionName, definition)
		if err != nil {
			return nil, errors.Wrap(err, "could not derive a json schema")

		}
		definitionJSONSchema.Version = jsonSchema.Version

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

var (
	nestedAdditionalProperties = map[string]json.RawMessage{}
)

// Converts an OpenAPI "Items" into a JSON-Schema:
func convertItems(openAPISpec *openAPI.Spec, itemName string, openAPISchema *openAPI.Schema) (definitionJSONSchema jsonSchema.Type, err error) {
	var nestedProperties map[string]*openAPI.Schema
	// Prepare a new jsonschema:
	definitionJSONSchema = jsonSchema.Type{
		AdditionalProperties: generateAdditionalProperties(blockAdditionalProperties),
		Description:          strings.Replace(openAPISchema.Description, "`", "'", -1),
		MaxLength:            openAPISchema.MaxLength,
		MinLength:            openAPISchema.MinLength,
		Pattern:              openAPISchema.Pattern,
		Properties:           make(map[string]*jsonSchema.Type),
		Title:                openAPISchema.ProtoName,
		Minimum:              openAPISchema.Minimum,
		Maximum:              openAPISchema.Maximum,
	}

	// // Self-contained schemas:
	// if openAPISchema.Items != nil {
	// 	itemsMap, recurseError := recurseNestedSchemas(openAPISpec, map[string]*openAPI.Schema{"items": openAPISchema.Items})
	// 	err = recurseError
	// 	definitionJSONSchema = *itemsMap["items"]
	// 	return
	// }

	// Arrays of self-defined parameters:
	if openAPISchema.Ref == "" && openAPISchema.Type.Contains(gojsonschema.TYPE_ARRAY) {
		itemsMap, recurseError := recurseNestedSchemas(openAPISpec, map[string]*openAPI.Schema{"items": openAPISchema.Items})
		err = recurseError
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if openAPISchema.Ref == "" && !openAPISchema.Type.Contains(gojsonschema.TYPE_ARRAY) && openAPISchema.Items == nil {
		definitionJSONSchema.Properties, err = recurseNestedSchemas(openAPISpec, openAPISchema.Properties)

		if allowNullValues {
			if openAPISchema.AdditionalProperties != nil && len(openAPISchema.AdditionalProperties.Type) == 1 {
				definitionJSONSchema.AdditionalProperties = json.RawMessage(fmt.Sprintf("{\"type\": \"%v\"}", openAPISchema.AdditionalProperties.Type[0]))
				nestedAdditionalProperties[itemName] = definitionJSONSchema.AdditionalProperties
			}

			if openAPISchema.AdditionalProperties != nil && openAPISchema.AdditionalProperties.Ref != "" {
				_, name, _ := splitReferencePath(openAPISchema.AdditionalProperties.Ref)
				if p, ok := nestedAdditionalProperties[name]; ok {
					definitionJSONSchema.AdditionalProperties = p
				}
			}

			definitionJSONSchema.OneOf = []*jsonSchema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: mapOpenAPITypeToJSONSchemaType(openAPISchema.Type)},
			}
		} else {
			definitionJSONSchema.Type = mapOpenAPITypeToJSONSchemaType(openAPISchema.Type)
		}

		definitionJSONSchema.Required = openAPISchema.Required
		definitionJSONSchema.Enum = mapEnums(openAPISchema.Enum, openAPISchema.Type)

		if openAPISchema.Format != "" {
			definitionJSONSchema.Format = openAPISchema.Format
		}
	}

	// Referenced models:
	if openAPISchema.Ref != "" {
		var enum []string
		var lookedupReferenceType string
		nestedProperties, lookedupReferenceType, definitionJSONSchema.Required, enum, err = lookupReference(openAPISpec, openAPISchema.Ref)
		if allowNullValues {
			definitionJSONSchema.OneOf = []*jsonSchema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: lookedupReferenceType},
			}
		} else {
			definitionJSONSchema.Type = lookedupReferenceType
		}
		definitionJSONSchema.Properties, err = recurseNestedSchemas(openAPISpec, nestedProperties)
		definitionJSONSchema.Enum = mapEnums(enum, []string{definitionJSONSchema.Type})

		if openAPISchema.Ref != "" {
			_, name, _ := splitReferencePath(openAPISchema.Ref)
			if p, ok := nestedAdditionalProperties[name]; ok {
				definitionJSONSchema.AdditionalProperties = p
			}
		}
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {

		// if we have any nested items in the object then we should process them
		if additionalPropertiesSchema := openAPISchema.AdditionalProperties; additionalPropertiesSchema != nil {
			var schema jsonSchema.Type
			var raw json.RawMessage
			schema, err = convertItems(openAPISpec, itemName, additionalPropertiesSchema)

			// Annoyingly since "additionalProperties" can actually be a
			// boolean or an object we have to marshal the resulting schema
			// so we can assign the raw bytes to back
			raw, err = json.Marshal(schema)
			definitionJSONSchema.AdditionalProperties = raw
			return
		}

	}

	return
}

func mapEnums(items []string, openAPISchemaTypes openAPI.SchemaType) []interface{} {
	var result []interface{}

	for _, item := range items {
		var value interface{}
		if openAPISchemaTypes.Contains(gojsonschema.TYPE_NUMBER) {
			value, _ = strconv.Atoi(item)
		} else {
			value = item
		}
		result = append(result, value)
	}

	return result
}

// Map OpenAPI types to JSONSchema types:
func mapOpenAPITypeToJSONSchemaType(openAPISchemaTypes openAPI.SchemaType) string {

	// Make sure we were actually given a type:
	if len(openAPISchemaTypes) == 0 {
		logWithLevel(logWarn, "Can't determine JSONSchema type (%v)", openAPISchemaTypes)
		return gojsonschema.TYPE_NULL
	}

	// Switch on the first type:
	switch openAPISchemaTypes[0] {
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
		logWithLevel(logWarn, "Can't determine JSONSchema type (%v)", openAPISchemaTypes)
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
func lookupReference(openAPISpec *openAPI.Spec, referencePath string) (nestedProperties map[string]*openAPI.Schema, definitionJSONSchemaType string, requiredProperties []string, enum []string, err error) {

	// Break up the path:
	_, reference, err := splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	logWithLevel(logDebug, "Found a referenced model (%s)", reference)
	referencedDefinition, ok := openAPISpec.Definitions[reference]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", reference)
		return
	}

	// Use the model's items, type, and required-properties:
	nestedProperties = referencedDefinition.Properties
	definitionJSONSchemaType = mapOpenAPITypeToJSONSchemaType(referencedDefinition.Type)
	requiredProperties = referencedDefinition.Required
	enum = referencedDefinition.Enum

	return
}

// recurseNestedSchemas converts nested openAPISchemas:
func recurseNestedSchemas(openAPISpec *openAPI.Spec, nestedSchemas map[string]*openAPI.Schema) (properties map[string]*jsonSchema.Type, err error) {
	properties = make(map[string]*jsonSchema.Type)

	// Recurse nested items:
	for nestedSchemaName, nestedSchema := range nestedSchemas {
		logWithLevel(logDebug, "Processing nested-items: %s", nestedSchemaName)
		recursedJSONSchema, err := convertItems(openAPISpec, nestedSchemaName, nestedSchema)
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
