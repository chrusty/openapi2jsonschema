package oapi3

import (
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Converter performs schema conversion:
type Converter struct {
	config  *types.Config
	logger  *logrus.Logger
	swagger *openapi3.Swagger
}

// New takes a config and returns a new Converter:
func New(config *types.Config, logger *logrus.Logger) (*Converter, error) {

	// Load the OpenAPI spec:
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(config.SpecPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to load spec (%s)", config.SpecPath)
	}

	logger.WithField("title", swagger.Info.Title).WithField("description", swagger.Info.Description).Info("Prepared a converter for Swagger / OpenAPI3")

	// Return a new *Converter:
	return &Converter{
		config:  config,
		logger:  logger,
		swagger: swagger,
	}, nil
}

// GenerateJSONSchemas takes an OpenAPI "Spec" and converts each definition into a JSONSchema:
func (c *Converter) GenerateJSONSchemas() ([]types.GeneratedJSONSchema, error) {

	c.logger.Debug("Converting API")

	// Store the output in here:
	generatedJSONSchemas, err := c.mapOpenAPIDefinitionsToJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "could not map openapi definitions to jsonschema")
	}

	return generatedJSONSchemas, nil
}
