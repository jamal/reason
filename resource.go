package reason

import "errors"

// ErrNotFound should be returned when a resource cannot be found, will cause the
// server to return http.StatusNotFound.
var ErrNotFound = errors.New("Resource not found")

// ResourceHandler does thingz
type ResourceHandler interface {
	Path() string
}

// Getter implementers will expose a GET method to fetch a specific resource.
type Getter interface {
	GetResource(resourceID string) (interface{}, error)
}

// Lister implementers will expose a GET method to fetch a list of that
// resource.
type Lister interface {
	ListResource() ([]interface{}, error)
}

// Creator implementers will expose a POST method to create a new resource.
type Creator interface {
	CreateResource(resource interface{}) (interface{}, error)
}

// Updater implementers will expose a POST/PUT method to update a single
// resource.
type Updater interface {
	Getter
	UpdateResource(resource interface{}, data interface{}) (interface{}, error)
}

// Deleter implements will expose a DELETE method to delete a single resource.
type Deleter interface {
	Getter
	DeleteResource(resource interface{}) error
}
