package database

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathConfigRoot(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/root",
		Fields: map[string]*framework.FieldSchema{
			"public_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "MongoDB Atlas Programmatic Public Key",
			},
			"private_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "MongoDB Atlas Programmatic Private Key",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathConfigRootWrite,
		},
		HelpSynopsis:    pathConfigRootHelpSyn,
		HelpDescription: pathConfigRootHelpDesc,
	}
}

func (b *Backend) pathConfigRootWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	b.clientMutex.Lock()
	defer b.clientMutex.Unlock()

	entry, err := logical.StorageEntryJSON("config/root", rootConfig{
		PublicKey:  data.Get("public_key").(string),
		PrivateKey: data.Get("private_key").(string),
	})
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// Clean cached client (if any)
	b.client = nil

	return nil, nil
}

type rootConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

const pathConfigRootHelpSyn = `
Configure the root credentials that are used to manage Database Users.
`

const pathConfigRootHelpDesc = `
Before doing anything, the Atlas backend needs credentials that are able
to manage databaseusers, access keys, etc. This endpoint is used to 
configure those credentials.
`
