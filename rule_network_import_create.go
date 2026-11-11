package networkmanager

import "webtyp.com/crud"

// Create es la regla de negocio de crear network_import: valida lo que solo tu dominio sabe y responde si se guarda y si se avisa.
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (p *NetworkImport) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	p.Status = "pending"
	p.ActorId = a.UserID
	return crud.ApplyAndPublish(NetworkImportRunImport{}), nil
}
