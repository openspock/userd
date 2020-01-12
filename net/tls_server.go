// Package net contains structures and functions to enable userd to run as a secure
// tls server
package net

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"net"

	log "github.com/openspock/log"
)

// Command encapsulates all properties required by the tls server to execute an operation.
// Currently, command will only support authorization and authentication.
type Command struct {
	Op              string
	Email           string
	Password        string
	Description     string
	RoleName        string
	AdminEmail      string
	AdminPwd        string
	Verbose         string
	Resource        string
	Expiration      string
	NewPassword     string
	ConfirmPassword string
}

// Listen starts a tls server on port provided and listens to incoming
// connections.
func Listen(port, certLocation string) error {
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var cmd Command

	r := bufio.NewReader(conn)
	for {
		cmdStr, err := r.ReadString('\n')
		if err != nil {
			log.Error(err.Error(), log.SysLog, map[string]interface{}{})
			return
		}

		json.Unmarshal([]byte(cmdStr), &cmd)
		log.Info(cmdStr, log.AppLog, map[string]interface{}{})

		_, err = conn.Write([]byte(`{"result":"success"}`))
		if err != nil {
			log.Error(err.Error(), log.SysLog, map[string]interface{}{})
			return
		}
	}
}
