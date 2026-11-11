package networkmanager

import "webtyp.com/crud"

func (m *Module) StoreNetworkSetting() crud.Store {
	return m.store
}

func (m *Module) StoreNetworkApply() crud.Store {
	return m.store
}

func (m *Module) StoreNetworkImport() crud.Store {
	return m.store
}
