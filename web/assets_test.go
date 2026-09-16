package web

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestAssetHashes(t *testing.T) {
	for name, hash := range map[string]string{
		"app.css": CSSHash,
		"app.js":  JSHash,
	} {
		if len(hash) != 4 {
			t.Fatalf("%s hash length %d", name, len(hash))
		}
		b, err := FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		want := hex.EncodeToString(sum[:])[:4]
		if hash != want {
			t.Fatalf("%s hash %q want %q", name, hash, want)
		}
	}
}
