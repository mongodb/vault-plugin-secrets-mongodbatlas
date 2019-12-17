package mongodbatlas

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
	envVarRunAccTests    = "VAULT_ACC"
	envVarPrivateKey     = "ATLAS_PRIVATE_KEY"
	envVarPublicKey      = "ATLAS_PUBLIC_KEY"
	envVarProjectID      = "ATLAS_PROJECT_ID"
	envVarOrganizationID = "ATLAS_ORGANIZATION_ID"
)

var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

func TestAcceptanceProgrammaticAPIKey(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRole)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_WithProjectID(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRoleWithProjectID)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_ProjectWithIPWhitelist(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRoleProjectWithIP)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_WithIPWhitelist(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRoleWithIP)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_WithCIDRWhitelist(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRoleWithCIDR)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_AssignToProject(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role", acceptanceTestEnv.AddProgrammaticAPIKeyRoleWithProjectIDAndOrgID)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func TestAcceptanceProgrammaticAPIKey_WithTTL(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add programmatic API Key role with TTL", acceptanceTestEnv.AddProgrammaticAPIKeyRoleWithTTL)
	t.Run("read programmatic API key cred", acceptanceTestEnv.ReadProgrammaticAPIKeyRule)
	t.Run("check lease for programmatic API key cred", acceptanceTestEnv.CheckLease)
	t.Run("renew programmatic API key creds", acceptanceTestEnv.RenewProgrammaticAPIKeys)
	t.Run("revoke programmatic API key creds", acceptanceTestEnv.RevokeProgrammaticAPIKeys)

}

func newAcceptanceTestEnv() (*testEnv, error) {
	ctx := context.Background()

	maxLease, _ := time.ParseDuration("60s")
	defaultLease, _ := time.ParseDuration("30s")
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLease,
			MaxLeaseTTLVal:     maxLease,
		},
		Logger: logging.NewVaultLogger(log.Debug),
	}
	b, err := Factory(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &testEnv{
		PublicKey:      os.Getenv(envVarPublicKey),
		PrivateKey:     os.Getenv(envVarPrivateKey),
		ProjectID:      os.Getenv(envVarProjectID),
		OrganizationID: os.Getenv(envVarOrganizationID),
		Backend:        b,
		Context:        ctx,
		Storage:        &logical.InmemStorage{},
	}, nil
}
