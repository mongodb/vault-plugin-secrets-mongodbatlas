package database

import (
	"context"
	"os"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	envVarRunAccTests = "VAULT_ACC"
	envVarPrivateKey  = "ATLAS_PRIVATE_KEY"
	envVarPublicKey   = "ATLAS_PUBLIC_KEY"
	envVarProjectID   = "ATLAS_PROJECT_ID"
)

var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

func TestAcceptanceDatabaseUser(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add role", acceptanceTestEnv.AddRole)
	t.Run("read database user creds", acceptanceTestEnv.ReadDatabaseUserCreds)
	t.Run("renew database user creds", acceptanceTestEnv.RenewDatabaseUserCreds)
	t.Run("revoke database user creds", acceptanceTestEnv.RevokeDatabaseUsersCreds)
}

func newAcceptanceTestEnv() (*testEnv, error) {
	ctx := context.Background()
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: time.Hour,
			MaxLeaseTTLVal:     time.Hour,
		},
		Logger: logging.NewVaultLogger(log.Debug),
	}
	b, err := Factory(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &testEnv{
		PublicKey:  os.Getenv(envVarPublicKey),
		PrivateKey: os.Getenv(envVarPrivateKey),
		ProjectID:  os.Getenv(envVarProjectID),
		Backend:    b,
		Context:    ctx,
		Storage:    &logical.InmemStorage{},
	}, nil
}
