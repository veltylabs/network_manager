package networkmanager

import (
	"webtyp.com/crud"
	"webtyp.com/network"
)

func (p *NetworkApply) ApplyGateway(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	desired, err := m.Desired(m.tenantID)
	if err != nil {
		return nil, err
	}
	plan, err := m.gateway.Apply(desired, network.Fingerprint(p.Fingerprint))
	if err != nil {
		return nil, err
	}
	p.Status = "done"
	p.Changes = int64(len(plan.Changes))
	return crud.Fields("status", "changes"), nil
}
