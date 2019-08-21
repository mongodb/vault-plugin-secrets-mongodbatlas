---
layout: "api"
page_title: "MongoDB Atlas - Database - Secrets Engines - HTTP API"
sidebar_title: "MongoDB Atlas"
sidebar_current: "api-http-secret-databases-mongodbatlas"
description: |-
  The MongoDB Atlas plugin for Vault's database secrets engine generates database credentials to access MongoDB Atlas databases.
---

# MongoDB Database Plugin HTTP API

The MongoDB Atlas database plugin is one of the supported plugins for the database
secrets engine. This plugin generates database credentials dynamically based on
configured roles for the MongoDB Atlas database.

  ~> **Notice:** The following will be accurate after review and approval by Hashicorp, which is in
    progress. Until then follow the instructions in the [README developing section](./../../../../../README.md).


## Configure Connection

In addition to the parameters defined by the [Database
Backend](/api/secret/databases/index.html#configure-connection), this plugin
has a number of parameters to further configure a connection.

| Method   | Path                         |
| :--------------------------- | :--------------------- |
| `POST`   | `/database/config/:name`     |

### Parameters

- `public_key` `(string: <required>)` – The Public Key used to authenticate with MongoDB Atlas API.
- `private_key` `(string: <required>)` - The Private Key used to connect with MongoDB Atlas API.
- `project_id` `(string: <required>)` - The Project ID to which the database belongs to.

     ~> **Notice:** Do not use your MongoDB Atlas root account credentials.
     Instead generate a dedicated Programmatic API key with appropriate roles.

### Sample Payload

```json
{
  "plugin_name": "mongodbatlas-database-plugin",
  "allowed_roles": "readonly",
  "public_key": "aPublicKey",
  "private_key": "aPrivateKey",
  "project_id": "aProjectID",
}
```

### Sample Request

```
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/v1/database/config/mongodbatlas
```

## Statements

Statements are configured during role creation and are used by the plugin to
determine what is sent to the database on user creation, renewing, and
revocation. For more information on configuring roles see the [Role
API](/api/secret/databases/index.html#create-role) in the database secrets engine docs.

### Parameters

The following are the statements used by this plugin. If not mentioned in this
list the plugin does not support that statement type.

- `creation_statements` `(string: <required>)` – Specifies the database
  statements executed to create and configure a user. Must be a
  serialized JSON object, or a base64-encoded serialized JSON object.
  The object can optionally contain a "database_name", the name of
  the authentication database to log into MongoDB. In Atlas deployments of
  MongoDB, the authentication database is always the admin database. And
  also must contain a "roles" array. This array contains objects that holds
  a series of roles "roleName", an optional "databaseName" and "collectionName"
  value. For more information regarding the `roles` field, refer to
  [MongoDB Atlas documentation](https://docs.atlas.mongodb.com/reference/api/database-users-create-a-user/).


### Sample Creation Statement

```json
{
	"database_name": "admin",
	"roles": [{
		"databaseName": "admin",
		"roleName": "atlasAdmin"
	}]
}
```
