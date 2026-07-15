package sentinel

import "testing"

func TestNewIsPassiveAndRejectsNilConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) returned nil error")
	}

	client, err := New(&Conn{MasterName: "master", Address: []string{"127.0.0.1:0"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()
}
