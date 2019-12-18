package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openspock/log"
	config "github.com/openspock/userd/config"
	user "github.com/openspock/userd/user"
)

const (
	nilCredentials = "<nil>"
)

var op string
var email string
var password string
var description string
var roleName string
var location string
var help bool
var adminEmail string
var adminPwd string
var verbose bool
var resource string
var expiration string
var newPassword string
var confirmPassword string

func init() {
	flag.StringVar(&op, "op", "", "Userd operation\n\t* create_user\n\t* create_role\n\t* assign_fp (assign file permissions)\n\t* list_roles (you will require the uuid when creating a user)\n\t* is_authorized (check if user is authorized to access resource/file)")
	flag.StringVar(&email, "email", "", "User email")
	flag.StringVar(&password, "password", "", "User password")
	flag.StringVar(&adminEmail, "admin-email", "", "Admin email * mandatory")
	flag.StringVar(&adminPwd, "admin-password", "", "Admin password * mandatory")
	flag.StringVar(&description, "description", "", "User description - please enter a string in quotes")
	flag.StringVar(&roleName, "role", "", "Role name")
	flag.StringVar(&location, "location", "", "Userd location * mandatory - this is the location of your userd config and data files. By default, this is C:\\Userd in windows and /etc/userd in *nix systems")
	flag.BoolVar(&help, "help", false, "Prints help")
	flag.BoolVar(&verbose, "verbose", false, "Print verbose logging information")
	flag.StringVar(&resource, "resource", "", "File URL to provide access to either a user email or role. If both are provided, role will be ignored.")
	flag.StringVar(&expiration, "expiration", "", "expiration date in yyyy-MM-dd format")
	flag.StringVar(&newPassword, "new-password", "", "New password")
	flag.StringVar(&confirmPassword, "confirm-password", "", "Confirm password")
}

func printHelp() {
	flag.PrintDefaults()
}

func handleError(msg interface{}) {
	if msg == nil {
		return
	}
	fmt.Println("####################    error    ####################")
	fmt.Println()
	fmt.Println(msg)
	fmt.Println()
	fmt.Println("#####################################################")
	printHelp()
	os.Exit(1)
}

func validateMandatory() {
	switch op {
	case "is_authorized":
	case "change_password":
		break
	default:
		if adminEmail == "" || adminPwd == "" {
			handleError("Admin email(user) and password are mandatory")
		}
	}
}

func getRoleID() string {
	roleID, err := user.GetRoleIDFor(roleName)
	if err != nil {
		handleError(err)
	}
	return roleID
}

func getRole() user.Role {
	return user.RoleTable[getRoleID()]
}

func getExpirationDate() time.Time {
	if expiration == "" {
		handleError("expiration is required in yyyy-MM-dd format")
	}
	date, err := time.Parse("2006-01-02", expiration)
	if err != nil {
		handleError(err)
	}

	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
}

func handleLocation() {
	if location == "" {
		location = config.GetDefaultLocation()
		_, err := user.NewConfig(location)
		handleError(err)
		if len(user.UserTable) == 0 && adminEmail == "" {
			adminEmail = string(nilCredentials)
			adminPwd = string(nilCredentials)
		}
	}

	if !strings.HasPrefix(location, "file://") {
		location = "file://" + location
	}
}

