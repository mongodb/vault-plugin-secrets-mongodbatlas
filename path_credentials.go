package atlas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/strutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathListCredentials(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "credentials/?$",

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.pathCredentialList,
		},

		HelpSynopsis:    pathListRolesHelpSyn,
		HelpDescription: pathListRolesHelpDesc,
	}
}

func (b *Backend) pathCredentialList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.credentialMutex.RLock()
	defer b.credentialMutex.RUnlock()
	entries, err := req.Storage.List(ctx, "credentials/")
	if err != nil {
		return nil, err
	}

	if b.logger.IsDebug() {
		b.logger.Debug(fmt.Sprintf("Entries %+v", entries))
	}

	return logical.ListResponse(entries), nil
}

func pathCredentials(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "credentials/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the Credentials",
				DisplayName: "Credential Name",
			},

			"credential_type": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Type of credential to retrieve. Must be one of %s or %s", databaseUser, programmaticAPIKey),
			},
			"project_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Project ID the credential belongs to, required for %s", databaseUser),
			},
			"database_name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Database name the credential belongs to, required for %s", databaseUser),
			},
			"roles": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Roles for the credential, required for %s", databaseUser),
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathCredentialsDelete,
			logical.ReadOperation:   b.pathCredentialsRead,
			logical.UpdateOperation: b.pathCredentialsWrite,
		},

		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

func (b *Backend) pathCredentialsDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "credential/"+d.Get("name").(string))
	return nil, err
}

func (b *Backend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entry, err := b.credentialRead(ctx, req.Storage, d.Get("name").(string), true)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return &logical.Response{
		Data: entry.toResponseData(),
	}, nil
}

func (b *Backend) pathCredentialsWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	var resp logical.Response

	credentialName := d.Get("name").(string)
	if credentialName == "" {
		return logical.ErrorResponse("missing role name"), nil
	}

	b.credentialMutex.Lock()
	defer b.credentialMutex.Unlock()
	credentialEntry, err := b.credentialRead(ctx, req.Storage, credentialName, false)
	if err != nil {
		return nil, err
	}

	if credentialEntry == nil {
		credentialEntry = &atlasCredentialEntry{}
	}

	if credentialTypeRaw, ok := d.GetOk("credential_type"); ok {
		credentialType := credentialTypeRaw.(string)
		allowedCredentialTypes := []string{databaseUser, programmaticAPIKey}
		if credentialType == "" {
			return logical.ErrorResponse("emtpy credential_type"), nil
		}
		if !strutil.StrListContains(allowedCredentialTypes, credentialType) {
			return logical.ErrorResponse(fmt.Sprintf("unrecognized credential_type %q, not one of %#v", credentialType, allowedCredentialTypes)), nil
		}
		credentialEntry.CredentialType = credentialType

		if credentialType == databaseUser {
			if projectIDRaw, ok := d.GetOk("project_id"); ok {
				projectID := projectIDRaw.(string)
				credentialEntry.ProjectID = projectID
			} else {
				resp.AddWarning(fmt.Sprintf("project_id required for %s", databaseUser))
			}

			if databaseNameRaw, ok := d.GetOk("database_name"); ok {
				databaseName := databaseNameRaw.(string)
				credentialEntry.DatabaseName = databaseName
			} else {
				resp.AddWarning(fmt.Sprintf("project_id required for %s", databaseUser))
			}

			if rolesRaw, ok := d.GetOk("roles"); ok {
				compacted := rolesRaw.(string)
				if len(compacted) > 0 {
					compacted, err = compactJSON(rolesRaw.(string))
					if err != nil {
						return logical.ErrorResponse(fmt.Sprintf("cannot parse roles: %q", rolesRaw.(string))), nil
					}
				}
				credentialEntry.Roles = compacted
			}

		}
	}

	err = setAtlasCredential(ctx, req.Storage, credentialName, credentialEntry)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func setAtlasCredential(ctx context.Context, s logical.Storage, credentialName string, credentialEntry *atlasCredentialEntry) error {
	if credentialName == "" {
		return fmt.Errorf("empty role name")
	}
	if credentialEntry == nil {
		return fmt.Errorf("emtpy credentialEntry")
	}
	entry, err := logical.StorageEntryJSON("credential/"+credentialName, credentialEntry)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("nil result when writing to storage")
	}
	if err := s.Put(ctx, entry); err != nil {
		return err
	}
	return nil

}

func (b *Backend) credentialRead(ctx context.Context, s logical.Storage, credentialName string, shouldLock bool) (*atlasCredentialEntry, error) {
	if credentialName == "" {
		return nil, fmt.Errorf("missing credential name")
	}
	if shouldLock {
		b.credentialMutex.RLock()
	}
	entry, err := s.Get(ctx, "credential/"+credentialName)
	if shouldLock {
		b.credentialMutex.RUnlock()
	}
	if err != nil {
		return nil, err
	}
	var credentialEntry atlasCredentialEntry
	if entry != nil {
		if err := entry.DecodeJSON(&credentialEntry); err != nil {
			return nil, err
		}
		return &credentialEntry, nil
	}

	if shouldLock {
		b.credentialMutex.Lock()
		defer b.credentialMutex.Unlock()
	}
	entry, err = s.Get(ctx, "credential/"+credentialName)
	if err != nil {
		return nil, err
	}

	if entry != nil {
		if err := entry.DecodeJSON(&credentialEntry); err != nil {
			return nil, err
		}
		return &credentialEntry, nil
	}
	return &credentialEntry, nil
}

type atlasCredentialEntry struct {
	CredentialType string `json:"credential_type"`
	ProjectID      string `json:"project_id"`
	DatabaseName   string `json:"database_name"`
	Roles          string `json:"roles"`
}

func (r atlasCredentialEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"credential_type": r.CredentialType,
		"project_id":      r.ProjectID,
		"database_name":   r.DatabaseName,
		"roles":           r.Roles,
	}
	return respData
}

func compactJSON(input string) (string, error) {
	var compacted bytes.Buffer
	err := json.Compact(&compacted, []byte(input))
	return compacted.String(), err
}

const pathListRolesHelpSyn = ``
const pathListRolesHelpDesc = ``
const pathCredentialsHelpSyn = ``
const pathCredentialsHelpDesc = ``
const databaseUser = `database_user`
const programmaticAPIKey = `programmatic_api_key`
