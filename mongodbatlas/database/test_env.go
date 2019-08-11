package database

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

type testEnv struct {
	PublicKey  string
	PrivateKey string
	ProjectID  string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	MostRecentSecret *logical.Secret
}

func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config/root",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"public_key":  e.PublicKey,
			"private_key": e.PrivateKey,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) AddRole(t *testing.T) {
	roles := `[{"databaseName":"admin","roleName":"atlasAdmin"}]`
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-credential",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"project_id":    e.ProjectID,
			"database_name": "admin",
			"roles":         roles,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	// if resp != nil {
	// 	t.Fatal("expected nil response to represent a 204")
	// }
}

func (e *testEnv) AddRoleWithTTL(t *testing.T) {
	roles := `[{"databaseName":"admin","roleName":"atlasAdmin"}]`
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-credential",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"project_id":    e.ProjectID,
			"database_name": "admin",
			"roles":         roles,
			"ttl":           2000,
			"max_ttl":       4000,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	// if resp != nil {
	// 	t.Fatal("expected nil response to represent a 204")
	// }
}

func (e *testEnv) ReadDatabaseUserCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-credential",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}

	if resp.Data["username"] == "" {
		t.Fatal("failed to receive access_key")
	}
	if resp.Data["password"] == "" {
		t.Fatal("failed to receive secret_key")
	}
	e.MostRecentSecret = resp.Secret
}

func (e *testEnv) RenewDatabaseUserCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RenewOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if resp.Secret != e.MostRecentSecret {
		t.Fatalf("expected %+v but got %+v", e.MostRecentSecret, resp.Secret)
	}
}

func (e *testEnv) RevokeDatabaseUsersCreds(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RevokeOperation,
		Storage:   e.Storage,
		Secret:    e.MostRecentSecret,
		Data: map[string]interface{}{
			"lease_id": "foo",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}
