package networkmanager

import "webtyp.com/crud"

// Update es la regla de negocio de actualizar network_import: old es la fila guardada y el receptor trae solo los campos enviados (fields).
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (next *NetworkImport) Update(a crud.CurrentUser, tx crud.Tx, old *NetworkImport, fields []string, m *Module) (crud.RuleAnswer, error) {
	return crud.RuleAnswer{}, nil


}
