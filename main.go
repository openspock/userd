package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/openspock/log"
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
	flag.StringVar(&location, "location", "", "Userd location - mandatory - this is the location of your userd config and data files. By default, this is C:\\Userd in windows and /etc/userd in *nix systems")
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
		log.Info("using default location as location is not passed", log.AppLog, map[string]interface{}{"location": location})
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
		// uninitialized userd
		log.Info(`It seems like you are using userd for the`+
			`first time or have never initialized it for - `+location,
			log.AppLog,
			map[string]interface{}{"location": location})
	} else {
		// authenticate admin
		if err := user.Authenticate(adminEmail, adminPwd, location); err != nil {
			handleError(err)
		}
	}

}
