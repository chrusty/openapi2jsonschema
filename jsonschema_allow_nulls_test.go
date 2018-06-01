package main

import (
	"testing"

	openapi2proto "github.com/NYTimes/openapi2proto/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateJSONSchemas_AllowNulls_FlatObject(t *testing.T) {

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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_flat-object.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 1, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}

func Test_GenerateJSONSchemas_AllowNulls_FlatObjectWithEnum(t *testing.T) {

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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_flat-object-with-enum.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 1, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}

func Test_GenerateJSONSchemas_AllowNulls_ObjectWithArrays(t *testing.T) {

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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_array-of-referenced-object.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 2, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}

func Test_GenerateJSONSchemas_AllowNulls_ObjectWithPattern(t *testing.T) {

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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_object-with-pattern.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 1, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}

func Test_GenerateJSONSchemas_AllowNulls_ReferencedObject(t *testing.T) {

	var expectedSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "required": [
        "user_id",
        "user_name"
    ],
    "properties": {
        "contact_additional_props_map": {
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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_referenced-object.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 2, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}

func Test_GenerateJSONSchemas_AllowNulls_NumberWithMinMax(t *testing.T) {

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

	allowNullValues = true
	api, err := openapi2proto.LoadFile("sample/swagger2_flat-object-with-number-options.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
	assert.Len(t, schemas, 1, "Unexpected number of schemas returned")
	assert.JSONEq(t, expectedSchema, string(schemas[0].Bytes), "Unexpected schema received")
}
