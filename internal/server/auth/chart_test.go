package auth

import (
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/yaml"
)

// The chart renders the document this package parses, and nothing in the Go build ties the two together: a key the
// template adds that the parser does not know makes every server pod exit at startup, and on an upgrade only at the
// next restart. These tests are that tie. CI keeps helm installed so the render check runs there, and diffs the
// rendered bytes against testdata/oidc.full.yaml in the chart job.
const (
	chartDir   = "../../../charts/resourcerer"
	schemaPath = "../../../charts/resourcerer/values.schema.json"
	valuesFile = "testdata/chart-values.yaml"
	goldenFile = "testdata/oidc.full.yaml"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

// TestFullConfigParses reads the checked-in document with the server's own parser and checks the values, so the file
// stays a real configuration and not just something that happens to parse.
func TestFullConfigParses(t *testing.T) {
	c, err := ParseConfig(string(readTestdata(t, goldenFile)))
	if err != nil {
		t.Fatalf("%s: %v", goldenFile, err)
	}
	if c.Issuer != "https://sso.example.com/realms/example" || c.ClientID != "resourcerer" ||
		c.RedirectURL != "https://resourcerer.example.com/auth/callback" || c.TTL() != 8*time.Hour {
		t.Errorf("parsed %+v", c)
	}
	if c.UsernameClaim != "preferred_username" || c.GroupsClaim != "groups" {
		t.Errorf("claims = %q, %q", c.UsernameClaim, c.GroupsClaim)
	}
	if !slices.Equal(c.Scopes, []string{"openid", "profile", "email", "groups"}) {
		t.Errorf("scopes = %v", c.Scopes)
	}
	if !slices.Equal(c.AllowedGroups, []string{"engineering", "yes", "sre"}) {
		t.Errorf("allowedGroups = %q, want the quoted yes to come back as a string", c.AllowedGroups)
	}
	if !slices.Equal(c.Roles.EditorGroups, []string{"engineering"}) || c.Roles.anyViewer() {
		t.Errorf("roles = %+v, want editors engineering and no read access for the rest", c.Roles)
	}
}

// TestFullConfigCoversEveryKey keeps the checked-in document and the parser's fields the same set: a field added to
// Config has to appear here, which is what pushes it into the chart too. A key the chart writes that the parser does
// not know fails the other way, in TestFullConfigParses, which reads strictly.
func TestFullConfigCoversEveryKey(t *testing.T) {
	doc := map[string]any{}
	if err := yaml.Unmarshal(readTestdata(t, goldenFile), &doc); err != nil {
		t.Fatal(err)
	}
	roles, _ := doc["roles"].(map[string]any)

	if got, want := keysOf(doc), jsonFields(Config{}); !slices.Equal(got, want) {
		t.Errorf("document keys = %v, the parser has %v: keep %s, %s and the chart in step", got, want, goldenFile, valuesFile)
	}
	if got, want := keysOf(roles), jsonFields(Roles{}); !slices.Equal(got, want) {
		t.Errorf("roles keys = %v, the parser has %v: keep %s, %s and the chart in step", got, want, goldenFile, valuesFile)
	}
}

// jsonFields is the set of keys a struct accepts, in the spelling the parser sees.
func jsonFields(v any) []string {
	var out []string
	rt := reflect.TypeOf(v)
	for i := range rt.NumField() {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if tag := strings.Split(f.Tag.Get("json"), ",")[0]; tag != "" && tag != "-" {
			out = append(out, tag)
		}
	}
	slices.Sort(out)
	return out
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// TestChartRendersTheParsedDocument renders the chart with every value set and feeds the ConfigMap it produces to the
// parser: what the chart writes has to be what the server reads.
func TestChartRendersTheParsedDocument(t *testing.T) {
	helm := lookupHelm(t)
	out, err := exec.Command(helm, "template", "rz", chartDir, "-n", "resourcerer", "-f", valuesFile).Output()
	if err != nil {
		t.Fatalf("helm template: %v", err)
	}
	doc, configMap, args := renderAuth(t, string(out))
	if configMap != "rz-resourcerer-auth" {
		t.Errorf("ConfigMap = %q", configMap)
	}
	// The flag and the object are named by the same helper; say so here so a rename has to break something.
	if !slices.Contains(args, "--auth-config-configmap=resourcerer/rz-resourcerer-auth") {
		t.Errorf("server args = %v, want the ConfigMap this render created", args)
	}

	want := string(readTestdata(t, goldenFile))
	if doc == want {
		return
	}
	var gotCfg, wantCfg any
	if err := yaml.Unmarshal([]byte(doc), &gotCfg); err != nil {
		t.Fatalf("the rendered %s is not yaml: %v", ConfigKey, err)
	}
	if err := yaml.Unmarshal([]byte(want), &wantCfg); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(gotCfg, wantCfg) {
		return // same document, different layout: the parser reads both the same way
	}
	t.Errorf("the chart renders a different document than %s, which the parser test reads:\n%s", goldenFile, doc)
}

// renderAuth picks the auth ConfigMap and the server's arguments out of a rendered release.
func renderAuth(t *testing.T, manifests string) (doc, configMap string, args []string) {
	t.Helper()
	for _, part := range strings.Split(manifests, "\n---\n") {
		var obj struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Data map[string]string `json:"data"`
			Spec struct {
				Template struct {
					Spec struct {
						Containers []struct {
							Args []string `json:"args"`
						} `json:"containers"`
					} `json:"spec"`
				} `json:"template"`
			} `json:"spec"`
		}
		if err := yaml.Unmarshal([]byte(part), &obj); err != nil {
			continue // a fragment that is not an object this test reads
		}
		switch {
		case obj.Kind == "ConfigMap" && strings.HasSuffix(obj.Metadata.Name, "-auth"):
			doc, configMap = obj.Data[ConfigKey], obj.Metadata.Name
		case obj.Kind == "Deployment" && strings.HasSuffix(obj.Metadata.Name, "-server"):
			if len(obj.Spec.Template.Spec.Containers) > 0 {
				args = obj.Spec.Template.Spec.Containers[0].Args
			}
		}
	}
	if doc == "" {
		t.Fatal("the render has no ConfigMap ending in -auth")
	}
	if _, err := ParseConfig(doc); err != nil {
		t.Errorf("the server would refuse what the chart rendered: %v\ndocument:\n%s", err, doc)
	}
	return doc, configMap, args
}

// TestSchemaAcceptsOnlyWhatTheParserDoes is the same rule from the other side: values.schema.json checks the shape of
// the OIDC values before rendering, and anything it lets through has to be a value this package accepts, or the pod
// exits with a message nobody asked for. The empty string is excluded: it is how values.yaml spells unset, and the
// chart leaves those keys out of the document.
func TestSchemaAcceptsOnlyWhatTheParserDoes(t *testing.T) {
	var schema struct {
		Definitions struct {
			OIDC struct {
				Properties struct {
					Issuer      jsonPattern `json:"issuer"`
					RedirectURL jsonPattern `json:"redirectURL"`
					SessionTTL  jsonPattern `json:"sessionTTL"`
				} `json:"properties"`
			} `json:"oidcConfig"`
		} `json:"definitions"`
	}
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	patterns := map[string]*regexp.Regexp{}
	for name, p := range map[string]string{
		"issuer":      schema.Definitions.OIDC.Properties.Issuer.Pattern,
		"redirectURL": schema.Definitions.OIDC.Properties.RedirectURL.Pattern,
		"sessionTTL":  schema.Definitions.OIDC.Properties.SessionTTL.Pattern,
	} {
		if p == "" {
			t.Fatalf("%s has no pattern in the schema", name)
		}
		patterns[name] = regexp.MustCompile(p)
	}

	// One direction, the one that hurts: whatever the schema wave through has to be a value the parser takes. The
	// schema may be stricter, which only costs an install that has to reword a value. It is stricter in three known
	// places: an IPv6 literal host, a host with a character outside the hostname alphabet, and a value with a space
	// anywhere but a percent escape.
	for _, value := range []string{
		"https://sso.example.com/realms/example", "https://sso.example.com", "https://sso.example.com/",
		"https://sso.example.com:8443/realms/example", "http://localhost:8080/realms/example",
		"https://sso.example.com/auth/callback?x=1", "https://sso.example.com/a%20b?r=%2Fx#frag",
		// Rejected by the schema and by the parser alike: no host, whitespace, no scheme, a scheme that is not http.
		"https://", "http://", "https://sso.example.com /realms", "https://sso.example.com/x y",
		"sso.example.com", "//sso.example.com/realms", "ftp://sso.example.com/realms",
		// And the shapes url.Parse alone let through, which the schema and httpURL now both refuse: credentials,
		// a port that is not a number, an escape that is not one, a bracketed host with no bracket, a backslash.
		"https://user@sso.example.com/realms", "https://user@", "https://sso.example.com:abc",
		"https://sso.example.com/a%zz", "https://[::1", "https://sso.example.com\\realms",
		"https://sso.example.com/a\tb", "http://sso.example.com:8080/realms\x01",
	} {
		for _, field := range []string{"issuer", "redirectURL"} {
			if !patterns[field].MatchString(value) {
				continue
			}
			if err := httpURL(field, value); err != nil {
				t.Errorf("the schema accepts %s %q and the parser refuses it: %v", field, value, err)
			}
		}
	}

	// The chart's TTL pattern bounds the shape, and templates/auth-configmap.yaml the range, which is what keeps a
	// duration the parser refuses from ever rendering.
	for _, value := range []string{
		"12h", "8h", "1h", "168h", "5m", "30m", "10080m",
		"1m", "0s", "200h", "1d", "1h30m", "forever", "1.5h", "1w",
	} {
		if !patterns["sessionTTL"].MatchString(value) {
			continue
		}
		d, err := time.ParseDuration(value)
		if err != nil {
			t.Errorf("the schema accepts sessionTTL %q, which is not a duration the server reads", value)
			continue
		}
		if d < minSessionTTL || d > maxSessionTTL {
			continue // out of range is templates/auth-configmap.yaml's fail, not the parser's shape
		}
		if _, err := ParseConfig("issuer: https://sso.example.com\nclientID: rz\nsessionTTL: " + value); err != nil {
			t.Errorf("the schema accepts sessionTTL %q and the parser refuses it: %v", value, err)
		}
	}
}

// TestSessionTTLRangeIsTheSameOnBothSides renders the chart once per TTL and compares its verdict with the server's.
// The chart's own range check is the only thing standing between a typo and a pod that exits at startup, and the two
// sides convert the number differently: Helm's int reads a leading zero as octal, where time.ParseDuration reads
// decimal, so "09m" was a refusal here and an acceptance there and "0200h" the other way round. The values all match
// the schema's TTL pattern, so a refusal by the chart is the range check and not the shape check.
func TestSessionTTLRangeIsTheSameOnBothSides(t *testing.T) {
	helm := lookupHelm(t)
	for _, ttl := range []string{
		"5m", "05m", "09m", "30m", "10080m", "1h", "012h", "168h",
		"0m", "1m", "01m", "4m", "10081m", "169h", "0200h", "020000m",
	} {
		rendered := exec.Command(helm, "template", "rz", chartDir, "-n", "resourcerer", "-f", valuesFile,
			"--set", "auth.oidc.sessionTTL="+ttl).Run() == nil
		_, err := ParseConfig("issuer: https://sso.example.com\nclientID: rz\nsessionTTL: " + ttl)
		if rendered != (err == nil) {
			t.Errorf("sessionTTL %s: the chart rendered it = %v, the server accepts it = %v (%v)",
				ttl, rendered, err == nil, err)
		}
	}
}

// TestBothCopiesOfTheURLPatternAreOnePattern reads the URL pattern from both places the schema states it — the
// definition every field uses, and the copy the strategy=oidc branch needs because there an unset issuer is not
// allowed — and fails when they drift, since the branch is what validates a real install and the definition is what
// the test above proves against the parser.
func TestBothCopiesOfTheURLPatternAreOnePattern(t *testing.T) {
	var schema struct {
		Properties struct {
			Auth struct {
				OneOf []struct {
					Properties struct {
						OIDC struct {
							Properties struct {
								Issuer jsonPattern `json:"issuer"`
							} `json:"properties"`
						} `json:"oidc"`
					} `json:"properties"`
				} `json:"oneOf"`
			} `json:"auth"`
		} `json:"properties"`
		Definitions struct {
			OIDC struct {
				Properties struct {
					Issuer      jsonPattern `json:"issuer"`
					RedirectURL jsonPattern `json:"redirectURL"`
				} `json:"properties"`
			} `json:"oidcConfig"`
		} `json:"definitions"`
	}
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	// The definition also admits the empty string, which is how values.yaml spells unset; the branch does not.
	want := strings.TrimPrefix(schema.Definitions.OIDC.Properties.Issuer.Pattern, "^$|")
	branch := ""
	for _, one := range schema.Properties.Auth.OneOf {
		if p := one.Properties.OIDC.Properties.Issuer.Pattern; p != "" {
			if branch != "" {
				t.Fatal("two strategy branches state an issuer pattern; this test reads one")
			}
			branch = p
		}
	}
	if branch == "" {
		t.Fatal("no strategy branch states an issuer pattern")
	}
	for name, got := range map[string]string{
		"the strategy=oidc branch": branch,
		"redirectURL":              schema.Definitions.OIDC.Properties.RedirectURL.Pattern,
	} {
		other := strings.TrimPrefix(got, "^$|")
		if other != want {
			t.Errorf("%s uses a different URL pattern than auth.oidc.issuer:\n%s\n%s", name, want, other)
		}
	}
}

func lookupHelm(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("helm")
	if err != nil {
		// CI installs helm so this guard cannot quietly stop running.
		if os.Getenv("CI") != "" {
			t.Fatalf("helm is not available: %v", err)
		}
		t.Skip("helm is not installed")
	}
	return path
}

type jsonPattern struct {
	Pattern string `json:"pattern"`
}
