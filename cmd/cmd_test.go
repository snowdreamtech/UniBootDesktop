// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running --help, got: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("expected non-empty output for --help")
	}
}

func TestVersionCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running version cmd, got: %v", err)
	}
}

func TestVersionFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running --version flag, got: %v", err)
	}
}

func TestEnvCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"env"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running env cmd, got: %v", err)
	}
}

func TestDfCmdJSON(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"df", "--json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running df --json, got: %v", err)
	}
}

func TestDoctorCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"doctor"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running doctor cmd, got: %v", err)
	}
}

func TestVmCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"vm", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running vm --help, got: %v", err)
	}
}

func TestConfigCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running config --help, got: %v", err)
	}
}

func TestCacheCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"cache", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running cache --help, got: %v", err)
	}
}

func TestCompletionCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"completion", "bash"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error running completion bash, got: %v", err)
	}
}
