package database

import (
	"context"
	"errors"

	"github.com/Sectorbob/mlab-ns2/gae/ns/digest"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
)

func (b *Backend) clientMongo(ctx context.Context, s logical.Storage) (*mongodbatlas.Client, error) {
	b.clientMutex.RLock()
	if b.client != nil {
		b.clientMutex.RUnlock()
		return b.client, nil
	}

	// Upgrade the lock for writing
	b.clientMutex.RUnlock()
	b.clientMutex.Lock()
	defer b.clientMutex.Unlock()

	// check client again, in the event that a client was being created while we
	// waited for Lock()
	if b.client != nil {
		return b.client, nil
	}

	client, err := nonCachedClient(ctx, s)
	if err != nil {
		return nil, err
	}
	b.client = client

	return b.client, nil
}

func nonCachedClient(ctx context.Context, s logical.Storage) (*mongodbatlas.Client, error) {
	transport, err := getRootConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	client, err := transport.Client()
	if err != nil {
		return nil, err
	}

	return mongodbatlas.NewClient(client), nil
}

func getRootConfig(ctx context.Context, s logical.Storage) (*digest.Transport, error) {

	entry, err := s.Get(ctx, "config/root")
	if err != nil {
		return nil, err
	}
	if entry != nil {
		var config rootConfig
		if err := entry.DecodeJSON(&config); err != nil {
			return nil, errwrap.Wrapf("error reading root configuration: {{err}}", err)
		}

		transport := digest.NewTransport(config.PublicKey, config.PrivateKey)

		// return the transport we are done
		return transport, nil

	}

	return nil, errors.New("Empty config/root entry")
}
