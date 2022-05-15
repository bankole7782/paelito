package paelito_internal

import (
  "embed"
)

//go:embed templates/*
var content embed.FS

//go:embed statics/*
var contentStatics embed.FS
