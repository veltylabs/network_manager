// Package migrate reconcilia el esquema de base de datos que network_manager posee:
// NetworkSetting y NetworkApply. Deliberadamente no es llamado por New, y está en su
// propio paquete para que webtyp.com/ddl nunca entre en una compilación WASM.
package migrate

import (
	"webtyp.com/ddl"

	networkmanager "github.com/veltylabs/network_manager"
)

// Ejemplo: err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	// Sync y no CreateTable: crea la tabla que falta y además agrega las columnas que un modelo
	// ganó, así un campo nuevo llega a una base ya instalada. Flujo: https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md
	return ddl.New(conn, ddlCompiler).Sync(
		&networkmanager.NetworkSetting{},
		&networkmanager.NetworkApply{},
	)
}
