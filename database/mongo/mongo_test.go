package mongo

import "testing"

func TestConnectionURIEscapesCredentialsAndOptions(t *testing.T) {
	got := connectionURI(&DBConn{
		Hosts:    []string{"db1:27017", "db2:27017"},
		Username: "user@example.com",
		Password: "pass:/?@",
		AuthDB:   "auth/db",
		Options: map[string]string{
			"z": "a b",
			"a": "x&y",
		},
	})
	want := "mongodb://user%40example.com:pass%3A%2F%3F%40@db1:27017,db2:27017/auth%2Fdb?a=x%26y&z=a+b"
	if got != want {
		t.Fatalf("connectionURI() = %q, want %q", got, want)
	}
}

func TestNewRejectsNilConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) returned nil error")
	}
}
