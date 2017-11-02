build:
	@go build -o openapi2jsonschema

test:
	@mkdir -p out
	@./openapi2jsonschema -debug -block_additional_properties -spec="sample/swagger2.yaml" -out="./out"
