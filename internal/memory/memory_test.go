package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vitualizz/ai-software-delivery-team/internal/memory"
)

func TestNullProvider_Load_ReturnsNilNil(t *testing.T) {
	p := memory.NullProvider{}
	data, err := p.Load(context.Background(), "any-key")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if data != nil {
		t.Errorf("Load: expected nil data, got %v", data)
	}
}

func TestNullProvider_Save_ReturnsNil(t *testing.T) {
	p := memory.NullProvider{}
	if err := p.Save(context.Background(), "any-key", []byte("data")); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}
}

func TestNullProvider_Name(t *testing.T) {
	p := memory.NullProvider{}
	if got := p.Name(); got != "null" {
		t.Errorf("Name: got %q, want %q", got, "null")
	}
}

func TestNullProvider_NoPanicOnEmptyKeyOrNilContext(t *testing.T) {
	p := memory.NullProvider{}

	// nil context — provider must not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Load panicked with nil context: %v", r)
			}
		}()
		//nolint:staticcheck // deliberately passing nil to test robustness
		_, _ = p.Load(nil, "")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Save panicked with nil context: %v", r)
			}
		}()
		//nolint:staticcheck
		_ = p.Save(nil, "", nil)
	}()
}

func TestEngramProvider_Load_ReturnsErrNotImplemented(t *testing.T) {
	p := memory.NewEngramProvider("http://localhost:8080", "test-project")
	_, err := p.Load(context.Background(), "some-key")
	if !errors.Is(err, memory.ErrNotImplemented) {
		t.Errorf("Load: expected ErrNotImplemented, got: %v", err)
	}
}

func TestEngramProvider_Save_ReturnsErrNotImplemented(t *testing.T) {
	p := memory.NewEngramProvider("http://localhost:8080", "test-project")
	err := p.Save(context.Background(), "some-key", []byte("data"))
	if !errors.Is(err, memory.ErrNotImplemented) {
		t.Errorf("Save: expected ErrNotImplemented, got: %v", err)
	}
}

func TestEngramProvider_Name(t *testing.T) {
	p := memory.NewEngramProvider("", "")
	if got := p.Name(); got != "engram" {
		t.Errorf("Name: got %q, want %q", got, "engram")
	}
}
