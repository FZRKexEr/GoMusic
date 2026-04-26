package static

import (
	"embed"
	"fmt"
	"mime"
	"path"
)

//go:embed index.html styles.css app.js
var files embed.FS

type Asset struct {
	Content     []byte
	ContentType string
}

func Read(name string) (Asset, error) {
	name = path.Clean("/" + name)[1:]
	if name == "" {
		name = "index.html"
	}

	switch name {
	case "index.html", "styles.css", "app.js":
	default:
		return Asset{}, fmt.Errorf("static asset %q not found", name)
	}

	content, err := files.ReadFile(name)
	if err != nil {
		return Asset{}, err
	}

	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return Asset{
		Content:     content,
		ContentType: contentType,
	}, nil
}
