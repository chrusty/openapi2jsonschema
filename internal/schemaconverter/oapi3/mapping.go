package oapi3

import (
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"
	"github.com/davecgh/go-spew/spew"
)

// mapOpenAPIDefinitionsToJSONSchema converts an OpenAPI "Spec" into a JSONSchema:
func (c *Converter) mapOpenAPIDefinitionsToJSONSchema() ([]types.GeneratedJSONSchema, error) {
	var generatedJSONSchemas []types.GeneratedJSONSchema

	spew.Dump(c.swagger)

	// List any schemas we find:
	for schemaName := range c.swagger.Components.Schemas {
		c.logger.WithField("schema_name", schemaName).Trace("Found a schema")
	}

	return generatedJSONSchemas, nil

}
