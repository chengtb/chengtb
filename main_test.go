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
	if len(got) != 2 {
		t.Fatalf("unexpected params count: %d", len(got))
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

func TestValidateInputMissingOutput(t *testing.T) {
	err := validateInput(inputConfig{
		imagePath: "./main.go",
		rawParams: []string{"name=zhangsan"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "missing required -output") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateInputMissingParam(t *testing.T) {
	err := validateInput(inputConfig{
		imagePath: "./main.go",
		outputPDF: "./out.pdf",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "missing required -param") {
		t.Fatalf("unexpected error: %v", err)
	}
}
