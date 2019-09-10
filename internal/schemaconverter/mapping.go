package schemaconverter

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

// GeneratedJSONSchema is a JSONSchema that has been mapped from an OpenAPI spec:
type GeneratedJSONSchema struct {
	Name  string
	Bytes []byte
}

// mapOpenAPIDefinitionsToJSONSchema converts an OpenAPI "Spec" into a JSONSchema:
func (c *Converter) mapOpenAPIDefinitionsToJSONSchema() ([]GeneratedJSONSchema, error) {
	var generatedJSONSchemas []GeneratedJSONSchema

	// if we have no definitions then copy them from parameters:
	if c.api.Definitions == nil {
		c.logger.Debug("No definitions found - copying from parameters")
		c.api.Definitions = map[string]*openAPI.Schema{}
	}

	// jam all the parameters into the normal 'definitions' for easier reference.
	for paramName, param := range c.api.Parameters {
		c.logger.WithField("parameter_name", paramName).Trace("Found a parameter")
		c.api.Parameters[paramName] = param
	}

	// Iterate through the definitions, creating JSONSchemas for each:
	for definitionName, definition := range c.api.Definitions {

		var definitionJSONSchema jsonSchema.Type
		var generatedJSONSchema GeneratedJSONSchema
		var err error

		// Report:
		c.logger.WithField("definition_name", definitionName).Debug("Processing schema-definition")

		// Derive a jsonschema:
		definitionJSONSchema, err = c.convertItems(definitionName, definition)
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

// convertItems converts an OpenAPI "Items" into a JSON-Schema:
func (c *Converter) convertItems(itemName string, openAPISchema *openAPI.Schema) (definitionJSONSchema jsonSchema.Type, err error) {
	var nestedProperties map[string]*openAPI.Schema
	// Prepare a new jsonschema:
	definitionJSONSchema = jsonSchema.Type{
		AdditionalProperties: c.generateAdditionalProperties(),
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
		itemsMap, recurseError := c.recurseNestedSchemas(map[string]*openAPI.Schema{"items": openAPISchema.Items})
		err = recurseError
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if openAPISchema.Ref == "" && !openAPISchema.Type.Contains(gojsonschema.TYPE_ARRAY) && openAPISchema.Items == nil {
		definitionJSONSchema.Properties, err = c.recurseNestedSchemas(openAPISchema.Properties)

		if c.config.AllowNullValues {
			if openAPISchema.AdditionalProperties != nil && len(openAPISchema.AdditionalProperties.Type) == 1 {
				definitionJSONSchema.AdditionalProperties = json.RawMessage(fmt.Sprintf("{\"type\": \"%v\"}", openAPISchema.AdditionalProperties.Type[0]))
				nestedAdditionalProperties[itemName] = definitionJSONSchema.AdditionalProperties
			}

			if openAPISchema.AdditionalProperties != nil && openAPISchema.AdditionalProperties.Ref != "" {
				_, name, _ := c.splitReferencePath(openAPISchema.AdditionalProperties.Ref)
				if p, ok := nestedAdditionalProperties[name]; ok {
					definitionJSONSchema.AdditionalProperties = p
				}
			}

			definitionJSONSchema.OneOf = []*jsonSchema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: c.mapOpenAPITypeToJSONSchemaType(openAPISchema.Type)},
			}
		} else {
			definitionJSONSchema.Type = c.mapOpenAPITypeToJSONSchemaType(openAPISchema.Type)
		}

		definitionJSONSchema.Required = openAPISchema.Required
		definitionJSONSchema.Enum = c.mapEnums(openAPISchema.Enum, openAPISchema.Type)

		if openAPISchema.Format != "" {
			definitionJSONSchema.Format = openAPISchema.Format
		}
	}

	// Referenced models:
	if openAPISchema.Ref != "" {
		var enum []string
		var lookedupReferenceType string
		nestedProperties, lookedupReferenceType, definitionJSONSchema.Required, enum, err = c.lookupReference(openAPISchema.Ref)
		if c.config.AllowNullValues {
			definitionJSONSchema.OneOf = []*jsonSchema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: lookedupReferenceType},
			}
		} else {
			definitionJSONSchema.Type = lookedupReferenceType
		}
		definitionJSONSchema.Properties, err = c.recurseNestedSchemas(nestedProperties)
		definitionJSONSchema.Enum = c.mapEnums(enum, []string{definitionJSONSchema.Type})

		if openAPISchema.Ref != "" {
			_, name, _ := c.splitReferencePath(openAPISchema.Ref)
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
			schema, err = c.convertItems(itemName, additionalPropertiesSchema)

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

// mapEnums maps OpenAPI enums to JSONSchema types:
func (c *Converter) mapEnums(items []string, openAPISchemaTypes openAPI.SchemaType) []interface{} {
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

// mapOpenAPITypeToJSONSchemaType maps OpenAPI types to JSONSchema types:
func (c *Converter) mapOpenAPITypeToJSONSchemaType(openAPISchemaTypes openAPI.SchemaType) string {

	// Make sure we were actually given a type:
	if len(openAPISchemaTypes) == 0 {
		c.logger.WithField("type", openAPISchemaTypes).Warn("Can't determine JSONSchema type")
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
		c.logger.WithField("type", openAPISchemaTypes).Warn("Can't determine JSONSchema type")
		return gojsonschema.TYPE_NULL
	}
}

// splitReferencePath breaks up a reference path into its components:
func (c *Converter) splitReferencePath(ref string) (string, string, error) {

	// split on '/'
	refDatas := strings.Split(ref, "/")

	// Return the 2nd and 3rd components (source and definition name):
	if len(refDatas) > 1 {
		return refDatas[1], refDatas[2], nil
	}
	return ref, "", fmt.Errorf("Unable to split this reference (%s)", ref)
}

// lookupReference looks up a reference and returns its schema and metadata:
func (c *Converter) lookupReference(referencePath string) (nestedProperties map[string]*openAPI.Schema, definitionJSONSchemaType string, requiredProperties []string, enum []string, err error) {

	// Break up the path:
	_, reference, err := c.splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	c.logger.WithField("reference", reference).Trace("Found a referenced model")
	referencedDefinition, ok := c.api.Definitions[reference]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", reference)
		return
	}

	// Use the model's items, type, and required-properties:
	nestedProperties = referencedDefinition.Properties
	definitionJSONSchemaType = c.mapOpenAPITypeToJSONSchemaType(referencedDefinition.Type)
	requiredProperties = referencedDefinition.Required
	enum = referencedDefinition.Enum

	return
}

// recurseNestedSchemas converts nested openAPISchemas:
func (c *Converter) recurseNestedSchemas(nestedSchemas map[string]*openAPI.Schema) (properties map[string]*jsonSchema.Type, err error) {
	properties = make(map[string]*jsonSchema.Type)

	// Recurse nested items:
	for nestedSchemaName, nestedSchema := range nestedSchemas {
		c.logger.WithField("nested_schema_name", nestedSchemaName).Trace("Processing nested-items")
		recursedJSONSchema, err := c.convertItems(nestedSchemaName, nestedSchema)
		if err != nil {
			return properties, fmt.Errorf("Failed to convert items %s: %v", nestedSchemaName, err)
		}
		properties[nestedSchemaName] = &recursedJSONSchema
	}

	return
}

// generateAdditionalProperties returns true or false:
func (c *Converter) generateAdditionalProperties() []byte {
	// BlockAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if c.config.BlockAdditionalProperties {
		return []byte("false")
	}

	return []byte("true")
}
