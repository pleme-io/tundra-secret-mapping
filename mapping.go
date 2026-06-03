// Package secretmapping is the fleet's one typed JSON-secret field-mapping —
// so no integration adapter (ServiceNow, ESO, a UI rotator) forks its own
// "pull these fields out of the secret JSON" logic. A Mapping declares which
// JSON keys of a secret map to canonical credential fields (username, password,
// private key, passphrase, …), with per-secret-path overrides.
//
// It is a pure, generic field-mapping primitive: it models any JSON secret and
// names no particular project. The mapping itself is the public shape; the
// concrete per-secret mapping values a deployment uses are its own data.
package secretmapping

import (
	"encoding/json"
	"sort"

	errs "github.com/pleme-io/errors-go"
)

// Field is a canonical credential field name an integration consumes. New
// fields are additive.
type Field string

// Canonical credential fields.
const (
	Username   Field = "username"
	Password   Field = "password"
	PrivateKey Field = "private_key"
	Passphrase Field = "passphrase"
)

// FieldMap maps a canonical Field to the JSON key in the secret that supplies
// it. E.g. {Username: "user", Password: "pass"} reads secret["user"] into the
// canonical "username" slot.
type FieldMap map[Field]string

// Mapping is a typed JSON-secret field-mapping with per-secret overrides. The
// Default map applies to every secret; Overrides keyed by secret path replace
// the default for that path. The zero value (no default, no overrides) is valid
// but maps nothing; build it through New.
type Mapping struct {
	// Default maps canonical fields to JSON keys for all secrets.
	Default FieldMap
	// Overrides maps a secret path to a FieldMap that replaces the Default for
	// that path (per-field: an override entry shadows the default entry).
	Overrides map[string]FieldMap
}

// Option configures a Mapping.
type Option func(*Mapping)

// WithDefault sets the default field map.
func WithDefault(fm FieldMap) Option { return func(m *Mapping) { m.Default = fm } }

// WithOverride sets a per-secret-path override field map. Override entries
// shadow the corresponding default entries for that path.
func WithOverride(path string, fm FieldMap) Option {
	return func(m *Mapping) {
		if m.Overrides == nil {
			m.Overrides = map[string]FieldMap{}
		}
		m.Overrides[path] = fm
	}
}

// New builds a Mapping from options.
func New(opts ...Option) *Mapping {
	m := &Mapping{Default: FieldMap{}}
	for _, o := range opts {
		o(m)
	}
	if m.Default == nil {
		m.Default = FieldMap{}
	}
	return m
}

// Validate reports the first structural problem with the mapping as a typed,
// code-carrying error. A valid mapping has non-empty JSON keys for every entry.
func (m *Mapping) Validate() error {
	if err := validateFieldMap("default", m.Default); err != nil {
		return err
	}
	// Sort override paths for deterministic error reporting.
	paths := make([]string, 0, len(m.Overrides))
	for p := range m.Overrides {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if err := validateFieldMap("override "+p, m.Overrides[p]); err != nil {
			return err
		}
	}
	return nil
}

func validateFieldMap(scope string, fm FieldMap) error {
	for f, key := range fm {
		if f == "" {
			return errs.New("secretmapping: "+scope+" has an empty canonical field", errs.WithCode("mapping_empty_field"))
		}
		if key == "" {
			return errs.New("secretmapping: "+scope+" field "+string(f)+" maps to an empty json key", errs.WithCode("mapping_empty_key"))
		}
	}
	return nil
}

// effective returns the field map to use for a given secret path: the Default
// merged with any per-path Overrides (override entries win per-field).
func (m *Mapping) effective(path string) FieldMap {
	out := FieldMap{}
	for f, k := range m.Default {
		out[f] = k
	}
	for f, k := range m.Overrides[path] {
		out[f] = k
	}
	return out
}

// Apply reads the raw JSON secret bytes and projects them onto the canonical
// fields, using the mapping for the given secret path. It returns a map keyed by
// canonical Field. A canonical field whose mapped JSON key is absent from the
// secret is omitted from the result (not an error) — so callers can detect
// missing optional fields.
func (m *Mapping) Apply(path string, secretJSON []byte) (map[Field]string, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(secretJSON, &raw); err != nil {
		return nil, errs.Wrap(err, "secretmapping: secret is not a JSON object", errs.WithCode("mapping_bad_secret"))
	}
	fm := m.effective(path)
	out := map[Field]string{}
	for field, key := range fm {
		v, ok := raw[key]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			return nil, errs.New("secretmapping: secret key "+key+" is not a string", errs.WithCode("mapping_not_string"))
		}
		out[field] = s
	}
	return out, nil
}
