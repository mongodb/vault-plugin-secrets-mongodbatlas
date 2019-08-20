package mongodbatlas

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Sectorbob/mlab-ns2/gae/ns/digest"
	"github.com/hashicorp/vault/sdk/database/dbplugin"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
)

const envVarRunAccTests = "VAULT_ACC"

const testMongoDBAtlasRole = `{"roles": [{"databaseName":"admin","roleName":"atlasAdmin"}]}`

var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

func TestIntegrationDatabaseUser_Initialize(t *testing.T) {
	connectionDetails := map[string]interface{}{
		"public_key":  "aspergesme",
		"private_key": "domine",
	}
	db := new()

	_, err := db.Init(context.Background(), connectionDetails, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !db.Initialized {
		t.Fatal("Database should be initialized")
	}
}

func TestAcceptanceDatabaseUser_CreateUser(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")
	projectID := os.Getenv("ATLAS_PROJECT_ID")

	connectionDetails := map[string]interface{}{
		"public_key":  publicKey,
		"private_key": privateKey,
		"project_id":  projectID,
	}

	db := new()
	_, err := db.Init(context.Background(), connectionDetails, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	statements := dbplugin.Statements{
		Creation: []string{testMongoDBAtlasRole},
	}

	usernameConfig := dbplugin.UsernameConfig{
		DisplayName: "test",
		RoleName:    "test",
	}

	username, _, err := db.CreateUser(context.Background(), statements, usernameConfig, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := testCredsExists(projectID, publicKey, privateKey, username); err != nil {
		t.Fatalf("Credentials were not created: %s", err)
	}

	if err := deleteCredentials(projectID, publicKey, privateKey, username); err != nil {
		t.Fatalf("Credentials could not be deleted: %s", err)
	}

}

func TestAcceptanceDatabaseUser_RevokeUser(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")
	projectID := os.Getenv("ATLAS_PROJECT_ID")

	connectionDetails := map[string]interface{}{
		"public_key":  publicKey,
		"private_key": privateKey,
		"project_id":  projectID,
	}

	db := new()
	_, err := db.Init(context.Background(), connectionDetails, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	statements := dbplugin.Statements{
		Creation: []string{testMongoDBAtlasRole},
	}

	usernameConfig := dbplugin.UsernameConfig{
		DisplayName: "test",
		RoleName:    "test",
	}

	username, _, err := db.CreateUser(context.Background(), statements, usernameConfig, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test default revocation statement
	err = db.RevokeUser(context.Background(), statements, username)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAcceptanceDatabaseUser_SetCredentials(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	publicKey := os.Getenv("ATLAS_PUBLIC_KEY")
	privateKey := os.Getenv("ATLAS_PRIVATE_KEY")
	projectID := os.Getenv("ATLAS_PROJECT_ID")

	connectionDetails := map[string]interface{}{
		"public_key":  publicKey,
		"private_key": privateKey,
		"project_id":  projectID,
	}

	db := new()
	_, err := db.Init(context.Background(), connectionDetails, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// create the database user in advance, and test the connection
	dbUser := "testmongouser"
	startingPassword := "3>^chcBo7a7t-ZI"

	testCreateAtlasDBUser(t, projectID, publicKey, privateKey, dbUser, startingPassword)
	if err := testCredsExists(projectID, publicKey, privateKey, dbUser); err != nil {
		t.Fatalf("Could not connect with new credentials: %s", err)
	}

	newPassword, err := db.GenerateCredentials(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	usernameConfig := dbplugin.StaticUserConfig{
		Username: dbUser,
		Password: newPassword,
	}

	statements := dbplugin.Statements{
		Creation: []string{testMongoDBAtlasRole},
	}

	username, password, err := db.SetCredentials(context.Background(), statements, usernameConfig)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := testCredsExists(projectID, publicKey, privateKey, username); err != nil {
		t.Fatalf("Could not connect with new credentials: %s", err)
	}
	// confirm the original creds used to set still work (should be the same)
	if err := testCredsExists(projectID, publicKey, privateKey, username); err != nil {
		t.Fatalf("Could not connect with new credentials: %s", err)
	}

	if (dbUser != username) || (newPassword != password) {
		t.Fatalf("username/password mismatch: (%s)/(%s) vs (%s)/(%s)", dbUser, username, newPassword, password)
	}

	if err := deleteCredentials(projectID, publicKey, privateKey, dbUser); err != nil {
		t.Fatalf("Credentials could not be deleted: %s", err)
	}
}

func testCreateAtlasDBUser(t testing.TB, projectID, publicKey, privateKey, username, startingPassword string) {
	client, err := getClient(publicKey, privateKey)
	if err != nil {
		t.Fatalf("Error creating client %s", err)
	}

	databaseUserRequest := &mongodbatlas.DatabaseUser{
		Username:     username,
		Password:     startingPassword,
		DatabaseName: "admin",
		Roles: []mongodbatlas.Role{
			{
				DatabaseName: "admin",
				RoleName:     "atlasAdmin",
			},
		},
	}

	_, _, err = client.DatabaseUsers.Create(context.Background(), projectID, databaseUserRequest)
	if err != nil {
		t.Fatalf("Error Creating User %s", err)
	}

}

func testCredsExists(projectID, publicKey, privateKey, username string) error {
	client, err := getClient(publicKey, privateKey)
	if err != nil {
		return err
	}

	_, _, err = client.DatabaseUsers.Get(context.Background(), projectID, username)
	if err != nil {
		return err
	}

	return err
}

func deleteCredentials(projectID, publicKey, privateKey, username string) error {
	client, err := getClient(publicKey, privateKey)
	if err != nil {
		return err
	}
	_, err = client.DatabaseUsers.Delete(context.Background(), projectID, username)

	return err
}

func getClient(publicKey, privateKey string) (*mongodbatlas.Client, error) {
	transport := digest.NewTransport(publicKey, privateKey)
	cl, err := transport.Client()
	if err != nil {
		return nil, err
	}

	return mongodbatlas.New(cl)

}
