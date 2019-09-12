default: build

build:
	@echo "Generating binary (bin/openapi2jsonschema) ..."
	@mkdir -p bin
	@go build -o bin/openapi2jsonschema cmd/openapi2jsonschema/main.go

install:
	@GO111MODULE=on go get -u github.com/chrusty/openapi2jsonschema/cmd/openapi2jsonschema && go install github.com/chrusty/openapi2jsonschema/cmd/openapi2jsonschema

build_linux:
	@echo "Generating Linux-amd64 binary (bin/openapi2jsonschema.linux-amd64) ..."
	@GOOS=linux GOARCH=amd64 go build -o bin/openapi2jsonschema.linux-amd64 bin/openapi2jsonschema cmd/openapi2jsonschema/main.go


samples: build
	@echo "Generating sample JSON-Schemas ..."
	@mkdir -p out
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/swagger2/flat-object.yaml                 -go_constants  -block_additional_properties  -out=./out
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/swagger2/referenced-object.yaml           -go_constants  -block_additional_properties  -out=./out
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/swagger2/object-with-pattern.yaml         -go_constants  -block_additional_properties  -out=./out
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/swagger2/array-of-referenced-object.yaml  -go_constants  -block_additional_properties  -out=./out
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/openapi3/with_map.yaml                    -go_constants  -block_additional_properties  -out=./out -v3
	@bin/openapi2jsonschema  -spec=internal/schemaconverter/samples/openapi3/petstore.yaml                    -go_constants  -block_additional_properties  -out=./out -v3

test:
	@go test ./... -cover
