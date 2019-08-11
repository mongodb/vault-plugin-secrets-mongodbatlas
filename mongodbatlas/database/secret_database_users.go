package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	"github.com/sethvargo/go-password/password"
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
	// Get the lease (if any)
	leaseConfig, err := b.LeaseConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if leaseConfig == nil {
		leaseConfig = &configLease{}
	}

	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = leaseConfig.TTL
	resp.Secret.MaxTTL = leaseConfig.MaxTTL
	return resp, nil
}

func (b *Backend) pathDatabaseUserRollback(ctx context.Context, req *logical.Request, _kind string, data interface{}) error {

	var entry walEntry
	if err := mapstructure.Decode(data, &entry); err != nil {
		return err
	}
	username := entry.UserName
	projectID := entry.ProjectID

	// Get the client
	client, err := b.clientMongo(ctx, req.Storage)
	if err != nil {
		return nil
	}

	// check if the user exists or not
	_, res, err := client.DatabaseUsers.Get(context.Background(), projectID, username)
	// if the user is gone, move along
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	// now, delete the user
	res, err = client.DatabaseUsers.Delete(context.Background(), projectID, username)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	return nil
}

func (b *Backend) databaseUserRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// Get the username from the internal data
	usernameRaw, ok := req.Secret.InternalData["username"]
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}

	username, ok := usernameRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}
	projectIDRaw, ok := req.Secret.InternalData["projectid"]
	if !ok {
		return nil, fmt.Errorf("secret is missing projectid internal data")
	}

	projectID, ok := projectIDRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing projectid internal data")
	}
	// Use the user rollback mechanism to delete this database_user
	err := b.pathDatabaseUserRollback(ctx, req, "database_user", map[string]interface{}{
		"username":  username,
		"projectid": projectID,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *Backend) databaseUserCreate(ctx context.Context, s logical.Storage, displayName string, cred *atlasCredentialEntry, lease *configLease) (*logical.Response, error) {

	username := genUsername(displayName)
	client, err := b.clientMongo(ctx, s)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	walID, err := framework.PutWAL(ctx, s, "database_user", &walEntry{
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
	passwd, err := password.Generate(22, 3, 0, false, false)
	if err != nil {
		return nil, err
	}

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
		"username":  username,
		"password":  passwd,
		"projectid": cred.ProjectID,
	})

	resp.Secret.TTL = lease.TTL
	resp.Secret.MaxTTL = lease.MaxTTL

	return resp, nil
}
