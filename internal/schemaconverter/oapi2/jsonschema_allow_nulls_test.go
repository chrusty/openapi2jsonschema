package oapi2

import (
	"fmt"
	"testing"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSONSchemasAllowNullsFlatObject(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "id",
        "name"
    ],
    "properties": {
        "id": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some ID",
            "format": "date-time"
        },
        "name": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some name"
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/flat-object.yaml",
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

func TestGenerateJSONSchemasAllowNullsFlatObjectWithEnum(t *testing.T) {

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
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ]
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/flat-object-with-enum.yaml",
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

func TestGenerateJSONSchemasAllowNullsObjectWithArrays(t *testing.T) {

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
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "first_name": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "last_name": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "phone_number": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "spam": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "boolean"
                            }
                        ],
                        "description": "Send this person spam?"
                    }
                },
                "additionalProperties": true,
                "oneOf": [
                    {
                        "type": "null"
                    },
                    {
                        "type": "object"
                    }
                ]
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
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "id": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "integer"
                            }
                        ]
                    }
                },
                "additionalProperties": true,
                "oneOf": [
                    {
                        "type": "null"
                    },
                    {
                        "type": "object"
                    }
                ]
            },
            "additionalProperties": true
        },
        "group_id": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "integer"
                }
            ],
            "description": "Some ID"
        },
        "group_name": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some name"
        },
        "remarks": {
            "items": {
                "additionalProperties": true,
                "oneOf": [
                    {
                        "type": "null"
                    },
                    {
                        "type": "string"
                    }
                ]
            },
            "additionalProperties": true
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/array-of-referenced-object.yaml",
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

func TestGenerateJSONSchemasAllowNullsObjectWithPattern(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "id",
        "name"
    ],
    "properties": {
        "id": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some ID"
        },
        "locale": {
            "pattern": "^[a-z]{2}(?:-[A-Z][a-z]{3})?(?:-(?:[A-Z]{2}))?$",
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "BCP 47 locale string"
        },
        "name": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some name"
        },
        "phone_number": {
            "pattern": "^[\\d|\\+|\\(]+[\\)|\\d|\\s|-]*[\\d]$",
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Phone number"
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/object-with-pattern.yaml",
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

func TestGenerateJSONSchemasAllowNullsReferencedObject(t *testing.T) {

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
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "first_name": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "last_name": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "phone_number": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "string"
                            }
                        ]
                    },
                    "spam": {
                        "additionalProperties": true,
                        "oneOf": [
                            {
                                "type": "null"
                            },
                            {
                                "type": "boolean"
                            }
                        ],
                        "description": "Send this person spam?"
                    }
                },
                "additionalProperties": true,
                "oneOf": [
                    {
                        "type": "null"
                    },
                    {
                        "type": "object"
                    }
                ]
            },
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "object"
                }
            ]
        },
        "contact_ref": {
            "required": [
                "email_address"
            ],
            "properties": {
                "email_address": {
                    "additionalProperties": true,
                    "oneOf": [
                        {
                            "type": "null"
                        },
                        {
                            "type": "string"
                        }
                    ]
                },
                "first_name": {
                    "additionalProperties": true,
                    "oneOf": [
                        {
                            "type": "null"
                        },
                        {
                            "type": "string"
                        }
                    ]
                },
                "last_name": {
                    "additionalProperties": true,
                    "oneOf": [
                        {
                            "type": "null"
                        },
                        {
                            "type": "string"
                        }
                    ]
                },
                "phone_number": {
                    "additionalProperties": true,
                    "oneOf": [
                        {
                            "type": "null"
                        },
                        {
                            "type": "string"
                        }
                    ]
                },
                "spam": {
                    "additionalProperties": true,
                    "oneOf": [
                        {
                            "type": "null"
                        },
                        {
                            "type": "boolean"
                        }
                    ],
                    "description": "Send this person spam?"
                }
            },
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "object"
                }
            ]
        },
        "user_id": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "integer"
                }
            ],
            "description": "Some ID"
        },
        "user_name": {
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ],
            "description": "Some name"
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/referenced-object.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 2)
	if !assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes)) {
		fmt.Println(string(generatedJSONSchemas[0].Bytes))
	}
}

func TestGenerateJSONSchemasAllowNullsNumberWithMinMax(t *testing.T) {

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
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "string"
                }
            ]
        },
        "latitude": {
            "maximum": 90,
            "minimum": -90,
            "additionalProperties": true,
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "number"
                }
            ],
            "description": "The latitude in degrees. It must be in the range [-90.0, +90.0]",
            "format": "double"
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ],
    "description": "Specifies a geographic location in terms of its Latitude and Longitude"
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/flat-object-with-number-options.yaml",
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

func TestGenerateJSONSchemasMapAllowingNullValues(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "additionalProperties": {
        "additionalProperties": true,
        "oneOf": [
            {
                "type": "null"
            },
            {
                "type": "string"
            }
        ]
    },
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/with_map.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 1)
	if !assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[0].Bytes)) {
		fmt.Println(string(generatedJSONSchemas[0].Bytes))
	}
}

func TestGenerateJSONSchemasMapInAReffedObjectAllowingNullValues(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "properties": {
        "object_with_map": {
            "additionalProperties": {
                "additionalProperties": true,
                "oneOf": [
                    {
                        "type": "null"
                    },
                    {
                        "type": "object"
                    }
                ]
            },
            "oneOf": [
                {
                    "type": "null"
                },
                {
                    "type": "object"
                }
            ]
        }
    },
    "additionalProperties": true,
    "oneOf": [
        {
            "type": "null"
        },
        {
            "type": "object"
        }
    ]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/with_map_in_ref.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 2)
	if !assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[1].Bytes)) {
		fmt.Println(string(generatedJSONSchemas[1].Bytes))
	}

}

func TestGenerateJSONSchemasMapInAReffedObjectAllowingNullValues2(t *testing.T) {

	var expectedSchema = `{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"properties": {
	    "object_with_map": {
		    "additionalProperties": {
		        "type": "string"
		    },
		    "oneOf": [
		        {
			        "type": "null"
		        },
		        {
			        "type": "object"
		        }
		    ]
	    }
	},
	"additionalProperties": true,
	"oneOf": [
	    {
		    "type": "null"
	    },
	    {
		    "type": "object"
	    }
	]
}`

	// Prepare a new schema converter:
	schemaConverter, err := New(&types.Config{
		AllowNullValues:           true,
		BlockAdditionalProperties: false,
		JSONSchemaFileExtention:   "jsonschema",
		SpecPath:                  "../samples/swagger2/with_map_in_ref_2.yaml",
	}, logrus.New())
	require.NoError(t, err)

	// Convert the spec:
	generatedJSONSchemas, err := schemaConverter.GenerateJSONSchemas()
	require.NoError(t, err)

	assert.NoError(t, err)
	assert.NotNil(t, generatedJSONSchemas)
	assert.Len(t, generatedJSONSchemas, 2)
	if !assert.JSONEq(t, expectedSchema, string(generatedJSONSchemas[1].Bytes)) {
		fmt.Println(string(generatedJSONSchemas[1].Bytes))
	}
}
