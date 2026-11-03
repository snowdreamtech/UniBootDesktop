// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	defaultSecretService = "github.com/snowdreamtech/unigodesktop"
	proxyPasswordAccount = "proxy-password"
)

type secretStore interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

type keyringStore struct{}

func (keyringStore) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (keyringStore) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (keyringStore) Delete(service, user string) error {
	return keyring.Delete(service, user)
}

var globalSecretStore secretStore = keyringStore{}

// SaveSecret stores an arbitrary secret for the given account key in the OS credential store.
func SaveSecret(account, secret string) error {
	if secret == "" {
		return DeleteSecret(account)
	}
	if err := globalSecretStore.Set(defaultSecretService, account, secret); err != nil {
		return fmt.Errorf("save secret %q to system credential store: %w", account, err)
	}
	return nil
}

// LoadSecret retrieves a secret for the given account key from the OS credential store.
func LoadSecret(account string) (string, error) {
	secret, err := globalSecretStore.Get(defaultSecretService, account)
	if err != nil {
		return "", fmt.Errorf("load secret %q from system credential store: %w", account, err)
	}
	return secret, nil
}

// DeleteSecret removes a secret for the given account key from the OS credential store.
func DeleteSecret(account string) error {
	if err := globalSecretStore.Delete(defaultSecretService, account); err != nil {
		return fmt.Errorf("delete secret %q from system credential store: %w", account, err)
	}
	return nil
}

// SaveProxyPassword stores the network proxy password in the operating system credential store.
func SaveProxyPassword(password string) error {
	return SaveSecret(proxyPasswordAccount, password)
}

// LoadProxyPassword retrieves the network proxy password from the operating system credential store.
func LoadProxyPassword() (string, error) {
	return LoadSecret(proxyPasswordAccount)
}

// DeleteProxyPassword removes the network proxy password from the operating system credential store.
func DeleteProxyPassword() error {
	return DeleteSecret(proxyPasswordAccount)
}
