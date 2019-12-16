package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

func init() {
	flag.StringVar(&op, "op", "", "Userd operation")
	flag.StringVar(&email, "email", "", "User email")
	flag.StringVar(&password, "password", "", "User password")
	flag.StringVar(&adminEmail, "admin-email", "", "Admin email * mandatory")
	flag.StringVar(&adminPwd, "admin-password", "", "Admin password * mandatory")
	flag.StringVar(&description, "description", "", "User description")
	flag.StringVar(&roleName, "role", "", "Role name")
	flag.StringVar(&location, "location", "", "Userd location * mandatory - this is the location of your userd config and data files. By default, this is C:\\Userd in windows and /etc/userd in *nix systems")
	flag.BoolVar(&help, "help", false, "Prints help")
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
	if adminEmail == "" || adminPwd == "" {
		handleError("Admin email(user) and password are mandatory")
	}
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
}

// parse parses and handles all flags passed to userd
func parse() {
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	handleLocation()

	validateMandatory()
}

func main() {
	parse()

	if adminEmail == nilCredentials {
		log.Disabled = true
		// uninitialized userd
		fmt.Println(`It seems like you are using userd for the` +
			`first time or have never initialized it for - ` + strings.Split(location, "://")[1])

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

			if !strings.HasPrefix(location, "file://") {
				location = "file://" + location
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
	} else {
		// authenticate admin
		if err := user.Authenticate(adminEmail, adminPwd, location); err != nil {
			handleError(err)
		}
	}

}
