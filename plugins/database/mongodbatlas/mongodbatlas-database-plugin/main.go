package main

import (
	"log"
	"os"

	"github.com/mongodb/vault-plugin-secrets-mongodbatlas/plugins/database/mongodbatlas"

	"github.com/hashicorp/vault/api"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	err := mongodbatlas.Run(apiClientMeta.GetTLSConfig())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
