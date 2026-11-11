package networkmanager

import "webtyp.com/crud"

// RunImport es la llamada externa de network_import (RunImport): llama al sistema externo usando key como clave de idempotencia.
// crud guarda los campos devueltos en el registro y reintenta o marca según el modo.
func (p *NetworkImport) RunImport(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	return nil, nil


}
