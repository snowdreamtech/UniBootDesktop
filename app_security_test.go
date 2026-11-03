// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !nogui

package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivateOrLocalIP(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{host: "127.0.0.1", want: true},
		{host: "::1", want: true},
		{host: "10.0.0.1", want: true},
		{host: "192.168.1.1", want: true},
		{host: "172.16.5.4", want: true},
		{host: "169.254.1.1", want: true},
		{host: "0.0.0.0", want: true},
		{host: "localhost", want: true},
		{host: "foo.local", want: true},
		{host: "fc00::1", want: true},
		{host: "8.8.8.8", want: false},
		{host: "example.com", want: false},
		{host: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.want, isPrivateOrLocalIP(tt.host))
		})
	}
}

func TestOpenBrowserURLRejectsPrivateIPv4(t *testing.T) {
	app := &App{}
	err := app.OpenBrowserURL("https://10.1.2.3/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "private/local"))
}

func TestOpenBrowserURLRejectsNonHTTPS(t *testing.T) {
	app := &App{}
	err := app.OpenBrowserURL("file:///etc/passwd")
	assert.Error(t, err)
}
