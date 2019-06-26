package atlas

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestBackend_PathListCredentials(t *testing.T) {
	var resp *logical.Response
	var err error
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b := NewBackend()
	if err := b.Setup(context.Background(), config); err != nil {
		t.Fatal(err)
	}

	credData := map[string]interface{}{
		"credential_type": "database_user",
	}

	credReq := &logical.Request{
		Operation: logical.UpdateOperation,
		Storage:   config.StorageView,
		Data:      credData,
	}

	for i := 1; i <= 10; i++ {
		credReq.Path = "credentials/testcred" + strconv.Itoa(i)
		resp, err = b.HandleRequest(context.Background(), credReq)
		if err != nil || (resp != nil && resp.IsError()) {
			t.Fatalf("bad: credential creation failed:. resp:%#v err:%v", resp, err)
		}

	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      "credentials/",
		Storage:   config.StorageView,
	})
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: listing credentials failed. resp:%#v\n err:%v", resp, err)
	}

	if len(resp.Data["keys"].([]string)) != 10 {
		t.Fatalf("failed to list all 10 credentials")
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      "credential/",
		Storage:   config.StorageView,
	})
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: listing credentials failed. resp:%#v\n err:%v", resp, err)
	}

	if len(resp.Data["keys"].([]string)) != 10 {
		t.Fatalf("failed to list all 10 credentials")
	}
}
