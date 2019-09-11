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
