package assets

import (
	"embed"
)

//go:embed fonts/NotoSansTC-Light.otf
var FontData []byte

//go:embed icons/*
var IconFS embed.FS
