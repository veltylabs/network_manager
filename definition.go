package networkmanager

import (
	"webtyp.com/input"
	"webtyp.com/model"

	"github.com/veltylabs/network_manager/kinds"
)

// ModelName es la identidad de este módulo: cada operación se cualifica como
// "network_manager.<name>" en la red.
// Ejemplo: moduleName := networkmanager.ModelName
var ResourceNetwork = model.FeatureResource("network")

const (


)

// Categorías de PlanItem — el ÚNICO lugar donde existen estos literales.
const (
	PlanItemChange   = "change"
	PlanItemConflict = "conflict"
	PlanItemWarning  = "warning"
)

// Tipos de PlanItem (nombres de network.ChangeKind).
const (
	ChangeKindAdd    = "add"
	ChangeKindUpdate = "update"
	ChangeKindRemove = "remove"
	ChangeKindAdopt  = "adopt"
)

// Orígenes de ConnectionRow (nombres de network.Source).
const (
	SourceDHCP = "dhcp"
	SourceARP  = "arp"
)

const TopicNetworkApplied = "network_manager.network.applied"





const (
	ErrNotConfigured model.Invalid  = "network settings not configured"
	ErrNotFound      model.NotFound = "network setting not found"
	ErrCannotRead    model.Invalid  = "network_manager: cannot read the request"
)

// ValidationError marca un error del cliente (400) desde un método de servicio.




// Los nombres de objetos de RouterOS llevan espacios, '-', '_' y '.' ("dhcp LAN", "default-dhcp").
// Una lista blanca explícita de campos reemplaza el mínimo predeterminado del tipo Text.
var NetworkSettingModel = model.Definition{
	Name:  "network_setting",
	Allow: model.Create | model.Read | model.Update,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "dhcp_server", Type: input.Text(), NotNull: true, Item: model.ItemLabel, Permitted: model.Permitted{Letters: true, Numbers: true, Spaces: true, Extra: []rune{'-', '_', '.'}, Minimum: 1, Maximum: 64}},
		{Name: "dynamic_pool", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Letters: true, Numbers: true, Spaces: true, Extra: []rune{'-', '_', '.'}, Maximum: 64}},
		{Name: "filter_dns", Type: input.IP(), NotNull: true, Item: model.ItemDescription},
		{Name: "unregistered", Type: kinds.UnregisteredPolicy(), NotNull: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

// NetworkApply: una aplicación exitosa (historial). Solo salida: tipos base.
var NetworkApplyModel = model.Definition{
	Name:  "network_apply",
	Allow: model.Create | model.Read | model.Update,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule},
	ExternalCalls: []model.ExternalCallDefinition{
		{Name: "ApplyGateway", Mode: model.RetryInBackground, MaxAttempts: 3},
	},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, Item: model.ItemDescription, ServerOnly: true},
		{Name: "fingerprint", Type: model.Text(), NotNull: true, Item: model.ItemLabel},
		{Name: "status", Type: model.Text(), ServerOnly: true},
		{Name: "changes", Type: model.Int(), ServerOnly: true},
		{Name: "actor_id", Type: model.Text(), ServerOnly: true},
		{Name: "applied_at", Type: model.NanoTime(), ServerOnly: true},
	},
}

var NetworkImportModel = model.Definition{
	Name:  "network_import",
	Allow: model.Create | model.Read | model.Update,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule},
	ExternalCalls: []model.ExternalCallDefinition{
		{Name: "RunImport", Mode: model.RetryInBackground, MaxAttempts: 3},
	},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), ServerOnly: true},
		{Name: "status", Type: model.Text(), ServerOnly: true},
		{Name: "imported", Type: model.Int(), ServerOnly: true},
		{Name: "skipped", Type: model.Int(), ServerOnly: true},
		{Name: "actor_id", Type: model.Text(), ServerOnly: true},
		{Name: "created_at", Type: model.NanoTime(), ServerOnly: true},
	},
}

// Filas solo de transporte y salida (Field.DB es nil): nunca se almacenan.

var PlanItemModel = model.Definition{
	Name:  "plan_item",
	Allow: model.Read,
	Wire:  true,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "category", Type: model.Text(), OmitEmpty: true},
		{Name: "kind", Type: model.Text(), OmitEmpty: true},
		{Name: "object", Type: model.Text(), Item: model.ItemLabel, OmitEmpty: true},
		{Name: "name", Type: model.Text(), OmitEmpty: true},
		{Name: "mac", Type: model.Text(), OmitEmpty: true},
		{Name: "before", Type: model.Text(), Item: model.ItemDescription, OmitEmpty: true},
		{Name: "after", Type: model.Text(), Item: model.ItemDescription, OmitEmpty: true},
		{Name: "reason", Type: model.Text(), Item: model.ItemDescription, OmitEmpty: true},
		{Name: "online", Type: model.Bool(), OmitEmpty: true},
		{Name: "fingerprint", Type: model.Text(), OmitEmpty: true},
	},
}

var ConnectionRowModel = model.Definition{
	Name:  "connection_row",
	Allow: model.Read,
	Wire:  true,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "mac", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "ip", Type: model.Text(), Item: model.ItemDescription, OmitEmpty: true},
		{Name: "host_name", Type: model.Text(), OmitEmpty: true},
		{Name: "source", Type: model.Text(), OmitEmpty: true},
		{Name: "registered", Type: model.Bool(), Item: model.ItemDescription, OmitEmpty: true},
		{Name: "name", Type: model.Text(), Item: model.ItemLabel, OmitEmpty: true},
	},
}

var DiscoveredRowModel = model.Definition{
	Name:  "discovered_row",
	Allow: model.Read,
	Wire:  true,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "mac", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "ip", Type: model.Text(), OmitEmpty: true},
		{Name: "name", Type: model.Text(), OmitEmpty: true},
		{Name: "internet", Type: model.Bool(), OmitEmpty: true},
	},
}

var ImportReportModel = model.Definition{
	Name: "import_report",
	Fields: model.Fields{
		{Name: "created", Type: model.Int()},
		{Name: "skipped", Type: model.Int()},
	},
}

var ApplyNetworkArgsModel = model.Definition{
	Name: "apply_network_args",
	Wire: true,
	Fields: model.Fields{
		{Name: "fingerprint", Type: model.Text(), NotNull: true},
	},
}

var Definition = model.ModuleDefinition{
	Name: "network_manager",
	Entities: []model.Definition{
		NetworkSettingModel,
		NetworkApplyModel,
		NetworkImportModel,
		PlanItemModel,
		ConnectionRowModel,
		DiscoveredRowModel,
	},
}
