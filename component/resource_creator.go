package component

import (
	"os"
	"path"
)

// TODO Store this in a config file
const BASE_DIR = "/home/leon/.luxbox"

type ResourceCreator struct{}

// TODO Move out of the resource creator (it doesn't create anything)
func (this *ResourceCreator) Exists(resourcePath ...string) bool {
	resource := buildPath(resourcePath)

	if _, err := os.Stat(resource); err != nil {
		return false
	}

	return true
}

func (this *ResourceCreator) Create(path ...string) (*os.File, error) {
	return create(os.O_CREATE|os.O_WRONLY, path)
}

func (this *ResourceCreator) Truncate(path ...string) (*os.File, error) {
	return create(os.O_TRUNC|os.O_CREATE|os.O_WRONLY, path)
}

func create(flags int, resourcePath []string) (*os.File, error) {
	resource := buildPath(resourcePath)

	// Make sure path to the resource exists
	err := os.MkdirAll(path.Dir(resource), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Attempt to create the resource
	f, err := os.OpenFile(resource, flags, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func buildPath(resourcePath []string) string {
	resource := BASE_DIR
	for _, p := range resourcePath {
		resource = path.Join(resource, p)
	}

	return resource
}
