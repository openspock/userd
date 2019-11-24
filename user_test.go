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
	if len(UserTable) != 1 {
		t.Error("UserTable should have exactly 1 entry")
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
	fp := FilePermissionTable[u.UserID]
	for k, v := range fp {
		if k != "./filepermission.conf" {
			t.Error("file name incorrect - " + k)
			t.Fail()
		}
		if v[0].UserID != u.UserID {
			t.Error("file permission for incorrect user id found - " + v[0].UserID)
			t.Fail()
		}
		if v[0].Roles.Name == "admin" {
			t.Error("incorrect role assigned for file permission - " + v[0].Roles.Name)
			t.Fail()
		}
	}
}
