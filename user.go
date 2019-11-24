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
	c.Init()
	return &c, nil
}

// parsing logic for User, FilePermission and Role

type parseRecord func([]string) (interface{}, string, error)

func parseUser(record []string) (interface{}, string, error) {
	createdTime, err := time.Parse(time.RFC3339, record[5])
	if err != nil {
		return User{}, "", err
	}
	u := User{record[0], record[1], record[2], record[3], record[4], createdTime}
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

// Init initializes userd configuration.
//
// 1. init user conf
// 2. int fperm conf
func (c *Configuration) Init() error {
	err := c.initFile(c.Location+"/user.conf", parseUser, userTableInsert)
	err = c.initFile(c.Location+"/role.conf", parseRoles, roleTableInsert)
	err = c.initFile(c.Location+"/filepermission.conf", parseFilePermission, filePermissionTableInsert)
	return err
}

func (c *Configuration) initFile(file string, handler parseRecord, insertIntoTable tableInsert) error {
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

// UserTable is a map of user email to User
var UserTable = make(map[string]User)

// FilePermissionTable is a map of UserID to a map of File to FilePermission
var FilePermissionTable = make(map[string]map[string][]FilePermission)

// RoleTable is a map of RoleID to Role
var RoleTable = make(map[string]Role)
