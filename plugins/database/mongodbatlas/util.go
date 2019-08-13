package mongodbatlas

import "github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"

type mongoDBAtlasStatement struct {
	ProjectID    string              `json:"project_id"`
	DatabaseName string              `json:"database_name"`
	Roles        []mongodbatlas.Role `json:"roles,omitempty"`
}
