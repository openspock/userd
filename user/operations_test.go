package user

import "testing"

func TestCreateUser(t *testing.T) {
	err := CreateUser("testing@email.org", "whyilovewritingunittests", "a test description", "1", "file://./config/sample", "", "")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestAuthenticateUser(t *testing.T) {
	err := Authenticate("testing@email.org", "whyilovewritingunittests", "file://./config/sample")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestCreateRoleShouldFailForExistingRole(t *testing.T) {
	if err := CreateRole("admin", "file://./config/sample"); err == nil {
		t.Errorf("CreateRole should fail for an existing role")
		t.FailNow()
	}
}

func TestCreateRoleForNonExistingRole(t *testing.T) {
	if err := CreateRole("api", "file://./config/sample"); err != nil {
		t.Errorf("CreateRole should fail for an existing role")
		t.FailNow()
	}
}
