package networkmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Create es la regla de negocio de crear network_apply: valida lo que solo tu dominio sabe y responde si se guarda y si se avisa.
// crud ya validó el esquema, los campos solo del servidor, los límites y el permiso antes de llamarte.
func (p *NetworkApply) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	plan, _, err := m.Plan(m.tenantID)
	if err != nil {
		return crud.RuleAnswer{}, err
	}
	if p.Fingerprint != string(plan.Fingerprint) {
		return crud.RuleAnswer{}, &model.FieldError{Field: "fingerprint", Err: model.Invalid("stale fingerprint")}
	}

	p.Status = "pending"
	p.ActorId = a.UserID

	return crud.ApplyAndPublish(NetworkApplyApplyGateway{}), nil
}
