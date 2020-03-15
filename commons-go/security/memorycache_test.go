package security // import "golang.binggl.net/commons/security"

import (
	"testing"
	"time"
)

// copied and derived from https://github.com/goenning/go-cache-demo
// see https://github.com/goenning/go-cache-demo/blob/master/LICENSE
// initially created by: https://github.com/goenning

const key = "MY_KEY"

func TestMemCacheGetEmpty(t *testing.T) {
	cache := NewMemCache(parse("5s"))
	user := cache.Get(key)
	if user != nil {
		t.Fatalf("exptected empty/nil user!")
	}
}

func TestMemCacheGetValue(t *testing.T) {
	cache := NewMemCache(parse("5s"))
	u := User{
		DisplayName: "a",
		Email:       "a.b@c.de",
		UserID:      "1",
		Username:    "u",
		Roles:       []string{"role"},
	}
	cache.Set(key, &u)
	user := cache.Get(key)
	if user == nil {
		t.Fatalf("exptected cached used got nil!")
	}
	if user != &u {
		t.Fatalf("the returned object/address is not the same!")
	}
}

func TestMemCacheGetExpiredValue(t *testing.T) {
	cache := NewMemCache(parse("1s"))
	u := User{
		DisplayName: "a",
		Email:       "a.b@c.de",
		UserID:      "1",
		Username:    "u",
		Roles:       []string{"role"},
	}
	cache.Set(key, &u)

	time.Sleep(parse("1s200ms"))

	user := cache.Get(key)
	if user != nil {
		t.Fatalf("exptected expiry of user object!")
	}
}

func parse(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
