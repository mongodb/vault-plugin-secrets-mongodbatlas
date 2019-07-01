package atlas

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathDatabaseUser(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the user",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathDatabaseUserRead,
			logical.UpdateOperation: b.pathDatabaseUserRead,
		},

		HelpSynopsis:    pathDatabaseUserHelpSyn,
		HelpDescription: pathDatabaseUserHelpDesc,
	}

}

func (b *Backend) pathDatabaseUserRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	userName := d.Get("name").(string)

	cred, err := b.credentialRead(ctx, req.Storage, userName, true)
	if err != nil {
		return nil, errwrap.Wrapf("error retrieving credential: {{err}}", err)
	}

	// Get lease configuration(if any)
	leaseConfig, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if leaseConfig == nil {
		leaseConfig = &configLease{}
	}

	switch cred.CredentialType {
	case databaseUser:
		return b.databaseUserCreate(ctx, req.Storage, userName, cred, leaseConfig)
	case programmaticAPIKey:
		return nil, nil
	}

	return nil, nil
}

type walDatabaseUser struct {
	UserName  string
	ProjectID string
}

func genUsername(displayName string) (ret string) {
	midString := fmt.Sprintf("%s-",
		normalizeDisplayName(displayName))
	ret = fmt.Sprintf("vault-%s%d-%d", midString, time.Now().Unix(), rand.Int31n(10000))
	return
}

func normalizeDisplayName(displayName string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9+=,.@_-]")
	return re.ReplaceAllString(displayName, "_")

}

const pathDatabaseUserHelpSyn = ``
const pathDatabaseUserHelpDesc = ``
