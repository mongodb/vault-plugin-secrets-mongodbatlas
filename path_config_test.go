package mongodbatlas

import (
	"context"
	"testing"

	"github.com/go-test/deep"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestBackend_PathConfig(t *testing.T) {
	var resp *logical.Response
	var err error
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	b := NewBackend()
	if err := b.Setup(context.Background(), config); err != nil {
		t.Fatal(err)
	}

	// Test write operation
	configData := map[string]interface{}{
		"public_key":  "my_public_key",
		"private_key": "my_private_key",
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      configData,
		Storage:   config.StorageView,
	})

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("config write failed:. resp:%#v err:%v", resp, err)
	}

	// Test read operation
	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Data:      configData,
		Storage:   config.StorageView,
	})

	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("config write failed:. resp:%#v err:%v", resp, err)
	}

	expected := map[string]interface{}{
		"public_key": "my_public_key",
	}

	if diff := deep.Equal(expected, resp.Data); diff != nil {
		t.Fatalf("bad response. expected %v, got: %v", expected, resp.Data)
	}

	// Test bad data on write

	// Missing public key
	configData = map[string]interface{}{
		"private_key": "my_private_key",
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      configData,
		Storage:   config.StorageView,
	})

	if err == nil {
		t.Fatal("expect error response but got nil")
	}

	// Missing private key
	configData = map[string]interface{}{
		"public_key": "my_public_key",
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      configData,
		Storage:   config.StorageView,
	})

	if err == nil {
		t.Fatal("expect error response but got nil")
	}
}
