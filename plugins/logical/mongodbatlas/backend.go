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

//Factory ...
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := NewBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// NewBackend ...
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
			pathConfigLease(&b),
		},

		Secrets: []*framework.Secret{
			programmaticAPIKeys(&b),
		},

		WALRollbackMinAge: minUserRollbackAge,
		BackendType:       logical.TypeLogical,
	}

	return &b
}

// Backend ...
type Backend struct {
	*framework.Backend

	// Mutex to protect access to client and client config
	clientMutex     sync.RWMutex
	credentialMutex sync.RWMutex

	client *mongodbatlas.Client

	logger hclog.Logger
	system logical.SystemView
}

// Setup ...
func (b *Backend) Setup(ctx context.Context, config *logical.BackendConfig) error {
	b.logger = config.Logger
	b.system = config.System
	return nil
}

// LeaseConfig returns the lease configuration
func (b *Backend) LeaseConfig(ctx context.Context, s logical.Storage) (*configLease, error) {
	entry, err := s.Get(ctx, "config/lease")
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result configLease
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

const backendHelp = ``
const minUserRollbackAge = 5 * time.Minute
