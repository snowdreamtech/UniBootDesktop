// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"sync"
	"testing"
)

type memorySecretStore struct {
	mu      sync.RWMutex
	secrets map[string]string
}

func newMemorySecretStore() *memorySecretStore {
	return &memorySecretStore{
		secrets: make(map[string]string),
	}
}

func (m *memorySecretStore) key(service, user string) string {
	return service + "::" + user
}

func (m *memorySecretStore) Get(service, user string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.secrets[m.key(service, user)]
	if !ok || val == "" {
		return "", errors.New("secret not found")
	}
	return val, nil
}

func (m *memorySecretStore) Set(service, user, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[m.key(service, user)] = password
	return nil
}

func (m *memorySecretStore) Delete(service, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(service, user)
	if _, ok := m.secrets[k]; !ok {
		return errors.New("secret not found")
	}
	delete(m.secrets, k)
	return nil
}

func TestCredentialsStore(t *testing.T) {
	prev := globalSecretStore
	store := newMemorySecretStore()
	globalSecretStore = store
	t.Cleanup(func() { globalSecretStore = prev })

	// Test generic Secret helpers
	if err := SaveSecret("custom-token", "tok_123456"); err != nil {
		t.Fatalf("SaveSecret failed: %v", err)
	}
	val, err := LoadSecret("custom-token")
	if err != nil {
		t.Fatalf("LoadSecret failed: %v", err)
	}
	if val != "tok_123456" {
		t.Fatalf("expected tok_123456, got %s", val)
	}

	if err := DeleteSecret("custom-token"); err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}
	if _, err := LoadSecret("custom-token"); err == nil {
		t.Fatal("expected error after DeleteSecret, got nil")
	}

	// Test ProxyPassword convenience helpers
	if err := SaveProxyPassword("super-secret-password"); err != nil {
		t.Fatalf("SaveProxyPassword failed: %v", err)
	}
	pwd, err := LoadProxyPassword()
	if err != nil {
		t.Fatalf("LoadProxyPassword failed: %v", err)
	}
	if pwd != "super-secret-password" {
		t.Fatalf("expected super-secret-password, got %s", pwd)
	}

	// Saving empty password should delete it
	if err := SaveProxyPassword(""); err != nil {
		t.Fatalf("SaveProxyPassword('') failed: %v", err)
	}
	if _, err := LoadProxyPassword(); err == nil {
		t.Fatal("expected empty password to trigger deletion")
	}
}
