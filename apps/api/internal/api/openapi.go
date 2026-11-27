package api

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// contract_annotations holds public descriptions, examples and compatibility
// constraints that cannot be inferred from the Go wire types alone.
//
//go:embed contract_annotations.yaml
var contractAnnotations []byte

type document = map[string]any

// OpenAPIYAML exports the same operation registry used by the HTTP server.
func OpenAPIYAML() ([]byte, error) {
	api := newHuma(http.NewServeMux())
	registerRoutes(api, nil)
	api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[ErrorResponse](), true, "")

	raw, err := json.Marshal(api.OpenAPI())
	if err != nil {
		return nil, err
	}
	var generated document
	if err := json.Unmarshal(raw, &generated); err != nil {
		return nil, err
	}
	var spec document
	if err := yaml.Unmarshal(contractAnnotations, &spec); err != nil {
		return nil, err
	}
	if err := annotateContract(spec, generated); err != nil {
		return nil, err
	}
	var output yaml.Node
	if err := output.Encode(spec); err != nil {
		return nil, err
	}
	// Retain the example order and field order used by the mock fixture generator.
	var source yaml.Node
	if err := yaml.Unmarshal(contractAnnotations, &source); err != nil {
		return nil, err
	}
	examples := mappingValue(mappingValue(source.Content[0], "components"), "examples")
	*mappingValue(mappingValue(&output, "components"), "examples") = *examples
	var result bytes.Buffer
	encoder := yaml.NewEncoder(&result)
	encoder.SetIndent(2)
	if err := encoder.Encode(&output); err != nil {
		return nil, err
	}
	return result.Bytes(), encoder.Close()
}

func annotateContract(spec, generated document) error {
	components := object(spec, "components")
	schemas := object(components, "schemas")
	for name, value := range object(object(generated, "components"), "schemas") {
		metadata, documented := schemas[name].(document)
		if documented {
			schemas[name] = annotateSchema(value.(document), metadata)
		} else if name != "Error" && name != "ErrorDetail" && name != "ErrorModel" && name != "ErrorDetailModel" {
			schemas[name] = value
		}
	}

	annotations := object(spec, "operations")
	paths := object(generated, "paths")
	for path, item := range paths {
		methods := item.(document)
		var sharedParameters []any
		for method, value := range methods {
			op := value.(document)
			id := op["operationId"].(string)
			metadata := object(annotations, id)
			if metadata == nil {
				return fmt.Errorf("falta documentación para %s %s (%s)", method, path, id)
			}

			pathParams := metadata["pathParameters"].([]any)
			delete(metadata, "pathParameters")
			sharedParameters = pathParams
			params, _ := metadata["parameters"].([]any)
			if err := checkParameters(op, components, append(append([]any(nil), pathParams...), params...)); err != nil {
				return fmt.Errorf("%s: %w", id, err)
			}

			// Keep the published references and examples while taking body
			// schemas from the registered Go input and output types.
			if body := object(metadata, "requestBody"); body != nil {
				applyMediaSchema(body, object(op, "requestBody"))
			}
			for code, response := range object(metadata, "responses") {
				if strings.HasPrefix(code, "2") {
					applyMediaSchema(response.(document), object(object(op, "responses"), code))
				}
			}
			metadata["operationId"] = id
			methods[method] = metadata
			delete(annotations, id)
		}
		if len(sharedParameters) > 0 {
			methods["parameters"] = sharedParameters
		}
	}
	if len(annotations) > 0 {
		return fmt.Errorf("quedan %d operaciones documentadas sin registro", len(annotations))
	}
	delete(spec, "operations")
	spec["paths"] = paths
	return nil
}

func checkParameters(op, components document, parameters []any) error {
	registered := make(map[string]bool)
	if params, ok := op["parameters"].([]any); ok {
		for _, value := range params {
			param := value.(document)
			registered[parameterKey(param)] = param["required"] == true
		}
	}
	for _, value := range parameters {
		param := value.(document)
		if ref, ok := param["$ref"].(string); ok {
			param = object(object(components, "parameters"), strings.TrimPrefix(ref, "#/components/parameters/"))
		}
		key := parameterKey(param)
		required, ok := registered[key]
		if !ok || required != (param["required"] == true) {
			return fmt.Errorf("el parámetro %q difiere del registro Huma", key)
		}
		delete(registered, key)
	}
	if len(registered) != 0 {
		return fmt.Errorf("quedan %d parámetros sin documentar", len(registered))
	}
	return nil
}

func parameterKey(param document) string {
	return fmt.Sprint(param["in"], ":", param["name"])
}

func applyMediaSchema(metadata, generated document) {
	media := object(object(metadata, "content"), "application/json")
	if media != nil && media["schema"] == nil {
		media["schema"] = object(object(generated, "content"), "application/json")["schema"]
	}
}

func annotateSchema(generated, metadata document) document {
	metadata["type"] = generated["type"]
	if required := generated["required"]; required != nil {
		metadata["required"] = required
	}
	properties := object(generated, "properties")
	if len(properties) == 0 {
		return metadata
	}
	annotations := object(metadata, "properties")
	for name, value := range properties {
		base := value.(document)
		annotation := object(annotations, name)
		if annotation == nil {
			annotation = document{}
		}
		if annotation["type"] == nil && annotation["$ref"] == nil {
			annotation["type"] = base["type"]
		}
		properties[name] = annotation
	}
	metadata["properties"] = properties
	return metadata
}

func object(parent document, key string) document {
	value, _ := parent[key].(document)
	return value
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	panic("missing contract mapping: " + key)
}
