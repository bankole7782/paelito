package main

import (
  "embed"
)

//go:embed templates/*
var content embed.FS
