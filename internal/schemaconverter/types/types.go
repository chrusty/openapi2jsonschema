package types

// Config represents all the options for the converter:
type Config struct {
	AllowNullValues           bool
	BlockAdditionalProperties bool
	JSONSchemaFileExtention   string
	GoConstants               bool
	GoConstantsFilename       string
	OutPath                   string
	SpecPath                  string
}

// GeneratedJSONSchema is a JSONSchema that has been mapped from an OpenAPI spec:
type GeneratedJSONSchema struct {
	Name  string
	Bytes []byte
}
