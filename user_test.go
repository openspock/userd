package userd

import (
	"testing"
)

func TestParseUser(t *testing.T) {
	_, err := NewConfig("file://./config/sample")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	u := UserTable["test@openspock.org"]
	if &u == nil {
		t.Error("User can't be nil")
		t.Fail()
	}
	t.Log(u.Email)
	if u.Email != "test@openspock.org" {
		t.Error("user email invalid " + u.Email)
		t.Fail()
	}
}
