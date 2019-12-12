# userd - a simple user management (non)daemon program.

`userd` defines clear boundaries for user management as follows - 
* user - a user of the system.
* role - each user should have a certain role to access resources of the system.
* file permissions - file permissions define basic ACL structure to access resources.

## userd - lifecycle functions

`userd` defines the following lifecycle functions - 
* `create_user` - creates new user.
* `create_role` - creates new role.
* `assign_fp` - assign file permissions to a user.