//go:build linux

package main

import (
	_ "embed"
	"syscall"
)

//go:embed assets/GodotPCKExplorer_1.5.3_native-console-linux-64.zip
var pckExplorerBinZip []byte

const (
	pckExplorerZipName = "GodotPCKExplorer_1.5.3_native-console-linux-64.zip"
	pckBinName         = "GodotPCKExplorer.Console"
)

func steamPathFromRegistry() string { return "" }

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
