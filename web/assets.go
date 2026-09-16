package web

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	AssetCSSPlaceholder = "__ASSET_CSS__"
	AssetJSPlaceholder  = "__ASSET_JS__"
)

var (
	CSSHash string
	JSHash  string
)

func init() {
	CSSHash = assetHash("app.css")
	JSHash = assetHash("app.js")
}

func assetHash(name string) string {
	b, err := FS.ReadFile(name)
	if err != nil {
		return "0000"
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:4]
}
