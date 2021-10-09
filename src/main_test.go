package main

import (
	"testing"

	"github.com/rendon/testcli"
)

func TestNoArgs(t *testing.T) {
	testcli.Run("snapkup")
	if testcli.Success() {
		t.Fatal("App without arguments should fail")
	}
}

func TestVersion(t *testing.T) {
	testcli.Run("snapkup", "--version")
	if !testcli.Success() {
		t.Fatal("App with --version argument should NOT fail")
	}

	if !testcli.StderrContains("____") {
		t.Fatal("Expected banner")
	}
}

func TestHelp(t *testing.T) {
	testcli.Run("snapkup", "--help")
	if !testcli.Success() {
		t.Fatal("App with --help argument should NOT fail")
	}

	if !testcli.StderrContains("usage") {
		t.Fatal("Expected help text")
	}
}

func TestInit(t *testing.T) {
	testcli.Run("snapkup", "--help")
	if !testcli.Success() {
		t.Fatal("App with --help argument should NOT fail")
	}

	if !testcli.StderrContains("usage") {
		t.Fatal("Expected help text")
	}
}
