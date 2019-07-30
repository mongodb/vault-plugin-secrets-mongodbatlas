package atlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb-partners/go-client-mongodb-atlas/mongodbatlas"
)

func programmaticAPIKeys(b *Backend) *framework.Secret {
	return &framework.Secret{
		Type: programmaticAPIKey,
		Fields: map[string]*framework.FieldSchema{
			"public_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Programmatic API Key Public Key",
			},

			"private_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Programmatic API Key Private Key",
			},
			"security_token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Security Token",
			},
		},

		Renew:  b.databaseUserRenew,
		Revoke: b.programmaticAPIKeyRevoke,
	}
}

func (b *Backend) programmaticAPIKeyCreate(ctx context.Context, s logical.Storage, displayName string, cred *atlasCredentialEntry, lease *configLease) (*logical.Response, error) {

	username := genUsername(displayName)
	client, err := b.clientMongo(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	walID, err := framework.PutWAL(ctx, s, databaseUser, &walDatabaseUser{
		UserName: username,
	})
	if err != nil {
		return nil, errwrap.Wrapf("error writing WAL entry: {{err}}", err)
	}

	key, _, err := client.APIKeys.Create(context.Background(), cred.OrganizationID, &mongodbatlas.APIKeyInput{
		Desc:  username,
		Roles: cred.ProgrammaticKeyRoles,
	})

	if err != nil {
		if walErr := framework.DeleteWAL(ctx, s, walID); walErr != nil {
			dbUserErr := errwrap.Wrapf("error creating programmaticAPIKey user: {{err}}", err)
			return nil, errwrap.Wrap(errwrap.Wrapf("failed to delete WAL entry: {{err}}", walErr), dbUserErr)
		}
		return logical.ErrorResponse(fmt.Sprintf(
			"Error creating programmatic api key user user: %s", err)), err
	}

	if err := framework.DeleteWAL(ctx, s, walID); err != nil {
		return nil, errwrap.Wrapf("failed to commit WAL entry: {{err}}", err)
	}

	resp := b.Secret(programmaticAPIKey).Response(map[string]interface{}{
		"public_key":  key.PublicKey,
		"private_key": key.PrivateKey,
		"description": username,
	}, map[string]interface{}{
		"programmaticapikeyid": key.ID,
		"projectid":            cred.ProjectID,
		"organizationid":       cred.OrganizationID,
	})

	resp.Secret.TTL = lease.TTL
	resp.Secret.MaxTTL = lease.MaxTTL

	return resp, nil
}

func (b *Backend) programmaticAPIKeyRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	programmaticAPIKeyIDRaw, ok := req.Secret.InternalData["programmaticapikeyid"]
	if !ok {
		return nil, fmt.Errorf("secret is missing programmatic api key id internal data")
	}

	programmaticAPIKeyID, ok := programmaticAPIKeyIDRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing programmatic api key id internal data")
	}
	organizationIDRaw, ok := req.Secret.InternalData["organizationid"]
	if !ok {
		return nil, fmt.Errorf("secret is missing organization id internal data")
	}

	organizationID, ok := organizationIDRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing organization id internal data")
	}

	// Use the user rollback mechanism to delete this database_user
	err := b.pathProgrammaticAPIKeyRollback(ctx, req, programmaticAPIKey, map[string]interface{}{
		"organizationid":       organizationID,
		"programmaticapikeyid": programmaticAPIKeyID,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *Backend) pathProgrammaticAPIKeyRollback(ctx context.Context, req *logical.Request, _kind string, data interface{}) error {

	var entry walDatabaseUser
	if err := mapstructure.Decode(data, &entry); err != nil {
		return err
	}
	organizationID := entry.OrganizationID
	programmaticAPIKeyID := entry.ProgrammaticAPIKeyID

	// Get the client
	client, err := b.clientMongo(ctx, req.Storage)
	if err != nil {
		return nil
	}

	// check if the user exists or not
	_, res, err := client.APIKeys.Get(context.Background(), organizationID, programmaticAPIKeyID)
	// if the user is gone, move along
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	// now, delete the user
	res, err = client.APIKeys.Delete(context.Background(), organizationID, programmaticAPIKeyID)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	return nil
}
