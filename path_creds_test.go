package atlas

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestDatabaseUser(t *testing.T) {
	var resp *logical.Response
	var err error
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b := NewBackend()
	if err := b.Setup(context.Background(), config); err != nil {
		t.Fatal(err)
	}

	jsonRoles := `[{"databaseName":"admin","roleName":"atlasAdmin"}]`

	compactRoles, err := compactJSON(jsonRoles)
	if err != nil {
		t.Fatalf("error compacting roles %s", err)
	}

	projectID := os.Getenv("ATLAS_PROJECTID")
	if len(projectID) == 0 {
		t.Fatal("ATLAS_PROJECTID not set")
	}

	publicKey := os.Getenv("ATLAS_PUBLICKEY")
	if len(projectID) == 0 {
		t.Fatal("ATLAS_PUBLICKEY not set")
	}

	privateKey := os.Getenv("ATLAS_PRIVATEKEY")
	if len(projectID) == 0 {
		t.Fatal("ATLAS_PRIVATEKEY not set")
	}

	configData := map[string]interface{}{
		"public_key":  publicKey,
		"private_key": privateKey,
	}

	configReq := &logical.Request{
		Operation: logical.UpdateOperation,
		Storage:   config.StorageView,
		Data:      configData,
	}

	configReq.Path = "config/root"
	resp, err = b.HandleRequest(context.Background(), configReq)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: config/root write failed: resp:%#v err:%v", resp, err)
	}

	credData := map[string]interface{}{
		"credential_type": "database_user",
		"database_name":   "admin",
		"project_id":      projectID,
		"roles":           compactRoles,
	}

	credReq := &logical.Request{
		Operation: logical.UpdateOperation,
		Storage:   config.StorageView,
		Data:      credData,
	}

	credReq.Path = "roles/testcred"
	resp, err = b.HandleRequest(context.Background(), credReq)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: credential creation failed: resp:%#v err:%v", resp, err)
	}

	userReq := &logical.Request{
		Operation: logical.ReadOperation,
		Storage:   config.StorageView,
	}
	userReq.Path = "creds/testcred"
	resp, err = b.HandleRequest(context.Background(), userReq)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: read database_user failed: resp:%#v err:%v", resp, err)
	}

}
