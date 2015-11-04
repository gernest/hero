package hero

import (
	"testing"
)

func TestExtraScopes(t *testing.T) {
	sample := []struct {
		access, refresh string
		result          bool
	}{
		{"one,two,three", "one", true},
		{"one,two,three", "none", false},
	}

	for _, scope := range sample {
		if e := extraScopes(scope.access, scope.refresh); e != scope.result {
			t.Errorf("expected %v got %v  aceess: %s refresh:: %s", scope.result, e, scope.access, scope.refresh)
		}
	}
}

func TestValidURL(t *testing.T) {

	link := "http://www.example.com"

	sample := []struct {
		info, base, redir string
		valid             bool
	}{
		{"exact match", "/hero", "/hero", true},
		{"trailing slash", "/hero", "/hero/", true},
		{"exact match with trailing slash", "/hero/", "/hero/", true},
		{"subpath", "/hero", "/hero/sub/path", true},
		{"subpath with trailing slash", "/hero/", "/hero/sub/path", true},
		{"subpath with traversal like", "/hero", "/hero/.../..sub../...", true},
		{"traversal", "/hero/../allow", "/hero/../allow/sub/path", true},
		{"base path mismatch", "/hero", "/heroine", false},
		{"base path mismatch slash", "/hero/", "/hero", false},
		{"traversal", "/hero", "/hero/..", false},
		{"embed traversal", "/hero", "/hero/../sub", false},
		{"not subpath", "/hero", "/hero../sub", false},
	}

	for _, v := range sample {
		if v.valid {
			err := validateURI(link+v.base, link+v.redir)
			if err != nil {
				t.Errorf("some fish for %s : %v", v.info, err)
			}
		} else {
			err := validateURI(link+v.base, link+v.redir)
			if err == nil {
				t.Errorf("expected error for for %s : got %v", v.info, err)
			}
		}
	}

	sampleList := []struct {
		base, redir, sep string
		valid            bool
	}{
		{"http://www.example.com/hero", "http://www.example.com/hero", "", true},
		{"http://www.example.com/hero", "http://www.example.com/app", "", false},
		{"http://xxx:14000/hero;http://www.example.com/hero", "http://www.example.com/hero", ";", true},
		{"http://xxx:14000/hero;http://www.example.com/hero", "http://www.example.com/app", ";", false},
	}

	for _, v := range sampleList {
		if v.valid {
			err := validateURIList(v.base, v.redir, v.sep)
			if err != nil {
				t.Error(err)
			}
		} else {
			err := validateURIList(v.base, v.redir, v.sep)
			if err == nil {
				t.Error("expected an error")
			}
		}
	}
}
