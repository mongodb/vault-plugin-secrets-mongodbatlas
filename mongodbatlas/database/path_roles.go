package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathRoles(b *Backend) *framework.Path {
	return &framework.Path{
		Pattern: "roles/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the Credentials",
				DisplayName: "Credential Name",
			},
			"project_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Project ID the credential belongs to, required for MongoDB Atlas %s", databaseUser),
			},
			"database_name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Database name the credential belongs to, required for MongoDB Atlas %s", databaseUser),
			},
			"roles": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Roles for the credential, required for MongoDB Atlas %s", databaseUser),
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: `Duration in seconds after which the issued token should expire. Defaults to 0, in which case the value will fallback to the system/mount defaults.`,
			},
			"max_ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "The maximum allowed lifetime of tokens issued using this role.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathRolesDelete,
			logical.ReadOperation:   b.pathRolesRead,
			logical.UpdateOperation: b.pathRolesWrite,
		},

		HelpSynopsis:    pathRolesHelpSyn,
		HelpDescription: pathRolesHelpDesc,
	}
}

func (b *Backend) pathRolesDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "roles/"+d.Get("name").(string))
	return nil, err
}

func (b *Backend) pathRolesRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
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

func (b *Backend) pathRolesWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
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

	if databaseNameRaw, ok := d.GetOk("database_name"); ok {
		databaseName := databaseNameRaw.(string)
		credentialEntry.DatabaseName = databaseName
	} else {
		resp.AddWarning(fmt.Sprintf("database_name required for %s", databaseUser))
	}

	if projectIDRaw, ok := d.GetOk("project_id"); ok {
		projectID := projectIDRaw.(string)
		credentialEntry.ProjectID = projectID
	} else {
		resp.AddWarning(fmt.Sprintf("project_id required for %s ", databaseUser))
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

	if ttlRaw, ok := d.GetOk("ttl"); ok {
		credentialEntry.TTL = time.Duration(ttlRaw.(int)) * time.Second
	}

	if maxttlRaw, ok := d.GetOk("max_ttl"); ok {
		credentialEntry.MaxTTL = time.Duration(maxttlRaw.(int)) * time.Second
	}

	if credentialEntry.MaxTTL > 0 && credentialEntry.TTL > credentialEntry.MaxTTL {
		return logical.ErrorResponse("ttl exceeds max_ttl"), nil
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
	entry, err := logical.StorageEntryJSON("roles/"+credentialName, credentialEntry)
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
	entry, err := s.Get(ctx, "roles/"+credentialName)
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
	entry, err = s.Get(ctx, "roles/"+credentialName)
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
	ProjectID    string        `json:"project_id"`
	DatabaseName string        `json:"database_name"`
	Roles        string        `json:"roles"`
	TTL          time.Duration `json:"ttl"`
	MaxTTL       time.Duration `json:"max_ttl"`
}

func (r atlasCredentialEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"project_id":    r.ProjectID,
		"database_name": r.DatabaseName,
		"roles":         r.Roles,
		"ttl":           r.TTL.String(),
		"max_ttl":       r.MaxTTL.String(),
	}
	return respData
}

func compactJSON(input string) (string, error) {
	var compacted bytes.Buffer
	err := json.Compact(&compacted, []byte(input))
	return compacted.String(), err
}

const pathRolesHelpSyn = ``
const pathRolesHelpDesc = ``
const databaseUser = `database_user`
