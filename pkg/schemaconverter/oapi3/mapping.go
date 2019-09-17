package oapi3

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	jsonSchema "github.com/alecthomas/jsonschema"
	"github.com/chrusty/openapi2jsonschema/pkg/schemaconverter/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// mapOpenAPIDefinitionsToJSONSchema converts an OpenAPI "Spec" into a JSONSchema:
func (c *Converter) mapOpenAPIDefinitionsToJSONSchema() ([]types.GeneratedJSONSchema, error) {
	var generatedJSONSchemas []types.GeneratedJSONSchema

	// Iterate through any schemas we find, creating JSONSchemas for each:
	for schemaName, schema := range c.swagger.Components.Schemas {
		var generatedJSONSchema types.GeneratedJSONSchema

		c.logger.WithField("schema_name", schemaName).Trace("Found a schema")

		// Derive a jsonschema:
		definitionJSONSchema, err := c.convertSchema(schema)
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

// convertSchema converts an OpenAPI Schema into a JSON-Schema:
func (c *Converter) convertSchema(openAPISchema *openapi3.SchemaRef) (jsonSchema.Type, error) {

	// Prepare a new jsonschema:
	definitionJSONSchema := jsonSchema.Type{
		AdditionalProperties: c.generateAdditionalProperties(),
		Description:          strings.Replace(openAPISchema.Value.Description, "`", "'", -1),
		MinLength:            int(openAPISchema.Value.MinLength),
		Pattern:              openAPISchema.Value.Pattern,
		Properties:           make(map[string]*jsonSchema.Type),
	}

	if openAPISchema.Value.MaxLength != nil {
		definitionJSONSchema.MaxLength = int(*openAPISchema.Value.MaxLength)
	}

	if openAPISchema.Value.Min != nil {
		definitionJSONSchema.Minimum = int(*openAPISchema.Value.Min)
	}

	if openAPISchema.Value.Max != nil {
		definitionJSONSchema.Maximum = int(*openAPISchema.Value.Max)
	}

	// Arrays of self-defined parameters:
	if openAPISchema.Ref == "" && strings.Contains(openAPISchema.Value.Type, gojsonschema.TYPE_ARRAY) {
		itemsMap, err := c.recurseNestedSchemas(map[string]*openapi3.SchemaRef{"items": openAPISchema.Value.Items})
		if err != nil {
			return definitionJSONSchema, err
		}
		definitionJSONSchema.Items = itemsMap["items"]
	}

	// Single-instances of self-defined parameters:
	if openAPISchema.Ref == "" && !strings.Contains(openAPISchema.Value.Type, gojsonschema.TYPE_ARRAY) && openAPISchema.Value.Items == nil {
		properties, err := c.recurseNestedSchemas(openAPISchema.Value.Properties)
		definitionJSONSchema.Properties = properties
		if err != nil {
			return definitionJSONSchema, err
		}

		// See if there are any additionalProperties to convert:
		if openAPISchema.Value.AdditionalProperties != nil {
			if convertedAdditionalProperties, err := c.convertSchema(openAPISchema.Value.AdditionalProperties); err == nil {
				c.logger.
					WithField("AdditionalProperties.Value.Ref", openAPISchema.Value.AdditionalProperties.Ref).
					WithField("AdditionalProperties.Value.Type", openAPISchema.Value.AdditionalProperties.Value.Type).
					Tracef("Converted additional properties: %v", convertedAdditionalProperties)
				additionalPropertiesJSON, err := json.Marshal(convertedAdditionalProperties)
				if err != nil {
					return definitionJSONSchema, errors.Wrapf(err, "Unable to marshal additionalProperties to JSON")
				}
				definitionJSONSchema.AdditionalProperties = additionalPropertiesJSON
			}
		}

		// If we allow nulls then make NULL an option:
		if c.config.AllowNullValues {
			definitionJSONSchema.OneOf = []*jsonSchema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: c.mapOpenAPITypeToJSONSchemaType(openAPISchema.Value.Type)},
			}
		} else {
			definitionJSONSchema.Type = c.mapOpenAPITypeToJSONSchemaType(openAPISchema.Value.Type)
		}

		definitionJSONSchema.Required = openAPISchema.Value.Required
		definitionJSONSchema.Enum = openAPISchema.Value.Enum

		if openAPISchema.Value.Format != "" {
			definitionJSONSchema.Format = openAPISchema.Value.Format
		}
	}

	// Referenced models:
	if openAPISchema.Ref != "" {
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
		definitionJSONSchema.Enum = enum

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

		// If we have any nested items in the object then we should process them:
		if openAPISchema.Value.AdditionalProperties != nil {
			schema, err := c.convertSchema(openAPISchema.Value.AdditionalProperties)
			if err != nil {
				return definitionJSONSchema, err
			}

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

// mapOpenAPITypeToJSONSchemaType maps OpenAPI types to JSONSchema types:
func (c *Converter) mapOpenAPITypeToJSONSchemaType(openAPISchemaType string) string {

	// Make sure we were actually given a type:
	if openAPISchemaType == "" {
		c.logger.WithField("type", openAPISchemaType).Warn("Can't determine JSONSchema type")
		return gojsonschema.TYPE_NULL
	}

	// Switch on the first type:
	switch openAPISchemaType {
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
		c.logger.WithField("type", openAPISchemaType).Warn("Can't determine JSONSchema type")
		return gojsonschema.TYPE_NULL
	}
}

// splitReferencePath breaks up a reference path into its components (OpenAPI3 references look like "#/components/schemas/Something"):
func (c *Converter) splitReferencePath(ref string) (string, error) {

	// split on '/':
	refDatas := strings.Split(ref, "/")

	// Return the 4th component (definition name):
	if len(refDatas) > 2 {
		return refDatas[3], nil
	}
	return "", fmt.Errorf("Unable to split this reference (%s)", ref)
}

// lookupReference looks up a reference and returns its schema and metadata:
func (c *Converter) lookupReference(referencePath string) (nestedProperties map[string]*openapi3.SchemaRef, definitionJSONSchemaType string, requiredProperties []string, enum []interface{}, err error) {
	c.logger.WithField("referencePath", referencePath).Trace("Looking up reference")

	// Break up the path:
	referenceName, err := c.splitReferencePath(referencePath)
	if err != nil {
		return
	}

	// Look up the referenced model:
	c.logger.WithField("reference", referenceName).Trace("Found a referenced model")
	referencedDefinition, ok := c.swagger.Components.Schemas[referenceName]
	if !ok {
		err = fmt.Errorf("Unable to find a referenced model (%s)", referenceName)
		return
	}

	// Use the model's items, type, and required-properties:
	return referencedDefinition.Value.Properties, c.mapOpenAPITypeToJSONSchemaType(referencedDefinition.Value.Type), referencedDefinition.Value.Required, referencedDefinition.Value.Enum, nil
}

// recurseNestedSchemas converts nested openAPISchemas:
func (c *Converter) recurseNestedSchemas(nestedSchemas map[string]*openapi3.SchemaRef) (properties map[string]*jsonSchema.Type, err error) {
	properties = make(map[string]*jsonSchema.Type)

	// Recurse nested items:
	for nestedSchemaName, nestedSchema := range nestedSchemas {
		c.logger.WithField("nested_schema_name", nestedSchemaName).Trace("Processing nested-items")
		recursedJSONSchema, err := c.convertSchema(nestedSchema)
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
