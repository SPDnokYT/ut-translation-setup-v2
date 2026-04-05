//go:build windows

package main

import (
	_ "embed"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

//go:embed assets/GodotPCKExplorer_1.5.3_native-console-win-64.zip
var pckExplorerBinZip []byte

const (
	pckExplorerZipName = "GodotPCKExplorer_1.5.3_native-console-win-64.zip"
	pckBinName         = "GodotPCKExplorer.Console.exe"
)

func steamPathFromRegistry() string {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Valve\Steam`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return ""
	}
	defer key.Close()

	value, _, err := key.GetStringValue("SteamPath")
	if err != nil {
		return ""
	}

	return value
}

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
