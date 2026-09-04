package static

import "embed"

//go:embed all:app/*
var FS embed.FS
