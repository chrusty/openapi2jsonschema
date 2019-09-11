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
	@bin/openapi2jsonschema  -spec="internal/schemaconverter/samples/swagger2_flat-object.yaml"                 -go_constants  -block_additional_properties  -out="./out"
	@bin/openapi2jsonschema  -spec="internal/schemaconverter/samples/swagger2_referenced-object.yaml"           -go_constants  -block_additional_properties  -out="./out"
	@bin/openapi2jsonschema  -spec="internal/schemaconverter/samples/swagger2_object-with-pattern.yaml"         -go_constants  -block_additional_properties  -out="./out"
	@bin/openapi2jsonschema  -spec="internal/schemaconverter/samples/swagger2_array-of-referenced-object.yaml"  -go_constants  -block_additional_properties  -out="./out"

test:
	@go test ./... -cover
