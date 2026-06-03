package secretmapping

import (
	"testing"

	errs "github.com/pleme-io/errors-go"
)

func TestApplyDefault(t *testing.T) {
	m := New(WithDefault(FieldMap{Username: "user", Password: "pass"}))
	got, err := m.Apply("/db/prod", []byte(`{"user":"alice","pass":"s3cret","extra":"x"}`))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got[Username] != "alice" {
		t.Fatalf("username = %q", got[Username])
	}
	if got[Password] != "s3cret" {
		t.Fatalf("password = %q", got[Password])
	}
	if len(got) != 2 {
		t.Fatalf("got %d fields, want 2: %v", len(got), got)
	}
}

func TestApplyOverrideShadowsPerField(t *testing.T) {
	m := New(
		WithDefault(FieldMap{Username: "user", Password: "pass"}),
		// only override the username key for this specific path; password keeps default.
		WithOverride("/ssh/host", FieldMap{Username: "login", PrivateKey: "key"}),
	)
	got, err := m.Apply("/ssh/host", []byte(`{"login":"root","pass":"p","key":"PEM"}`))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got[Username] != "root" { // from override key "login"
		t.Fatalf("username = %q, want root", got[Username])
	}
	if got[Password] != "p" { // from default key "pass"
		t.Fatalf("password = %q, want p", got[Password])
	}
	if got[PrivateKey] != "PEM" { // from override key "key"
		t.Fatalf("private_key = %q, want PEM", got[PrivateKey])
	}
}

func TestApplyOtherPathUnaffectedByOverride(t *testing.T) {
	m := New(
		WithDefault(FieldMap{Username: "user"}),
		WithOverride("/ssh/host", FieldMap{Username: "login"}),
	)
	got, err := m.Apply("/db/prod", []byte(`{"user":"alice","login":"ignored"}`))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got[Username] != "alice" {
		t.Fatalf("username = %q, want alice (default applies to non-overridden path)", got[Username])
	}
}

func TestApplyMissingKeyOmitted(t *testing.T) {
	m := New(WithDefault(FieldMap{Username: "user", Passphrase: "phrase"}))
	got, err := m.Apply("/x", []byte(`{"user":"alice"}`))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, ok := got[Passphrase]; ok {
		t.Fatalf("passphrase should be omitted when absent, got %q", got[Passphrase])
	}
	if got[Username] != "alice" {
		t.Fatalf("username = %q", got[Username])
	}
}

func TestApplyErrors(t *testing.T) {
	tests := []struct {
		name       string
		mapping    *Mapping
		path       string
		secretJSON string
		code       string
	}{
		{
			name:       "non-object secret",
			mapping:    New(WithDefault(FieldMap{Username: "user"})),
			path:       "/x",
			secretJSON: `["not","an","object"]`,
			code:       "mapping_bad_secret",
		},
		{
			name:       "garbage json",
			mapping:    New(WithDefault(FieldMap{Username: "user"})),
			path:       "/x",
			secretJSON: `{not json`,
			code:       "mapping_bad_secret",
		},
		{
			name:       "non-string value",
			mapping:    New(WithDefault(FieldMap{Username: "user"})),
			path:       "/x",
			secretJSON: `{"user": 123}`,
			code:       "mapping_not_string",
		},
		{
			name:       "empty key in map",
			mapping:    New(WithDefault(FieldMap{Username: ""})),
			path:       "/x",
			secretJSON: `{"user":"a"}`,
			code:       "mapping_empty_key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.mapping.Apply(tt.path, []byte(tt.secretJSON))
			if err == nil {
				t.Fatalf("want code %q, got nil", tt.code)
			}
			if errs.CodeOf(err) != tt.code {
				t.Fatalf("code = %q, want %q", errs.CodeOf(err), tt.code)
			}
		})
	}
}

func TestValidateOK(t *testing.T) {
	m := New(
		WithDefault(FieldMap{Username: "user", Password: "pass"}),
		WithOverride("/p", FieldMap{PrivateKey: "key"}),
	)
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestNewZeroValue(t *testing.T) {
	m := New()
	got, err := m.Apply("/x", []byte(`{"a":"b"}`))
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("empty mapping should map nothing, got %v", got)
	}
}
