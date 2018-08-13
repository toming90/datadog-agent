// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package secrets

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	advapi32       = syscall.NewLazyDLL("advapi32.dll")
	procLogonUserA = advapi32.NewProc("LogonUserA")
)

const (
	// The user created at install time with low/no rights
	username             = "datadog_secretuser"
	passwordRegistryPath = "SOFTWARE\\Datadog\\Datadog Agent\\secrets"
)

func checkRights(path string) error {
	return nil
}

func getPasswordFromRegistry() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		passwordRegistryPath,
		registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return "", fmt.Errorf("Secret user password does not found in the registry")
		}
		return "", fmt.Errorf("can't read secrets user password from registry: %s", err)
	}
	defer k.Close()

	password, _, err = k.GetStringValue(username)
	if err != nil {
		return "", fmt.Errorf("Could not read password for secrets user: %s", err)
	}
	return password, nil
}

func setCmdSysProcAttr(cmd *exec.Cmd) error {
	password, err := getPasswordFromRegistry()
	if err != nil {
		return err
	}

	var token syscall.Handle
	res, _, err := procLogonUserA(
		username,
		".",                              // local account database
		password,                         // password
		Winbase.LOGON32_LOGON_BATCH,      // logon type
		Winbase.LOGON32_PROVIDER_DEFAULT, // logon provider (using default)
		uintptr(unsafe.Pointer(&token)),
	)

	// LogonUserA return 0 on failure
	if err != nil || int(res) == 0 {
		return nil, fmt.Errorf("failed to login as user %s: %s", username, err)
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Token = token
	return nil
}
