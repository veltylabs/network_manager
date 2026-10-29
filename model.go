package networkmanager

import (
	"webtyp.com/input"
	"webtyp.com/model"

	"github.com/veltylabs/network_manager/kinds"
)

// ModelName is this module's identity: every op is qualified as
// "network_manager.<name>" on the wire.
const ModelName = "network_manager"
var ResourceNetwork = model.FeatureResource("network")

const (


)

// PlanItem categories — the ONLY place these literals exist.
const (
	PlanItemChange   = "change"
	PlanItemConflict = "conflict"
	PlanItemWarning  = "warning"
)

// PlanItem kinds (network.ChangeKind names).
const (
	ChangeKindAdd    = "add"
	ChangeKindUpdate = "update"
	ChangeKindRemove = "remove"
	ChangeKindAdopt  = "adopt"
)

// ConnectionRow sources (network.Source names).
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

// ValidationError marks a client error (400) from a service method.




// RouterOS object names carry spaces, '-', '_' and '.' ("dhcp LAN", "default-dhcp").
// An explicit field whitelist replaces the Text kind's default floor.
var NetworkSettingModel = model.Definition{
	Name: "network_setting",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "dhcp_server", Type: input.Text(), NotNull: true, Item: model.ItemLabel, Permitted: model.Permitted{Letters: true, Numbers: true, Spaces: true, Extra: []rune{'-', '_', '.'}, Minimum: 1, Maximum: 64}},
		{Name: "dynamic_pool", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Letters: true, Numbers: true, Spaces: true, Extra: []rune{'-', '_', '.'}, Maximum: 64}},
		{Name: "filter_dns", Type: input.IP(), NotNull: true, Item: model.ItemDescription},
		{Name: "unregistered", Type: kinds.UnregisteredPolicy(), NotNull: true, Item: model.ItemDescription},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

// NetworkApply: one successful apply (history). Output-only: base kinds.
var NetworkApplyModel = model.Definition{
	Name: "network_apply",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, Item: model.ItemDescription},
		{Name: "fingerprint", Type: model.Text(), NotNull: true, Item: model.ItemLabel},
		{Name: "changes", Type: model.Int()},
		{Name: "applied_at", Type: model.NanoTime()},
	},
}

// Transport-only, output-only rows (Field.DB is nil): never stored.

var PlanItemModel = model.Definition{
	Name: "plan_item",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "category", Type: model.Text()},
		{Name: "kind", Type: model.Text()},
		{Name: "object", Type: model.Text(), Item: model.ItemLabel},
		{Name: "name", Type: model.Text()},
		{Name: "mac", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "before", Type: model.Text(), Item: model.ItemDescription},
		{Name: "after", Type: model.Text(), Item: model.ItemDescription},
		{Name: "reason", Type: model.Text(), Item: model.ItemDescription},
		{Name: "online", Type: model.Bool()},
		{Name: "fingerprint", Type: model.Text()},
	},
}

var ConnectionRowModel = model.Definition{
	Name: "connection_row",
	Fields: model.Fields{
		{Name: "mac", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "ip", Type: model.Text(), Item: model.ItemDescription},
		{Name: "host_name", Type: model.Text()},
		{Name: "source", Type: model.Text()},
		{Name: "registered", Type: model.Bool(), Item: model.ItemDescription},
		{Name: "name", Type: model.Text(), Item: model.ItemLabel},
	},
}

var DiscoveredRowModel = model.Definition{
	Name: "discovered_row",
	Fields: model.Fields{
		{Name: "mac", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "ip", Type: model.Text()},
		{Name: "name", Type: model.Text()},
		{Name: "internet", Type: model.Bool()},
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
	Fields: model.Fields{
		{Name: "fingerprint", Type: model.Text(), NotNull: true},
	},
}
