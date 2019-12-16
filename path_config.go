package mongodbatlas

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathConfig(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
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
			logical.UpdateOperation: b.pathConfigWrite,
		},
		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *Backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	entry, err := logical.StorageEntryJSON("config", config{
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

type config struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

const pathConfigHelpSyn = `
Configure the  credentials that are used to manage Database Users.
`

const pathConfigHelpDesc = `
Before doing anything, the Atlas backend needs credentials that are able
to manage databaseusers, access keys, etc. This endpoint is used to 
configure those credentials.
`
