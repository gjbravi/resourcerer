package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

func newKey(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func TestCodec(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	oldKey, newKey := newKey(t), newKey(t)
	c, err := NewCodec(oldKey, "")
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := c.Seal(SessionCookie, session{User: "alice", Groups: []string{"sre"}}, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	var got session
	if err := c.Open(SessionCookie, sealed, now, &got); err != nil || got.User != "alice" || got.Groups[0] != "sre" {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
	if err := c.Open(SessionCookie, sealed, now.Add(time.Hour), &got); !errors.Is(err, ErrExpired) {
		t.Errorf("at expiry = %v, want ErrExpired", err)
	}
	if err := c.Open(stateCookie, sealed, now, &got); !errors.Is(err, ErrInvalid) {
		t.Errorf("sealed for another cookie = %v, want ErrInvalid", err)
	}
	raw, _ := base64.RawURLEncoding.DecodeString(sealed)
	for _, i := range []int{0, len(raw) / 2, len(raw) - 1} {
		tampered := append([]byte(nil), raw...)
		tampered[i] ^= 1
		if err := c.Open(SessionCookie, base64.RawURLEncoding.EncodeToString(tampered), now, &got); !errors.Is(err, ErrInvalid) {
			t.Errorf("byte %d flipped = %v, want ErrInvalid", i, err)
		}
	}
	for _, bad := range []string{"", "!!", "AAAA"} {
		if err := c.Open(SessionCookie, bad, now, &got); !errors.Is(err, ErrInvalid) {
			t.Errorf("Open(%q) = %v, want ErrInvalid", bad, err)
		}
	}

	rotated, err := NewCodec(newKey, oldKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := rotated.Open(SessionCookie, sealed, now, &got); err != nil {
		t.Errorf("previous key no longer opens: %v", err)
	}
	fresh, _ := rotated.Seal(SessionCookie, session{User: "bob"}, now.Add(time.Hour))
	if err := c.Open(SessionCookie, fresh, now, &got); !errors.Is(err, ErrInvalid) {
		t.Errorf("rotated codec sealed with the previous key: %v", err)
	}
	onlyNew, _ := NewCodec(newKey, "")
	if err := onlyNew.Open(SessionCookie, fresh, now, &got); err != nil || got.User != "bob" {
		t.Errorf("rotated codec did not seal with the primary key: %+v, %v", got, err)
	}
}

func TestNewCodecRejectsBadKeys(t *testing.T) {
	short := base64.StdEncoding.EncodeToString(make([]byte, 16))
	for _, tc := range []struct{ primary, previous string }{
		{"", ""},
		{"not base64!", ""},
		{short, ""},
		{newKey(t), short},
		{base64.StdEncoding.EncodeToString(make([]byte, 32)), ""},
		{base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))), ""},
		{newKey(t), base64.StdEncoding.EncodeToString(make([]byte, 32))},
	} {
		if _, err := NewCodec(tc.primary, tc.previous); err == nil {
			t.Errorf("NewCodec(%q, %q) accepted", tc.primary, tc.previous)
		}
	}
}
