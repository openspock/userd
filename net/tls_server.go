// Package net contains structures and functions to enable userd to run as a secure
// tls server
package net

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"net"

	log "github.com/openspock/log"
	"github.com/openspock/userd/user"
)

// Command encapsulates all properties required by the tls server to execute an operation.
// Currently, command will only support authorization and authentication.
type Command struct {
	Op       string `json:"op"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Resource string `json:"resource"`
}

func (c Command) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		return "error"
	}
	return string(data)
}

// ExitCode indicates the type of response for a command (op) execution.
type ExitCode int

const (
	// Success indicates succesful execution of the command.
	Success ExitCode = iota
	// AuthenticationFailure indicates that the user could not be authenticated.
	AuthenticationFailure
	// AuthorizationFailure indicates that the user is not authorized.
	AuthorizationFailure
	// SystemError indicates an unexpected error on the server.
	SystemError
)

// Response is sent in response to an execution of a command on the server.
//
//
type Response struct {
	Code    ExitCode
	Message string
}

func (r Response) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		return "{\"code\":\"}" + err.Error() + "\"}"
	}
	return string(data)
}

// Listen starts a tls server on port provided and listens to incoming
// connections.
func Listen(port, certLocation string, location string) error {
	cer, err := tls.LoadX509KeyPair(certLocation+"/server.crt", certLocation+"/server.key")
	if err != nil {
		return err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":"+port, config)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Info("server started, ready to accept commands", log.SysLog, map[string]interface{}{})

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error(err.Error(), log.SysLog, map[string]interface{}{})
			continue
		}
		go handleConnection(conn, location)
	}
}

func handleConnection(conn net.Conn, location string) {
	defer conn.Close()

	var cmd Command

	r := bufio.NewReader(conn)
	//for {
	req := make([]byte, 1024)
	n, err := r.Read(req)

	if err != nil {
		log.Error(err.Error(), log.SysLog, map[string]interface{}{})
		return
	}

	json.Unmarshal([]byte(string(req[:n])), &cmd)
	log.Info(cmd.String(), log.AppLog, map[string]interface{}{})

	response := handleCommand(cmd, location)

	_, err = conn.Write([]byte(response.String()))
	if err != nil {
		log.Error(err.Error(), log.SysLog, map[string]interface{}{})
		return
	}
	//}
}

func handleCommand(cmd Command, location string) *Response {
	if cmd.Op != "is_authorized" {
		return &Response{Code: SystemError, Message: "command not supported"}
	}
	if err := user.Authorize(cmd.Email, cmd.Password, location, cmd.Resource); err != nil {
		return &Response{Code: SystemError, Message: err.Error()}
	}
	return &Response{Code: Success, Message: "Success"}
}
