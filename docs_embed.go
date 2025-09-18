package main

import _ "embed"

//go:embed docs/openapi.json
var openAPISpec []byte
