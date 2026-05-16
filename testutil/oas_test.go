package testutil_test

import (
	"context"
	"testing"

	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func TestAssertSuccessEnvelope_Valid(t *testing.T) {
	body := []byte(`{"success":true,"data":{"id":1}}`)
	env := testutil.AssertSuccessEnvelope(t, body)
	data, ok := env["data"].(map[string]any)
	if !ok || data["id"] == nil {
		t.Fatal("expected data.id to be present")
	}
}

func TestAssertSuccessEnvelope_MissingData(t *testing.T) {
	ft := &fakeT{}
	testutil.AssertSuccessEnvelope(ft, []byte(`{"success":true}`))
	if !ft.failed {
		t.Fatal("expected failure when data key is absent")
	}
}

func TestAssertErrorEnvelope_Valid(t *testing.T) {
	body := []byte(`{"success":false,"error":{"code":"NOT_FOUND","message":"not found"}}`)
	testutil.AssertErrorEnvelope(t, body)
}

func TestAssertErrorEnvelope_WrongSuccess(t *testing.T) {
	ft := &fakeT{}
	testutil.AssertErrorEnvelope(ft, []byte(`{"success":true,"error":{"code":"X","message":"y"}}`))
	if !ft.failed {
		t.Fatal("expected failure when success is true in error envelope")
	}
}

// fakeT is a minimal testing.TB stub that records failures without stopping.
type fakeT struct {
	testing.TB
	failed bool
}

func (f *fakeT) Helper()               {}
func (f *fakeT) Errorf(string, ...any) { f.failed = true }
func (f *fakeT) Fatalf(string, ...any) { f.failed = true }
func (f *fakeT) Log(...any)            {}
func (f *fakeT) Logf(string, ...any)   {}
func (f *fakeT) Fail()                 { f.failed = true }
func (f *fakeT) FailNow()              { f.failed = true }
func (f *fakeT) Failed() bool          { return f.failed }
func (f *fakeT) Skip(...any)           {}
func (f *fakeT) Skipf(string, ...any)  {}
func (f *fakeT) Name() string          { return "fakeT" }
func (f *fakeT) Cleanup(func())        {}
func (f *fakeT) TempDir() string       { return "" }
func (f *fakeT) Setenv(string, string) {}
func (f *fakeT) Chdir(string)          {}
func (f *fakeT) Context() context.Context {
	return context.Background()
}
