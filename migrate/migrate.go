// Package migrate reconciles the database schema network_manager owns:
// NetworkSetting and NetworkApply. Deliberately not called by New, and in its
// own package so webtyp.com/ddl never enters a WASM build.
package migrate

import (
	"webtyp.com/ddl"

	networkmanager "github.com/veltylabs/network_manager"
)

func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	d := ddl.New(conn, ddlCompiler)
	if err := d.CreateTable(&networkmanager.NetworkSetting{}); err != nil {
		return err
	}
	return d.CreateTable(&networkmanager.NetworkApply{})
}
