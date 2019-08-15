package mongodbatlas

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

func pathCredentials(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the user",
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

	defaultLease, err := b.leaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	// Get lease configuration
	leaseConfig := &configLease{
		TTL:    0,
		MaxTTL: 0,
	}

	if cred.TTL > 0 {
		leaseConfig.TTL = cred.TTL
	} else if defaultLease != nil {
		leaseConfig.TTL = defaultLease.TTL
	}

	if cred.MaxTTL > 0 {
		leaseConfig.MaxTTL = cred.MaxTTL
	} else if defaultLease != nil {
		leaseConfig.MaxTTL = defaultLease.MaxTTL
	}

	if leaseConfig.TTL > leaseConfig.MaxTTL {
		leaseConfig.TTL = leaseConfig.MaxTTL
	}

	return b.programmaticAPIKeyCreate(ctx, req.Storage, userName, cred, leaseConfig)

}

type walEntry struct {
	UserName             string
	ProjectID            string
	OrganizationID       string
	ProgrammaticAPIKeyID string
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

const pathCredentialsHelpSyn = `
Generate MongoDB Atlas Programmatic API from a specific Vault role.
`
const pathCredentialsHelpDesc = `
This path reads generates MongoDB Atlas Programmatic API Keys for
a particular role. Atlas Programmatic API Keys will be
generated on demand and will be automatically revoked when
the lease is up.
`
