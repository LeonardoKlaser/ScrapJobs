package templates

import "embed"

//go:embed emails/*.html
var EmailTemplates embed.FS
