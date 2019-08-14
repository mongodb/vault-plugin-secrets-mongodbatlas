package mongodbatlas

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
				Description: "Name of the Roles",
			},
			"project_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Project ID the %s API key belongs to.", projectProgrammaticAPIKey),
			},
			"roles": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: fmt.Sprintf("List of roles that the API Key should be granted. A minimum of one role must be provided. Any roles provided must be valid for the assigned Project, required for %s and %s keys.", orgProgrammaticAPIKey, projectProgrammaticAPIKey),
			},
			"organization_id": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: fmt.Sprintf("Organization ID required for an %s API key", orgProgrammaticAPIKey),
			},
			"ip_addresses": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: fmt.Sprintf("IP address to be added to the whitelist for the API key. Optional for %s and %s keys.", orgProgrammaticAPIKey, projectProgrammaticAPIKey),
			},
			"cidr_blocks": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: fmt.Sprintf("Whitelist entry in CIDR notation to be added for the API key. Optional for %s and %s keys.", orgProgrammaticAPIKey, projectProgrammaticAPIKey),
			},
			"project_roles": &framework.FieldSchema{
				Type:        framework.TypeCommaStringSlice,
				Description: fmt.Sprintf("Roles assigned when an %s API Key is assiged to a %s API key", orgProgrammaticAPIKey, projectProgrammaticAPIKey),
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

	if organizatioIDRaw, ok := d.GetOk("organization_id"); ok {
		organizatioID := organizatioIDRaw.(string)
		credentialEntry.OrganizationID = organizatioID
	}

	if err = getAPIWhitelistArgs(credentialEntry, d); err != nil {
		resp.AddWarning(fmt.Sprintf("%s", err))
	}

	if projectIDRaw, ok := d.GetOk("project_id"); ok {
		projectID := projectIDRaw.(string)
		credentialEntry.ProjectID = projectID
	}

	if err = getAPIWhitelistArgs(credentialEntry, d); err != nil {
		resp.AddWarning(fmt.Sprintf("%s", err))
	}

	if programmaticKeyRolesRaw, ok := d.GetOk("roles"); ok {
		credentialEntry.Roles = programmaticKeyRolesRaw.([]string)
	}

	if projectRolesRaw, ok := d.GetOk("project_roles"); ok {
		credentialEntry.ProjectRoles = projectRolesRaw.([]string)
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

func getAPIWhitelistArgs(credentialEntry *atlasCredentialEntry, d *framework.FieldData) error {

	if cidrBlocks, ok := d.GetOk("cidr_blocks"); ok {
		credentialEntry.CIDRBlocks = cidrBlocks.([]string)
	}
	if addresses, ok := d.GetOk("ip_addresses"); ok {
		credentialEntry.IPAddresses = addresses.([]string)
	}
	return nil
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
	ProjectID      string        `json:"project_id"`
	DatabaseName   string        `json:"database_name"`
	Roles          []string      `json:"roles"`
	OrganizationID string        `json:"organization_id"`
	CIDRBlocks     []string      `json:"cidr_blocks"`
	IPAddresses    []string      `json:"ip_addresses"`
	ProjectRoles   []string      `json:"project_roles"`
	TTL            time.Duration `json:"ttl"`
	MaxTTL         time.Duration `json:"max_ttl"`
}

func (r atlasCredentialEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"project_id":      r.ProjectID,
		"database_name":   r.DatabaseName,
		"roles":           r.Roles,
		"organization_id": r.OrganizationID,
		"cidr_blocks":     r.CIDRBlocks,
		"ip_addresses":    r.IPAddresses,
		"project_roles":   r.ProjectRoles,
		"ttl":             r.TTL.String(),
		"max_ttl":         r.MaxTTL.String(),
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
const orgProgrammaticAPIKey = `organization`
const projectProgrammaticAPIKey = `project`
const programmaticAPIKey = `programmatic_api_key`
