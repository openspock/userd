package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/openspock/crypto/hashes"
	"github.com/openspock/log"
)

// RoleType helps define the type of a role that can be used by operations to
// ascertain that only users with certain roles can perform these
// operations.
type RoleType int

const (
	// Disregard role type for this operation
	Disregard RoleType = iota
	// Admin role type for this operation
	Admin
)

func (t RoleType) String() string {
	return [...]string{"<nil>", "admin"}[t]
}

// CreateUser creates a new user.
func CreateUser(email, password, description, roleID, file, adminUsr, adminPwd string) error {
	log.Info("CreateUser", log.AppMsg, map[string]interface{}{"email": email, "description": description})

	if adminUsr != "init" {
		Authenticate(adminUsr, adminPwd, file)
	}

	c, err := NewConfig(file)
	if err != nil {
		return err
	}
	secret := make([]byte, 8)
	_, err = rand.Read(secret)
	if err != nil {
		return err
	}
	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return err
	}
	saltStr := base64.StdEncoding.EncodeToString(salt)
	hash, err := hashes.CalculateHmacSha256([]byte(password+saltStr), secret)
	if err != nil {
		return err
	}
	u, err := NewUser(email, description, string(secret), saltStr, string(hash), roleID)
	if err != nil {
		return err
	}
	if err := c.WriteUser(u); err != nil {
		return err
	}
	log.Info("CreateUser", log.AppMsg, map[string]interface{}{"email": email, "result": "success", "message": email + " has been created successfully"})
	return nil
}

// ChangePassword changes the password for a user.
func ChangePassword(email, password, newPassword, confirmPassword, file string) error {
	log.Info("ChangePassword", log.AppMsg, map[string]interface{}{"email": email})

	if newPassword != confirmPassword {
		return errors.New("new and confirm password not the same")
	}

	c, err := NewConfig(file)
	if err != nil {
		return err
	}

	if err := Authenticate(email, password, file); err != nil {
		return err
	}

	u := UserTable[email]

	hash, err := hashes.CalculateHmacSha256([]byte(newPassword+u.Salt), []byte(u.secret))
	if err != nil {
		return err
	}
	u.hash = string(hash)

	if err := c.WriteUser(&u); err != nil {
		return err
	}

	log.Info("ChangePassword", log.AppMsg, map[string]interface{}{"email": email, "result": "success", "message": "password changed successfully for " + email})

	return nil
}

// ExpireUser sets the expiration date of the user to now.
func ExpireUser(email string) {
	log.Info("ExpireUser", log.AppMsg, map[string]interface{}{"email": email})
}

// CreateRole creates a new role.
func CreateRole(name, file string) (*Role, error) {
	log.Info("CreateRole", log.AppMsg, map[string]interface{}{"role_name": name})
	c, err := NewConfig(file)
	if err != nil {
		return nil, err
	}
	r, err := NewRole(name)
	if err != nil {
		return nil, err
	}
	if err := c.WriteRole(r); err != nil {
		return nil, err
	}
	log.Info("CreateRole", log.AppMsg, map[string]interface{}{"role_name": name, "result": "success", "message": name + " has been created"})
	return r, nil
}

// CreateFP creates a new file permission for either a user or a role.
func CreateFP(file string, user *User, role *Role, expiration time.Time, location string) (*FilePermission, error) {
	log.Info("CreateFP", log.AppMsg, map[string]interface{}{"file": file})

	c, err := NewConfig(location)
	if err != nil {
		return nil, err
	}

	fp, err := NewFP(file, *user, *role, expiration)
	if err != nil {
		return nil, err
	}

	if err := c.WriteFP(fp); err != nil {
		return nil, err
	}

	log.Info("CreateFP", log.AppMsg, map[string]interface{}{"file": file, "result": "success", "message": "Permission for " + file + " has been created"})

	return fp, nil
}

// Authenticate authenticates a user's credentials for access to the system.
func Authenticate(email, password, file string) error {
	log.Info("Authenticate", log.AppMsg, map[string]interface{}{"email": email})

	if _, err := NewConfig(file); err != nil {
		return err
	}

	v, ok := UserTable[email]
	if !ok {
		return errors.New(email + " does not exist")
	}

	h, err := hashes.CalculateHmacSha256([]byte(password+v.Salt), []byte(v.secret))
	if err != nil {
		return err
	}
	if string(h) != v.hash {
		return errors.New("password does not match")
	}

	log.Info("Authenticate", log.AppMsg, map[string]interface{}{"email": email, "result": "success", "message": "user successfully authenticated"})

	return nil
}

// AuthenticateForRole authenticates a user's credentials and then validates
// if the user is of a required role for that operation.
func AuthenticateForRole(email, password, file string, roleType RoleType) error {
	if err := Authenticate(email, password, file); err != nil {
		return err
	}
	if roleType.String() != RoleTable[UserTable[email].RoleID].Name {
		return errors.New("role does not match " + roleType.String())
	}

	return nil
}

// Authorize authorizes acccess to a resource.
func Authorize(email, password, file, resource string) error {
	log.Info("Authorize", log.AppMsg, map[string]interface{}{"email": email})

	if err := Authenticate(email, password, file); err != nil {
		return err
	}

	u := UserTable[email]
	var fps []FilePermission
	var ok bool
	fps, ok = FilePermissionTable[u.UserID][resource]

	if !ok {
		// check for role specific perms
		fps, ok = FilePermissionTable[""][resource]
		if !ok {
			return errors.New(resource + " permission does not exist for " + email)
		}
	}
	var isRoleOk bool = false
	var isExpirationOk bool = false
	for _, fp := range fps {
		if !isRoleOk && fp.Role.RoleID == u.RoleID {
			isRoleOk = true
		}
		if time.Now().Before(fp.Expiration) {
			isExpirationOk = true
		}
	}

	if !isRoleOk {
		return errors.New("user does not have required role")
	}
	if !isExpirationOk {
		return errors.New("file permission expired")
	}

	log.Info("Authorize", log.AppMsg, map[string]interface{}{"email": email, "result": "success", "message": "user successfully authorized", "resource": resource})

	return nil
}

// GetRoleIDFor returns the RoleID for a RoleName
func GetRoleIDFor(name string) (string, error) {
	for _, v := range RoleTable {
		if v.Name == name {
			return v.RoleID, nil
		}
	}
	return "", errors.New("Role not found for name " + name)
}

// ListRoles lists all available roles.
func ListRoles() map[string]interface{} {
	m := make(map[string]interface{})

	for k, v := range RoleTable {
		m[k] = v
	}

	return m
}
