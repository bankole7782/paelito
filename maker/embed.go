package main

import (
  _ "embed"
)

//go:embed "english-stopwords.json"
var englishStopwords []byte
