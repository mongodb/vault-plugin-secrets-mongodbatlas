package mongodbatlas

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := NewBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func NewBackend() *Backend {
	var b Backend
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),

		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"config/root",
			},
		},

		Paths: []*framework.Path{
			pathRolesList(&b),
			pathRoles(&b),
			pathConfigRoot(&b),
			pathCredentials(&b),
		},

		Secrets: []*framework.Secret{
			programmaticAPIKeys(&b),
		},

		WALRollbackMinAge: minUserRollbackAge,
		BackendType:       logical.TypeLogical,
	}

	return &b
}

type Backend struct {
	*framework.Backend

	// Mutex to protect access to client and client config
	clientMutex     sync.RWMutex
	credentialMutex sync.RWMutex

	client *mongodbatlas.Client

	logger hclog.Logger
	system logical.SystemView
}

func (b *Backend) Setup(ctx context.Context, config *logical.BackendConfig) error {
	b.logger = config.Logger
	b.system = config.System
	return nil
}

const backendHelp = `
The MongoDB Atlas backend dynamically generates API keys for a set of 
Organization or Project roles. The API keys have a configurable lease 
set and are automatically revoked at the end of the lease.

After mounting this backend, the Public and Private keys to generate 
API keys must be configured with the "config" path and roles must be 
written  using the "roles/" endpoints before any API keys can be generated.

`
const minUserRollbackAge = 5 * time.Minute
