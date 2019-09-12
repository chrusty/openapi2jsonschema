package oapi3

import (
	"testing"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSONSchemasFlatObject(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "id",
        "name"
    ],
    "properties": {
        "id": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some ID",
            "format": "date-time"
        },
        "name": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some name"
        }
    },
    "additionalProperties": true,
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/flat-object.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasFlatObjectWithEnum(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "id",
        "name"
    ],
    "properties": {
        "id": {
            "additionalProperties": true,
            "enum": [
                "a",
                "b",
                "c"
            ],
            "type": "string"
        }
    },
    "additionalProperties": true,
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/flat-object-with-enum.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasObjectWithArrays(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "group_id",
        "group_name",
        "contacts_schema"
    ],
    "properties": {
        "contacts_schema": {
            "items": {
                "required": [
                    "email_address"
                ],
                "properties": {
                    "email_address": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "first_name": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "last_name": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "phone_number": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "spam": {
                        "additionalProperties": true,
                        "type": "boolean",
                        "description": "Send this person spam?"
                    }
                },
                "additionalProperties": true,
                "type": "object"
            },
            "additionalProperties": true
        },
        "crufts": {
            "items": {
                "required": [
                    "id"
                ],
                "properties": {
                    "description": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "id": {
                        "additionalProperties": true,
                        "type": "integer"
                    }
                },
                "additionalProperties": true,
                "type": "object"
            },
            "additionalProperties": true
        },
        "group_id": {
            "additionalProperties": true,
            "type": "integer",
            "description": "Some ID"
        },
        "group_name": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some name"
        },
        "remarks": {
            "items": {
                "additionalProperties": true,
                "type": "string"
            },
            "additionalProperties": true
        }
    },
    "additionalProperties": true,
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/array-of-referenced-object.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 2)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasObjectWithPattern(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "id",
        "name"
    ],
    "properties": {
        "id": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some ID"
        },
        "locale": {
            "pattern": "^[a-z]{2}(?:-[A-Z][a-z]{3})?(?:-(?:[A-Z]{2}))?$",
            "additionalProperties": true,
            "type": "string",
            "description": "BCP 47 locale string"
        },
        "name": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some name"
        },
        "phone_number": {
            "pattern": "^[\\d|\\+|\\(]+[\\)|\\d|\\s|-]*[\\d]$",
            "additionalProperties": true,
            "type": "string",
            "description": "Phone number"
        }
    },
    "additionalProperties": true,
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/object-with-pattern.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasReferencedObject(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "user_id",
        "user_name"
    ],
    "properties": {
        "contact_additional_props_map": {
            "additionalProperties": {
                "required": [
                    "email_address"
                ],
                "properties": {
                    "email_address": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "first_name": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "last_name": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "phone_number": {
                        "additionalProperties": true,
                        "type": "string"
                    },
                    "spam": {
                        "additionalProperties": true,
                        "type": "boolean",
                        "description": "Send this person spam?"
                    }
                },
                "additionalProperties": true,
                "type": "object"
            },
            "type": "object"
        },
        "contact_ref": {
            "required": [
                "email_address"
            ],
            "properties": {
                "email_address": {
                    "additionalProperties": true,
                    "type": "string"
                },
                "first_name": {
                    "additionalProperties": true,
                    "type": "string"
                },
                "last_name": {
                    "additionalProperties": true,
                    "type": "string"
                },
                "phone_number": {
                    "additionalProperties": true,
                    "type": "string"
                },
                "spam": {
                    "additionalProperties": true,
                    "type": "boolean",
                    "description": "Send this person spam?"
                }
            },
            "additionalProperties": true,
            "type": "object"
        },
        "user_id": {
            "additionalProperties": true,
            "type": "integer",
            "description": "Some ID"
        },
        "user_name": {
            "additionalProperties": true,
            "type": "string",
            "description": "Some name"
        }
    },
    "additionalProperties": true,
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/referenced-object.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 2)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasNumberWithMinMax(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "latitude"
    ],
    "properties": {
        "description": {
            "minLength": 1,
            "maxLength": 10,
            "additionalProperties": true,
            "type": "string"
        },
        "latitude": {
            "maximum": 90,
            "minimum": -90,
            "additionalProperties": true,
            "type": "number",
            "description": "The latitude in degrees. It must be in the range [-90.0, +90.0]",
            "format": "double"
        }
    },
    "additionalProperties": true,
    "type": "object",
    "description": "Specifies a geographic location in terms of its Latitude and Longitude"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/flat-object-with-number-options.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}

func TestGenerateJSONSchemasMap(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "additionalProperties": {
        "additionalProperties": true,
        "type": "string"
    },
    "type": "object"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           false,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/openapi3/with_map.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes))
}
