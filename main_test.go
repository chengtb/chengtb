package main

import (
	"strings"
	"testing"
)

func TestParseParams(t *testing.T) {
	got, err := parseParams([]string{"name=zhangsan", "order_no=A10001"})
	if err != nil {
		t.Fatalf("parseParams returned error: %v", err)
	}
	if got["name"] != "zhangsan" {
		t.Fatalf("unexpected name: %q", got["name"])
	}
	if got["order_no"] != "A10001" {
		t.Fatalf("unexpected order_no: %q", got["order_no"])
	}
}

func TestParseParamsInvalid(t *testing.T) {
	_, err := parseParams([]string{"invalid"})
	if err == nil {
		t.Fatal("expected error for invalid param")
	}
	if !strings.Contains(err.Error(), "expected key=value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateInputMissingRequired(t *testing.T) {
	err := validateInput(inputConfig{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "missing required -image") {
		t.Fatalf("unexpected error: %v", err)
	}
}
