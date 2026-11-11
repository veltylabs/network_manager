package networkmanager

import "webtyp.com/crud"

// ApplyGateway es la llamada externa de network_apply (ApplyGateway): llama al sistema externo usando key como clave de idempotencia.
// crud guarda los campos devueltos en el registro y reintenta o marca según el modo.
func (p *NetworkApply) ApplyGateway(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	return nil, nil


}
