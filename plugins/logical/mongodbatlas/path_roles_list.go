package mongodbatlas

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathRolesList(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "roles/?$",

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathRolesList,
		},

		HelpSynopsis:    pathRolesListHelpSyn,
		HelpDescription: pathRolesListHelpDesc,
	}
}

func (b *Backend) pathRolesList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.credentialMutex.RLock()
	defer b.credentialMutex.RUnlock()
	entries, err := req.Storage.List(ctx, "roles/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

const pathRolesListHelpSyn = ``
const pathRolesListHelpDesc = ``
