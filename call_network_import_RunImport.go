package networkmanager

import "webtyp.com/crud"

func (p *NetworkImport) RunImport(a crud.CurrentUser, m *Module, key crud.IdempotencyKey) (crud.FieldsToSave, error) {
	found, err := m.gateway.Discover()
	if err != nil {
		return nil, err
	}
	res, err := m.importer.ImportHosts(found)
	if err != nil {
		return nil, err
	}
	p.Imported = int64(res.Created)
	p.Skipped = int64(len(res.Skipped))
	p.Status = "done"
	return crud.Fields("status", "imported", "skipped"), nil
}
