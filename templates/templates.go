package templates

import _ "embed"

//go:embed project/main.craft
var Main string

//go:embed project/smoke.craft
var Test string
