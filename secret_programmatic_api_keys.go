package atlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
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

	apiKeyDescription := genUsername(displayName)
	client, err := b.clientMongo(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	walID, err := framework.PutWAL(ctx, s, programmaticAPIKey, &walDatabaseUser{
		UserName: apiKeyDescription,
	})
	if err != nil {
		return nil, errwrap.Wrapf("error writing WAL entry: {{err}}", err)
	}

	var key *mongodbatlas.APIKey

	switch cred.CredentialType {
	case orgProgrammaticAPIKey:
		key, _, err = client.APIKeys.Create(context.Background(), cred.OrganizationID,
			&mongodbatlas.APIKeyInput{
				Desc:  apiKeyDescription,
				Roles: cred.ProgrammaticKeyRoles,
			})
		if err == nil {
			err = addWhitelistEntry(client, cred.OrganizationID, key.ID, cred)
		}
	case projectProgrammaticAPIKey:
		key, _, err = client.ProjectAPIKeys.Create(context.Background(), cred.ProjectID,
			&mongodbatlas.APIKeyInput{
				Desc:  apiKeyDescription,
				Roles: cred.ProgrammaticKeyRoles,
			})
		if err == nil {
			err = addWhitelistEntry(client, key.Roles[0].OrgID, key.ID, cred)
		}
	}

	if err != nil {
		if walErr := framework.DeleteWAL(ctx, s, walID); walErr != nil {
			dbUserErr := errwrap.Wrapf("error creating programmaticAPIKey: {{err}}", err)
			return nil, errwrap.Wrap(errwrap.Wrapf("failed to delete WAL entry: {{err}}", walErr), dbUserErr)
		}
		return logical.ErrorResponse(fmt.Sprintf(
			"Error creating programmatic api key: %s", err)), err
	}

	if err := framework.DeleteWAL(ctx, s, walID); err != nil {
		return nil, errwrap.Wrapf("failed to commit WAL entry: {{err}}", err)
	}

	resp := b.Secret(programmaticAPIKey).Response(map[string]interface{}{
		"public_key":  key.PublicKey,
		"private_key": key.PrivateKey,
		"description": apiKeyDescription,
	}, map[string]interface{}{
		"programmaticapikeyid": key.ID,
		"projectid":            cred.ProjectID,
		"organizationid":       cred.OrganizationID,
		"credentialtype":       cred.CredentialType,
	})

	resp.Secret.TTL = lease.TTL
	resp.Secret.MaxTTL = lease.MaxTTL

	return resp, nil
}

func addWhitelistEntry(client *mongodbatlas.Client, orgID string, keyID string, cred *atlasCredentialEntry) error {
	if len(cred.CIDRBlocks) > 0 {
		cidrBlocks := make([]*mongodbatlas.WhitelistAPIKeysReq, len(cred.CIDRBlocks))
		for i, cidrBlock := range cred.CIDRBlocks {
			cidrBlocks[i].CidrBlock = cidrBlock
		}
		_, _, err := client.WhitelistAPIKeys.Create(context.Background(), orgID, keyID, cidrBlocks)
		if err != nil {
			return err
		}
	}

	if len(cred.IPAddresses) > 0 {
		ipAddresses := make([]*mongodbatlas.WhitelistAPIKeysReq, len(cred.IPAddresses))
		for i, ipAddress := range cred.IPAddresses {
			ipAddresses[i].IPAddress = ipAddress
		}
		_, _, err := client.WhitelistAPIKeys.Create(context.Background(), orgID, keyID, ipAddresses)
		if err != nil {
			return err
		}
	}

	return nil
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
	credentialTypeRaw, ok := req.Secret.InternalData["credentialtype"]
	if !ok {
		return nil, fmt.Errorf("secret is missing credential type internal data")
	}

	credentialType, ok := credentialTypeRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing credential type internal internal data")
	}

	var organizationID string
	var projectID string
	var data map[string]interface{}

	switch credentialType {
	case orgProgrammaticAPIKey:
		organizationIDRaw, ok := req.Secret.InternalData["organizationid"]
		if !ok {
			return nil, fmt.Errorf("secret is missing organization id internal data")
		}

		organizationID, ok = organizationIDRaw.(string)
		if !ok {
			return nil, fmt.Errorf("secret is missing organization id internal data")
		}

		data = map[string]interface{}{
			"organizationid":       organizationID,
			"programmaticapikeyid": programmaticAPIKeyID,
			"credentialtype":       credentialType,
		}

	case programmaticAPIKeyID:
		projectIDRaw, ok := req.Secret.InternalData["projectid"]
		if !ok {
			return nil, fmt.Errorf("secret is missing project id internal data")
		}

		projectID, ok = projectIDRaw.(string)
		if !ok {
			return nil, fmt.Errorf("secret is missing project id internal data")
		}

		data = map[string]interface{}{
			"projectid":            projectID,
			"programmaticapikeyid": programmaticAPIKeyID,
			"credentialtype":       credentialType,
		}
	case programmaticAPIKeyID:
		projectIDRaw, ok := req.Secret.InternalData["projectid"]
		if !ok {
			return nil, fmt.Errorf("secret is missing project id internal data")
		}

		projectID, ok = projectIDRaw.(string)
		if !ok {
			return nil, fmt.Errorf("secret is missing project id internal data")
		}

		data = map[string]interface{}{
			"projectid":            projectID,
			"programmaticapikeyid": programmaticAPIKeyID,
			"credentialtype":       credentialType,
		}
	}
	// Use the user rollback mechanism to delete this database_user
	err := b.pathProgrammaticAPIKeyRollback(ctx, req, programmaticAPIKey, data)
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

	// Get the client
	client, err := b.clientMongo(ctx, req.Storage)
	if err != nil {
		return nil
	}

	switch entry.CredentialType {
	case orgProgrammaticAPIKey:
		// check if the user exists or not
		_, res, err := client.APIKeys.Get(context.Background(), entry.OrganizationID, entry.ProgrammaticAPIKeyID)
		// if the user is gone, move along
		if err != nil {
			if res != nil && res.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}

		// now, delete the api key
		res, err = client.APIKeys.Delete(context.Background(), entry.OrganizationID, entry.ProgrammaticAPIKeyID)
		if err != nil {
			if res != nil && res.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
	case projectProgrammaticAPIKey:
		// now, delete the user
		res, err := client.ProjectAPIKeys.Unassign(context.Background(), entry.ProjectID, entry.ProgrammaticAPIKeyID)
		if err != nil {
			if res != nil && res.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
	}

	return nil
}
