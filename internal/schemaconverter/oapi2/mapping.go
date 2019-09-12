package oapi2

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	openAPI "github.com/NYTimes/openapi2proto/openapi"
	jsonSchema "github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// mapOpenAPIDefinitionsToJSONSchema converts an OpenAPI "Spec" into a JSONSchema:
func (c *Converter) mapOpenAPIDefinitionsToJSONSchema() ([]types.GeneratedJSONSchema, error) {
	var generatedJSONSchemas []types.GeneratedJSONSchema

	// // if we have no definitions then copy them from parameters:
	// if c.spec.Definitions == nil {
	// 	c.logger.Debug("No definitions found - copying from parameters")
	// 	c.spec.Definitions = map[string]*openAPI.Schema{}

	// 	// jam all the parameters into the normal 'definitions' for easier reference.
	// 	for paramName, param := range c.spec.Parameters {
	// 		c.logger.WithField("parameter_name", paramName).Trace("Found a parameter")
	// 		c.spec.Parameters[paramName] = param
	// 	}
	// }

	// Iterate through any schemas we find, creating JSONSchemas for each:
	for schemaName, schema := range c.spec.Definitions {
		var generatedJSONSchema types.GeneratedJSONSchema

		c.logger.WithField("schema_name", schemaName).Trace("Found a schema")

		// Derive a jsonschema:
		definitionJSONSchema, err := c.convertItems(schemaName, schema)
		if err != nil {
			return nil, errors.Wrap(err, "could not derive a json schema")
		}
		definitionJSONSchema.Version = jsonSchema.Version

		// Marshal the JSONSchema:
		generatedJSONSchema.Name = schemaName
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

// convertItems converts an OpenAPI "Items" into a JSON-Schema:
func (c *Converter) convertItems(itemName string, openAPISchema *openAPI.Schema) (jsonSchema.Type, error) {

	// Prepare a new jsonschema:
	definitionJSONSchema := jsonSchema.Type{
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
	// 	itemsMap, err := c.recurseNestedSchemas(map[string]*openAPI.Schema{"items": openAPISchema.Items})
	// 	if err != nil {
	// 		return definitionJSONSchema, err
	// 	}
	// 	return *itemsMap["items"], nil
	// }

	// Arrays of self-defined parameters:
	if openAPISchema.Ref == "" && openAPISchema.Type.Contains(gojsonschema.TYPE_ARRAY) {
		itemsMap, err := c.recurseNestedSchemas(map[string]*openAPI.Schema{"items": openAPISchema.Items})
		if err != nil {
			return definitionJSONSchema, err
		}
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if openAPISchema.Ref == "" && !openAPISchema.Type.Contains(gojsonschema.TYPE_ARRAY) && openAPISchema.Items == nil {
		properties, err := c.recurseNestedSchemas(openAPISchema.Properties)
		definitionJSONSchema.Properties = properties
		if err != nil {
			return definitionJSONSchema, err
		}

		if c.config.AllowNullValues {
			if openAPISchema.AdditionalProperties != nil && len(openAPISchema.AdditionalProperties.Type) == 1 {
				definitionJSONSchema.AdditionalProperties = json.RawMessage(fmt.Sprintf("{\"type\": \"%v\"}", openAPISchema.AdditionalProperties.Type[0]))
				c.nestedAdditionalProperties[itemName] = definitionJSONSchema.AdditionalProperties
			}

			if openAPISchema.AdditionalProperties != nil && openAPISchema.AdditionalProperties.Ref != "" {
				referenceName, err := c.splitReferencePath(openAPISchema.AdditionalProperties.Ref)
				if err == nil {
					if p, ok := c.nestedAdditionalProperties[referenceName]; ok {
						definitionJSONSchema.AdditionalProperties = p
					}
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
		nestedProperties, lookedupReferenceType, required, enum, err := c.lookupReference(openAPISchema.Ref)
		if err != nil {
			return definitionJSONSchema, err
		}
		definitionJSONSchema.Required = required
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
			referenceName, _ := c.splitReferencePath(openAPISchema.Ref)
			// if err == nil {
			if p, ok := c.nestedAdditionalProperties[referenceName]; ok {
				definitionJSONSchema.AdditionalProperties = p
			}
			// }
		}
	}

	// Maintain a list of required items:
	if definitionJSONSchema.Type == gojsonschema.TYPE_OBJECT {

		// if we have any nested items in the object then we should process them
		if additionalPropertiesSchema := openAPISchema.AdditionalProperties; additionalPropertiesSchema != nil {
			schema, err := c.convertItems(itemName, additionalPropertiesSchema)

			// Annoyingly since "additionalProperties" can actually be a
			// boolean or an object we have to marshal the resulting schema
			// so we can assign the raw bytes to back
			raw, err := json.Marshal(schema)
			definitionJSONSchema.AdditionalProperties = raw
			return definitionJSONSchema, err
		}
	}

	return definitionJSONSchema, nil
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

// splitReferencePath breaks up a reference path into its components (OpenAPI2 references look like "#/definitions/Something"):
func (c *Converter) splitReferencePath(ref string) (string, error) {

	// split on '/':
	refDatas := strings.Split(ref, "/")

	// Return the 3rd component (definition name):
	if len(refDatas) > 1 {
		return refDatas[2], nil
	}
	return "", fmt.Errorf("Unable to split this reference (%s)", ref)
}

// lookupReference looks up a reference and returns its schema and metadata:
func (c *Converter) lookupReference(referencePath string) (nestedProperties map[string]*openAPI.Schema, definitionJSONSchemaType string, requiredProperties []string, enum []string, err error) {
	c.logger.WithField("referencePath", referencePath).Trace("Looking up reference")

	// Break up the path:
	referenceName, err := c.splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	c.logger.WithField("reference", referenceName).Trace("Found a referenced model")
	referencedDefinition, ok := c.spec.Definitions[referenceName]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", referenceName)
		return
	}

	// Use the model's items, type, and required-properties:
	return referencedDefinition.Properties, c.mapOpenAPITypeToJSONSchemaType(referencedDefinition.Type), referencedDefinition.Required, referencedDefinition.Enum, nil
}

// recurseNestedSchemas converts nested openAPISchemas:
func (c *Converter) recurseNestedSchemas(nestedSchemas map[string]*openAPI.Schema) (properties map[string]*jsonSchema.Type, err error) {
	properties = make(map[string]*jsonSchema.Type)

	// Recurse nested items:
	for nestedSchemaName, nestedSchema := range nestedSchemas {
		c.logger.WithField("nested_schema_name", nestedSchemaName).Trace("Processing nested-items")
		recursedJSONSchema, err := c.convertItems(nestedSchemaName, nestedSchema)
		if err != nil {
			return properties, errors.Wrapf(err, "Failed to convert items (%s)", nestedSchemaName)
		}
		properties[nestedSchemaName] = &recursedJSONSchema
	}

	return properties, nil
}

// generateAdditionalProperties returns true or false:
func (c *Converter) generateAdditionalProperties() []byte {
	// BlockAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if c.config.BlockAdditionalProperties {
		return []byte("false")
	}

	return []byte("true")
}
