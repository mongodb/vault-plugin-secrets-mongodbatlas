package atlas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mongodb-partners/go-client-mongodb-atlas/mongodbatlas"
)

func databaseUsers(b *Backend) *framework.Secret {
	return &framework.Secret{
		Type: databaseUser,
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Username",
			},

			"password": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Password",
			},
			"security_token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Security Token",
			},
		},

		Renew:  b.databaseUserRenew,
		Revoke: b.databaseUserRevoke,
	}
}

func (b *Backend) databaseUserRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return nil, nil
}

func (b *Backend) databaseUserRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return nil, nil
}

func (b *Backend) databaseUserCreate(ctx context.Context, s logical.Storage, displayName string, cred *atlasCredentialEntry) (*logical.Response, error) {

	username := genUsername(displayName)
	client, err := b.clientMongo(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	walID, err := framework.PutWAL(ctx, s, "database_user", &walDatabaseUser{
		UserName: username,
	})
	if err != nil {
		return nil, errwrap.Wrapf("error writing WAL entry: {{err}}", err)
	}

	var roles []mongodbatlas.Role

	err = json.Unmarshal([]byte(cred.Roles), &roles)
	if err != nil {
		return nil, errwrap.Wrapf("error reading credential roles {{err}}", err)
	}
	passwd := getRandomPassword(22)

	_, _, err = client.DatabaseUsers.Create(context.Background(), cred.ProjectID, &mongodbatlas.DatabaseUser{
		Username:     username,
		Password:     passwd,
		GroupID:      cred.ProjectID,
		DatabaseName: cred.DatabaseName,
		Roles:        roles,
	})

	if err != nil {
		if walErr := framework.DeleteWAL(ctx, s, walID); walErr != nil {
			dbUserErr := errwrap.Wrapf("error creating databaseUser user: {{err}}", err)
			return nil, errwrap.Wrap(errwrap.Wrapf("failed to delete WAL entry: {{err}}", walErr), dbUserErr)
		}
		return logical.ErrorResponse(fmt.Sprintf(
			"Error creating database user user: %s", err)), err
	}

	if err := framework.DeleteWAL(ctx, s, walID); err != nil {
		return nil, errwrap.Wrapf("failed to commit WAL entry: {{err}}", err)
	}

	resp := b.Secret(databaseUser).Response(map[string]interface{}{
		"username": username,
		"password": passwd,
	}, map[string]interface{}{
		"username": username,
		"password": passwd,
	})

	return resp, nil
}
