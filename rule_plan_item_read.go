package networkmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

// Read responde la lectura de plan_item, que no es una tabla: tú produces las filas. filter trae los campos que envió quien llama.
// crud ya validó los argumentos, aplicó el tope de página y comprobó el permiso antes de llamarte.
func (p *PlanItem) Read(a crud.CurrentUser, filter *PlanItem, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	return nil, nil


}
