package security // import "golang.binggl.net/commons/security"

import (
	"fmt"
	"net/url"
	"strings"
)

// Authorize validates the given claims and verifies if
// they match the required claim
// a claim entry is in the form "name|url|role"
func Authorize(required Claim, claims []string) (roles []string, err error) {
	for _, claim := range claims {
		c := split(claim)
		ok, _ := compareURL(required.URL, c.URL)
		if required.Name == c.Name && matchRole(c.Roles, required.Roles) && ok {
			return c.Roles, nil
		}
	}
	return roles, fmt.Errorf("supplied claims are not sufficient")
}

func matchRole(a []string, b []string) bool {
	for _, r := range a {
		for _, s := range b {
			if s == r {
				return true
			}
		}
	}
	return false
}

func split(claim string) *Claim {
	parts := strings.Split(claim, "|")
	if len(parts) == 3 {
		r := strings.Split(parts[2], ";")
		return &Claim{Name: parts[0], URL: parts[1], Roles: r}
	}
	return &Claim{}
}

func compareURL(a, b string) (bool, error) {
	var (
		urlA *url.URL
		urlB *url.URL
		err  error
	)
	if urlA, err = url.Parse(a); err != nil {
		return false, err
	}
	if urlB, err = url.Parse(b); err != nil {
		return false, err
	}
	if urlA.Scheme != urlB.Scheme || urlA.Port() != urlB.Port() || urlA.Host != urlB.Host {
		return false, fmt.Errorf("The urls do not match: '%s vs. %s'", urlA, urlB)
	}

	if normalizePath(urlA.Path) != normalizePath(urlB.Path) {
		return false, fmt.Errorf("The path of the urls does not match: '%s vs. %s'", urlA.Path, urlB.Path)
	}
	return true, nil
}

func normalizePath(path string) string {
	if path != "" {
		end := path[len(path)-1:]
		if end == "/" {
			return path[:len(path)-1]
		}
	}
	return path
}
