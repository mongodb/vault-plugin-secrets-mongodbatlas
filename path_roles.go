package atlas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/strutil"
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

			"credential_type": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Type of credential to retrieve. Must be one of %s, %s or %s", databaseUser, orgProgrammaticAPIKey, projectProgrammaticAPIKey),
			},
			"project_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Project ID the credential belongs to, required for %s or %s", databaseUser, projectProgrammaticAPIKey),
			},
			"database_name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Database name the credential belongs to, required for %s", databaseUser),
			},
			"roles": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Roles for the credential, required for %s", databaseUser),
			},
			"programmatic_key_roles": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: fmt.Sprintf("Roles for a programmatic API key, required for %s or %s", orgProgrammaticAPIKey, projectProgrammaticAPIKey),
			},
			"organization_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Organization ID for the credential, required for %s", orgProgrammaticAPIKey),
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
		// Create for withelist
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

	if credentialTypeRaw, ok := d.GetOk("credential_type"); ok {
		credentialType := credentialTypeRaw.(string)
		allowedCredentialTypes := []string{databaseUser, orgProgrammaticAPIKey, projectProgrammaticAPIKey}
		if credentialType == "" {
			return logical.ErrorResponse("emtpy credential_type"), nil
		}
		if !strutil.StrListContains(allowedCredentialTypes, credentialType) {
			return logical.ErrorResponse(fmt.Sprintf("unrecognized credential_type %q, not one of %#v", credentialType, allowedCredentialTypes)), nil
		}
		credentialEntry.CredentialType = credentialType

		switch credentialType {
		case databaseUser:

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

		case orgProgrammaticAPIKey:
			if programmaticKeyRolesRaw, ok := d.GetOk("programmatic_key_roles"); ok {
				credentialEntry.ProgrammaticKeyRoles = programmaticKeyRolesRaw.([]string)
			} else {
				resp.AddWarning(fmt.Sprintf("programmatic_key_roles required for %s", orgProgrammaticAPIKey))
			}
			if organizatioIDRaw, ok := d.GetOk("organization_id"); ok {
				organizatioID := organizatioIDRaw.(string)
				credentialEntry.OrganizationID = organizatioID
			} else {
				resp.AddWarning(fmt.Sprintf("organization_id required for %s", orgProgrammaticAPIKey))
			}

		case projectProgrammaticAPIKey:
			if programmaticKeyRolesRaw, ok := d.GetOk("programmatic_key_roles"); ok {
				credentialEntry.ProgrammaticKeyRoles = programmaticKeyRolesRaw.([]string)
			} else {
				resp.AddWarning(fmt.Sprintf("programmatic_key_roles required for %s", orgProgrammaticAPIKey))
			}
			if projectIDRaw, ok := d.GetOk("project_id"); ok {
				projectID := projectIDRaw.(string)
				credentialEntry.ProjectID = projectID
			} else {
				resp.AddWarning(fmt.Sprintf("project_id required for %s ", databaseUser))
			}

		default:
			return logical.ErrorResponse("Unsupported credential_type %s", credentialType), nil
		}
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
	CredentialType       string        `json:"credential_type"`
	ProjectID            string        `json:"project_id"`
	DatabaseName         string        `json:"database_name"`
	Roles                string        `json:"roles"`
	ProgrammaticKeyRoles []string      `json:"programmatic_key_roles"`
	OrganizationID       string        `json:"organization_id"`
	TTL                  time.Duration `json:"ttl"`
	MaxTTL               time.Duration `json:"max_ttl"`
}

func (r atlasCredentialEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"credential_type":        r.CredentialType,
		"project_id":             r.ProjectID,
		"database_name":          r.DatabaseName,
		"roles":                  r.Roles,
		"programmatic_key_roles": r.ProgrammaticKeyRoles,
		"organization_id":        r.OrganizationID,
		"ttl":                    r.TTL.String(),
		"max_ttl":                r.MaxTTL.String(),
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
const orgProgrammaticAPIKey = `org_programmatic_api_key`
const projectProgrammaticAPIKey = `project_programmatic_api_key`
const programmaticAPIKey = `programmatic_api_key`
