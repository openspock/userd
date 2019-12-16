package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/openspock/crypto/hashes"
	"github.com/openspock/log"
)

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
	b := make([]byte, 8)
	_, err = rand.Read(b)
	if err != nil {
		return err
	}
	salt := make([]byte, 8)
	_, err = rand.Read(salt)
	if err != nil {
		return err
	}
	saltStr := base64.StdEncoding.EncodeToString(salt)
	secretBytes, err := hashes.CalculateHmacSha256([]byte(password+saltStr), b)
	if err != nil {
		return err
	}
	u, err := NewUser(email, description, string(b), saltStr, string(secretBytes), roleID)
	if err != nil {
		return err
	}
	if err := c.WriteUser(u); err != nil {
		return err
	}
	log.Info("CreateUser", log.AppMsg, map[string]interface{}{"email": email, "result": "success", "message": email + " has been created successfully"})
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