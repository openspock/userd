package userd

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/openspock/crypto/hashes"
	"github.com/openspock/log"
)

// CreateUser creates a new user.
func CreateUser(email string, password string, description string, roleID string, file string) error {
	log.Info("CreateUser", log.AppMsg, map[string]interface{}{"email": email, "description": description})
	b := make([]byte, 8)
	_, err := rand.Read(b)
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
	u, err := NewUser(email, description, string(secretBytes), saltStr, roleID)
	if err != nil {
		return err
	}
	c, err := NewConfig(file)
	c.WriteUser(u)
	return nil
}

// ExpireUser sets the expiration date of the user to now.
func ExpireUser(email string) {
	log.Info("ExpireUser", log.AppMsg, map[string]interface{}{"email": email})
}

// CreateRole creates a new role.
func CreateRole(name string) {
	log.Info("CreateRole", log.AppMsg, map[string]interface{}{"role_name": name})
}
