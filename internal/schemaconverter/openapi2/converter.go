package openapi2

import (
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	openapi2proto "github.com/NYTimes/openapi2proto/openapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Converter performs schema conversion:
type Converter struct {
	api    *openapi2proto.Spec
	config *types.Config
	logger *logrus.Logger
}

// New takes a config and returns a new Converter:
func New(config *types.Config, logger *logrus.Logger) (*Converter, error) {

	// Load the OpenAPI spec:
	api, err := openapi2proto.LoadFile(config.SpecPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to load spec (%s)", config.SpecPath)
	}

	logger.WithField("title", api.Info.Title).WithField("description", api.Info.Description).Info("Prepared a converter for API")

	// Return a new *Converter:
	return &Converter{
		api:    api,
		config: config,
		logger: logger,
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