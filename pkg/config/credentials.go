// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	proxyPasswordService = "github.com/snowdreamtech/unigodesktop"
	proxyPasswordAccount = "proxy-password"
)

type secretStore interface {
	Get(service string, user string) (string, error)
	Set(service string, user string, password string) error
	Delete(service string, user string) error
}

type keyringStore struct{}

func (keyringStore) Get(service string, user string) (string, error) {
	return keyring.Get(service, user)
}

func (keyringStore) Set(service string, user string, password string) error {
	return keyring.Set(service, user, password)
}

func (keyringStore) Delete(service string, user string) error {
	return keyring.Delete(service, user)
}

var proxyPasswordStore secretStore = keyringStore{}

// SaveProxyPassword stores the proxy password in the operating system credential store.
func SaveProxyPassword(password string) error {
	if password == "" {
		return DeleteProxyPassword()
	}
	if err := proxyPasswordStore.Set(proxyPasswordService, proxyPasswordAccount, password); err != nil {
		return fmt.Errorf("save proxy password to system credential store: %w", err)
	}
	return nil
}

// LoadProxyPassword retrieves the proxy password from the operating system credential store.
func LoadProxyPassword() (string, error) {
	password, err := proxyPasswordStore.Get(proxyPasswordService, proxyPasswordAccount)
	if err != nil {
		return "", fmt.Errorf("load proxy password from system credential store: %w", err)
	}
	return password, nil
}

// DeleteProxyPassword removes the proxy password from the operating system credential store.
func DeleteProxyPassword() error {
	if err := proxyPasswordStore.Delete(proxyPasswordService, proxyPasswordAccount); err != nil {
		return fmt.Errorf("delete proxy password from system credential store: %w", err)
	}
	return nil
}
