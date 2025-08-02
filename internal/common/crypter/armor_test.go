package crypter_test

import (
	"testing"

	"golang.binggl.net/monorepo/internal/common/crypter"
)

func TestArmor(t *testing.T) {

	tests := map[string]struct {
		in      string
		out     string
		isError bool
	}{
		"input with length of exactly 64": {
			in:  "YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NG",
			out: "-----BEGIN ENCRYPTED CONTENT-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+\nIHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NG\n-----END ENCRYPTED CONTENT-----",
		},
		"input is less than armorLength": {
			in:  "TEST",
			out: "-----BEGIN ENCRYPTED CONTENT-----\nTEST\n-----END ENCRYPTED CONTENT-----",
		},
		"empty input": {
			in:      "",
			isError: true,
		},
		"input % armorLength != 0": {
			in:  "YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NGACEWWG",
			out: "-----BEGIN ENCRYPTED CONTENT-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+\nIHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NG\nACEWWG\n-----END ENCRYPTED CONTENT-----",
		},
		"input length is 31": {
			in:  "YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0",
			out: "-----BEGIN ENCRYPTED CONTENT-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0\n-----END ENCRYPTED CONTENT-----",
		},
		"input length is 33": {
			in:  "YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0AB",
			out: "-----BEGIN ENCRYPTED CONTENT-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0A\nB\n-----END ENCRYPTED CONTENT-----",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			armor, err := crypter.Armor(test.in)
			if test.isError {
				if err == nil {
					t.Errorf("expected error for test '%s'", name)
				}
				return
			}
			if err != nil {
				t.Error(err)
			}
			if armor != test.out {
				t.Errorf("expected ouptut is wrong; want '%v', got '%s'", test.out, armor)
			}
		})
	}
}

func TestDeArmor(t *testing.T) {
	tests := map[string]struct {
		in      string
		out     string
		isError bool
	}{
		"valid armor input": {
			out: "YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NG",
			in:  "-----BEGIN ENCRYPTED CONTENT-----\nYWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+\nIHNjcnlwdCB1dVdYQ3pDbFJkSENNa2NG\n-----END ENCRYPTED CONTENT-----",
		},
		"short armor": {
			out: "TEST",
			in:  "-----BEGIN ENCRYPTED CONTENT-----\nTEST\n-----END ENCRYPTED CONTENT-----",
		},
		"no armor": {
			out:     "",
			in:      "TEST",
			isError: true,
		},
		"no input": {
			out:     "",
			in:      "",
			isError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			armor, err := crypter.DeArmor(test.in)
			if test.isError {
				if err == nil {
					t.Errorf("expected error for test '%s'", name)
				}
				return
			}
			if err != nil {
				t.Error(err)
			}
			if armor != test.out {
				t.Errorf("expected ouptut is wrong; want '%v', got '%s'", test.out, armor)
			}
		})
	}
}
