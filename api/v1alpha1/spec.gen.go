// Package v1alpha1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package v1alpha1

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
