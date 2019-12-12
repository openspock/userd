package userd

import "testing"

func TestCreateUser(t *testing.T) {
	err := CreateUser("testing@email.org", "whyilovewritingunittests", "a test description", "file://./config/sample")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}
