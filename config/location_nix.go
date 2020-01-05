// +build !windows

package config

// GetDefaultLocation returns the default userd location
func GetDefaultLocation() string {
	return "file:///etc/userd"
}

// GetUserConfFileName gets the file name for user conf file
func GetUserConfFileName() string {
	return "/user.conf"
}

// GetRoleConfFileName gets the file name for role conf file
func GetRoleConfFileName() string {
	return "/role.conf"
}

// GetFPFileName gets the file permission conf file
func GetFPFileName() string {
	return "/filepermission.conf"
}
