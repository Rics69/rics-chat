package openapi

import "embed"

// SwaggerFS — сгенерированный buf'ом swagger, вшитый в бинарь.
// rk-boot читает его через GlobalAppCtx.AddEmbedFS и показывает на /sw/
//
//go:embed *.swagger.json
var SwaggerFS embed.FS
