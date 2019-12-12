// Package userd contains structures and functions to manage users, their roles
// and permissions.
package userd

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AccessPermissions is a enumeration constant for resource access permissions.
type AccessPermissions int

const (
	// Read indicates read right
	Read AccessPermissions = 1 << iota
	// Write indicates write right
	Write AccessPermissions = 1 << iota
	// ReadWrite indicates both read and write right
	ReadWrite AccessPermissions = Read | Write
)

// Role represents a role assignable to a User
type Role struct {
	RoleID string
	Name   string
}

// NewRole creates a new Role and returns it.
func NewRole(name string) (*Role, error) {
	for _, v := range RoleTable {
		if v.Name == name {
			return nil, errors.New(name + " already exists")
		}
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &Role{RoleID: uuid.String(), Name: name}, nil
}

// User represents a user which requires authorization.
// Authentication is managed separately using FilePermission
//
// Users are stored in user.conf
type User struct {
	UserID      string
	secret      string
	Salt        string
	Email       string
	Description string
	Since       time.Time
	RoleID      string
}

func (u User) String() string {
	return u.Email
}

// NewUser creates a new user and stores it in user conf.
func NewUser(email string, description string, secret string, salt string, roleID string) (*User, error) {
	if _, ok := RoleTable[roleID]; !ok {
		return nil, errors.New(roleID + " does not exist")
	}
	for k := range UserTable {
		if email == k {
			return nil, errors.New(email + " already exists")
		}
	}
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	u := User{UserID: uuid.String(), secret: secret, Salt: salt, Email: email, Description: description, Since: time.Now(), RoleID: roleID}
	return &u, nil
}

// FilePermission represents permissions per file/ resource, per user.
//
// A file or resource can be identified with a URL.
// Examples -
// file:/etc/userd/user.conf
// https://openspock.org/userd/user.conf
//
// FilePermissions are persisted in fperm.conf
type FilePermission struct {
	File       string
	UserID     string
	Roles      Role
	Assignment time.Time
	Expiration time.Time
}

// Protocol has configuration file access protocol.
type Protocol int

const (
	// File is local file access protocol.
	File Protocol = iota << 1
)

// Configuration represents userd configuration.
type Configuration struct {
	Location           string
	FileAccessProtocol Protocol
}

// NewConfig builds a new Configuration by taking a
// file directory as an input. For e.g.
//
// file:/etc/userd
// https://openspock.org/userd
func NewConfig(file string) (*Configuration, error) {
	p := strings.Split(file, "://")
	if len(p) != 2 {
		return nil, errors.New("file doesn't have protocol information")
	}
	var protocol Protocol
	switch p[0] {
	case "file":
		protocol = File
	default:
		return nil, errors.New("unknown protocol")
	}
	c := Configuration{p[1], protocol}
	c.InitRead()
	return &c, nil
}

func (c *Configuration) userConfFileName() string {
	return c.Location + "/user.conf"
}

func (c *Configuration) roleConfFileName() string {
	return c.Location + "/role.conf"
}

func (c *Configuration) filePermissionFileName() string {
	return c.Location + "/filepermission.conf"
}

// InitRead initializes userd configuration.
//
// 1. init user conf
// 2. int fperm conf
func (c *Configuration) InitRead() error {
	err := c.initExisting(c.userConfFileName(), parseUser, userTableInsert)
	err = c.initExisting(c.roleConfFileName(), parseRoles, roleTableInsert)
	err = c.initExisting(c.filePermissionFileName(), parseFilePermission, filePermissionTableInsert)
	return err
}

func (c *Configuration) initExisting(file string, handler parseRecord, insertIntoTable tableInsert) error {
	config, err := os.Open(file)
	if err != nil {
		return err
	}
	r := csv.NewReader(config)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		u, key, err := handler(record)
		if err != nil {
			return err
		}
		insertIntoTable(key, u)
	}
	return nil
}

// WriteUser writes a user to the user conf file.
func (c *Configuration) WriteUser(u *User) error {
	return c.write(c.userConfFileName(), []string{u.UserID, u.secret, u.Salt, u.Email, u.Description, u.Since.Format(time.RFC3339)})
}

// WriteRole writes a role to the role conf file.
func (c *Configuration) WriteRole(r *Role) error {
	return c.write(c.roleConfFileName(), []string{r.RoleID, r.Name})
}

func (c *Configuration) write(file string, entry []string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write(entry); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

// parsing logic for User, FilePermission and Role

type parseRecord func([]string) (interface{}, string, error)

func parseUser(record []string) (interface{}, string, error) {
	createdTime, err := time.Parse(time.RFC3339, record[5])
	if err != nil {
		return User{}, "", err
	}
	u := User{record[0], record[1], record[2], record[3], record[4], createdTime, record[6]}
	return u, u.Email, nil
}

func parseRoles(record []string) (interface{}, string, error) {
	return Role{record[0], record[1]}, record[0], nil
}

func parseFilePermission(record []string) (interface{}, string, error) {
	assignment, err := time.Parse(time.RFC3339, record[3])
	if err != nil {
		return FilePermission{}, "", err
	}
	expiration, err := time.Parse(time.RFC3339, record[4])
	if err != nil {
		return FilePermission{}, "", err
	}
	role := RoleTable[record[2]]
	return FilePermission{record[0], record[1], role, assignment, expiration}, record[0], nil
}

// table insertion logic handlers

type tableInsert func(string, interface{})

func userTableInsert(key string, val interface{}) {
	if u, ok := val.(User); ok {
		UserTable[key] = u
	}
}

func roleTableInsert(key string, val interface{}) {
	if r, ok := val.(Role); ok {
		RoleTable[key] = r
	}
}

func filePermissionTableInsert(key string, val interface{}) {
	if fp, ok := val.(FilePermission); ok {
		if fpMap := FilePermissionTable[key]; fpMap == nil {
			FilePermissionTable[key] = make(map[string][]FilePermission)
		}
		if fps := FilePermissionTable[key][fp.File]; fps == nil {
			FilePermissionTable[key][fp.File] = []FilePermission{fp}
		} else {
			fps := FilePermissionTable[key][fp.File]
			fps = append(fps, fp)
			FilePermissionTable[key][fp.File] = fps
		}
	}
}

// UserTable is a map of user email to User
var UserTable = make(map[string]User)

// FilePermissionTable is a map of UserID to a map of File to FilePermission
var FilePermissionTable = make(map[string]map[string][]FilePermission)

// RoleTable is a map of RoleID to Role
var RoleTable = make(map[string]Role)
