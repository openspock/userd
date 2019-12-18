# userd - a simple user management (non)daemon program.

`userd` is a simple program that simplifies user management and ACL for various resource in a system.

It enables an admin to setup userd in a convenient location. An admin can then create users, assign permissions to file(resources - we will use file and resources alternatively through out the doc), create roles, etc.

A user can check if they are authorized to access a particular resource. The definition of a resource is generic and hence a http url can be a resource and also a file. A resource is usually identified by it's name or url. It's key is a `string` and is `case-sensitive`. A user can also change their password.

`userd` defines clear boundaries for user management as follows - 
* user - a user of the system.
* role - each user should have a certain role to access resources of the system.
* file permissions - file permissions define basic ACL structure to access resources.

## userd - lifecycle functions

`userd` defines the following commands or operations - 
* `create_user` - creates new user. This is an elevated operation and requires admin creds.
* `create_role` - creates new role. This is an elevated operation and requires admin creds.
* `assign_fp` - assign file permissions to a user. This is an elevated operation and requires admin creds.
* `list_roles` - list all roles supported.This is an elevated operation and requires admin creds.
* `is_authorized` - check if a user is authorized to access a resource. 
* `change_password` - resets user password, requires user credentials.

## default locations

* `C:\Userd` - Windows
* `/etc/userd` - *nix systems

## running userd for the first time

`userd` understands if it's being run for the `first time` by checking if configuration and data files are present in the location parameter(default location if it's not passed). When `userd` is run for the first time, it walks the user through setup by creating an `admin` role and then asking the user to setup an `admin` user. 

## who can invoke these lifecycle functions

Each `write` interaction with `userd` will require admin credentials. Read operations do not require admin credentials but require user credentials. 