func handleFirstTime() {
	log.Disabled = true
	// uninitialized userd
	fmt.Println(`It seems like you are using userd for the` +
		`first time or have never initialized it for - ` + strings.Split(location, "://")[1])
	fmt.Println("Please type <userd -help> if you think this message does not make sense.")
	fmt.Println("Would you like to initialize userd at this location? [y|n]")
	var answer rune
	if _, err := fmt.Scanf("%c\n", &answer); err != nil {
		handleError(err)
	}
	if answer != 'y' {
		fmt.Println("Please enter the location where you'll like to initialize userd. Press Ctrl|Cmd ^ C to exit the program.")
		if _, err := fmt.Scanf("%s\n", &location); err != nil {
			handleError(err)
		}
	}

	role, err := user.CreateRole("admin", location)
	if err != nil {
		handleError(err)
	}

	fmt.Println("Great! We've initialized userd at this location. Next, let's setup an admin user.")
	fmt.Println("We use emails as userid. Please enter an email you'd like to use as your username.")
	fmt.Println("We promise not to send unnecessary spam! :) ")
	fmt.Print("email: ")
	fmt.Scanln(&adminEmail)
	fmt.Print("password: ")
	fmt.Scanln(&adminPwd)

	if err := user.CreateUser(adminEmail, adminPwd, "Userd admin", role.RoleID, location, "init", "init"); err != nil {
		handleError(err)
	}
	fmt.Println("You're all set up and ready to go.")
	fmt.Println()
	fmt.Println("Please type <userd -help> to get a list of options.")
}

func createUser() {
	if email == "" {
		handleError("email is required")
	}
	if password == "" {
		handleError("password is required")
	}
	if description == "" {
		handleError("A description for this user is required")
	}

	roleID := getRoleID()

	if err := user.CreateUser(email, password, description, roleID, location, adminEmail, adminPwd); err != nil {
		handleError(err)
	}

	fmt.Println("User created successfully!")
}

func createRole() {
	if roleName == "" {
		handleError("roleName is required")
	}
	role, err := user.CreateRole(roleName, location)
	if err != nil {
		handleError(err)
	}
	fmt.Println("Role " + roleName + " created successfully with id: " + role.RoleID)
}

func listRoles() {
	if log.Disabled {
		for _, v := range user.ListRoles() {
			fmt.Printf("%s : %s \n", v.(user.Role).RoleID, v.(user.Role).Name)
		}
	}
	log.Info("All available roles", log.AppMsg, user.ListRoles())
}

func assignFP() {
	if resource == "" {
		handleError("resource is required")
	}

	if email == "" && roleName == "" {
		handleError("Either email or role name is required")
	}

	var u user.User
	var ok bool
	if email != "" {
		u, ok = user.UserTable[email]
		if !ok {
			handleError(email + " does not exist")
		}
	}

	var role user.Role
	if email == "" {
		role = getRole()
	}

	if _, err := user.CreateFP(resource, &u, &role, getExpirationDate(), location); err != nil {
		handleError(err)
	}
}

func isAuthorized() {
	if email == "" || password == "" {
		handleError("credentials are missing")
	}

	if resource == "" {
		handleError("resource is required")
	}

	if err := user.Authorize(email, password, location, resource); err != nil {
		handleError(err)
	}
}

func changePassword() {
	if email == "" || password == "" {
		handleError("email and password are required")
	}

	if newPassword == "" || confirmPassword == "" {
		handleError("new-password and confirm-password are required")
	}

	user.ChangePassword(email, password, newPassword, confirmPassword, location)
}

func handleOp() {
	if op == "" {
		fmt.Println("op is a mandatory parameter. Select one of the options specified for op.")
		handleError("op is mandatory")
	}

	switch op {
	case "create_role":
		createRole()
	case "create_user":
		createUser()
	case "list_roles":
		listRoles()
	case "assign_fp":
		assignFP()
	case "is_authorized":
		isAuthorized()
	case "change_password":
		changePassword()
	default:
		handleError("This op is not supported!")
	}
}

// parse parses and handles all flags passed to userd
func parse() {
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if !verbose {
		log.Disabled = true
	}

	handleLocation()

	validateMandatory()
}

func main() {
	parse()

	if adminEmail == nilCredentials {
		handleFirstTime()
	} else {
		// authenticate admin

		switch op {
		case "change_password":
		case "is_authorized":
			break
		default:
			if err := user.AuthenticateForRole(adminEmail, adminPwd, location, user.Admin); err != nil {
				handleError(err)
			}
		}

		handleOp()
	}
}
