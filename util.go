package hero

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"
)

var (
	errBlankURL    = errors.New("urls can not be blank")
	errFragmentURL = errors.New("url must not include fragment")
)

// firstURI returns the first string after spliting base using sep. if sep is an empty string
// then base is returned.
//
// This is used to find the first redirect url from a url list.
func firstURI(base, sep string) string {
	if sep != "" {
		l := strings.Split(base, sep)
		if len(l) > 0 {
			return l[0]
		}
	}
	return base
}

func validateURI(base, redir string) error {
	if base == "" || redir == "" {
		return errBlankURL
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return err
	}

	redirectURL, err := url.Parse(redir)
	if err != nil {
		return err
	}

	if baseURL.Fragment != "" || redirectURL.Fragment != "" {
		return errFragmentURL
	}
	if baseURL.Scheme != redirectURL.Scheme {
		return fmt.Errorf("%s : %s / %s", "scheme mismatch", base, redir)
	}
	if baseURL.Host != redirectURL.Host {
		return fmt.Errorf("%s : %s / %s", "host mismatch", base, redir)
	}

	if baseURL.Path == redirectURL.Path {
		return nil
	}

	reqPrefix := strings.TrimRight(baseURL.Path, "/") + "/"
	if !strings.HasPrefix(redirectURL.Path, reqPrefix) {
		return fmt.Errorf("%s : %s / %s", "path is not a subpath", base, redir)
	}

	for _, s := range strings.Split(strings.TrimPrefix(redirectURL.Path, reqPrefix), "/") {
		if s == ".." {
			return fmt.Errorf("%s : %s / %s", "subpath cannot contain path traversial", base, redir)
		}
	}
	return nil
}

func validateURIList(baseList, redir, sep string) error {
	var list []string
	if sep != "" {
		list = strings.Split(baseList, sep)
	} else {
		list = append(list, baseList)
	}
	for _, item := range list {
		if err := validateURI(item, redir); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%s : %s / %s", "url dot validate", baseList, redir)

}

// checks whether refresh is contained in the acess string, access and refresh are
// scope strings that is coma separated words(or group of utf-8 characters). Returns
// true if refresh is in the cope of access and false otherwise.
func extraScopes(access, refresh string) bool {
	acessList := strings.Split(access, ",")
	refreshList := strings.Split(refresh, ",")

	var found bool
END:
	for _, rScope := range refreshList {

		for _, aScope := range acessList {
			if rScope != "" && aScope != "" && aScope == rScope {
				found = true
				break END
			}
		}
	}
	return found
}

//isEmail returns true if the given string is meail
func isEmail(str string) bool {
	return govalidator.IsEmail(str)
}

func hashString(secret string) (string, error) {
	s, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func compareHashedString(hashed, str string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(str))
}
