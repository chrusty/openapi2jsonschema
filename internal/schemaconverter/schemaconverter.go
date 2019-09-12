package schemaconverter

import (
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/filewriter"
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/oapi2"
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/oapi3"
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	"github.com/sirupsen/logrus"
)

// New returns either an Oapi2 or Oapi3 converter (according to the config), plus a writer:
func New(config *types.Config, logger *logrus.Logger) (types.Converter, types.Writer, error) {

	writer := filewriter.New(config, logger)

	if config.V3 {
		converter, err := oapi3.New(config, logger)
		return converter, writer, err
	}

	converter, err := oapi2.New(config, logger)
	return converter, writer, err
}

// NewV2 returns an OpenAPIv2 schema converter:
func NewV2(config *types.Config, logger *logrus.Logger) (types.Converter, error) {
	return oapi2.New(config, logger)
}

// NewV3 returns an OpenAPIv3 schema converter:
func NewV3(config *types.Config, logger *logrus.Logger) (types.Converter, error) {
	return oapi3.New(config, logger)
}

// NewWriter returns a schema writer:
func NewWriter(config *types.Config, logger *logrus.Logger) types.Writer {
	return filewriter.New(config, logger)
}
