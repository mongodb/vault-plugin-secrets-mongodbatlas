package mongodbatlas

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/base62"
	"github.com/hashicorp/vault/sdk/logical"
)

var displayNameRegex = regexp.MustCompile("[^a-zA-Z0-9+=,.@_-]")

func pathCredentials(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathCredentialsRead,
			logical.UpdateOperation: b.pathCredentialsRead,
		},

		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}

}

func (b *Backend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	userName := d.Get("name").(string)

	cred, err := b.credentialRead(ctx, req.Storage, userName, true)
	if err != nil {
		return nil, errwrap.Wrapf("error retrieving credential: {{err}}", err)
	}

	if cred == nil {
		return nil, errors.New("error retrieving credential: credential is nil")
	}

	return b.programmaticAPIKeyCreate(ctx, req.Storage, userName, cred)

}

type walEntry struct {
	UserName             string
	ProjectID            string
	OrganizationID       string
	ProgrammaticAPIKeyID string
}

func genUsername(displayName string) (string, error) {

	midString := displayNameRegex.ReplaceAllString(displayName, "_")

	id, err := base62.Random(20)
	if err != nil {
		return "", err
	}
	ret := fmt.Sprintf("vault-%s%s-%d", midString, id, rand.Int31n(10000))
	return ret, nil
}

const pathCredentialsHelpSyn = `
Generate MongoDB Atlas Programmatic API from a specific Vault role.
`
const pathCredentialsHelpDesc = `
This path reads generates MongoDB Atlas Programmatic API Keys for
a particular role. Atlas Programmatic API Keys will be
generated on demand and will be automatically revoked when
the lease is up.
`
