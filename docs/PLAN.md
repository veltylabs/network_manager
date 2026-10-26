---
PLAN: "feat: migrate ui to view.Screen and drop layout dependency"
TAG: v0.1.83
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

# Migración de UI a view.Screen y eliminación de dependencia webtyp/layout

## Ecosystem search

- `webtyp.com/view` (`v0.6.45`): define el contrato de pantalla declarativa `view.Screen`, `view.Presenter`, `view.WithDescription`, `view.WithTitle`, y el validador `Screen.Validate()`.
- `veltylabs/network_manager`: actualmente exporta `ui.Browser(...)`, `ui.PlanBrowser(...)`, `ui.ConnectionsBrowser(...)`, `ui.HistoryBrowser(...)` importando `webtyp.com/layout/platformd`, `crudview`, forzando a este módulo a acoplarse con librerías de chasis visual y renderizado.

## Design gate

1. **Prior art**: `webtyp/view` v0.6.44 introdujo `view.Screen` como el contrato único para que los módulos declaren qué muestran como datos puros, sin importar renderizadores ni chasis.
2. **Novice-name test**: `ui.Screens(caller router.Caller) []view.Screen` devuelve las pantallas que ofrece este módulo al menú de la aplicación.
3. **Complexity ledger**: Elimina `webtyp.com/layout` (y sus dependencias indirectas) de `go.mod`.
4. **Where it belongs**: `ui.Screens` vive en `github.com/veltylabs/network_manager/ui`. Los constructores de Presenter viven en la raíz de `networkmanager`.
5. **What it deletes**: Borra `ui.Browser`, `ui.PlanBrowser`, `ui.ConnectionsBrowser`, `ui.HistoryBrowser` y el require de `webtyp.com/layout` en `go.mod`.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/view@v0.6.45
go get webtyp.com/router@v0.4.8
```

## APIs used

| Symbol | Signature | Tag | File |
|---|---|---|---|
| `view.Screen` | `type Screen struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Lookups []Lookup; Help lang.Text; Children []Screen }` | `webtyp.com/view v0.6.45` | `screen.go:35` |
| `view.Presenter` | `type Presenter interface { ... }` | `webtyp.com/view v0.6.45` | `presenter.go:12` |
| `view.WithTitle` | `func WithTitle(title string) Option` | `webtyp.com/view v0.6.45` | `view.go:124` |
| `view.WithDescription` | `func WithDescription(d lang.Text) Option` | `webtyp.com/view v0.6.45` | `view.go:138` |
| `view.New` | `func New(l Lister, record model.Model, opts ...Option) Presenter` | `webtyp.com/view v0.6.45` | `view.go:147` |
| `router.Caller` | `type Caller interface { Call(op string, args Encodable, out Decodable, done func(error)); Dispatch(op string, args Encodable) }` | `webtyp.com/router v0.4.8` | `router.go` |

## Deletions

- `ui.Browser`, `ui.PlanBrowser`, `ui.ConnectionsBrowser`, `ui.HistoryBrowser` en `ui/browser.go`. Criterio: `grep -rn "func.*Browser(" ui/` -> vacío.
- Todo import de `webtyp.com/layout` en el repositorio. Criterio: `grep -rn "webtyp.com/layout" .` -> vacío.

---

## Stage 1: Soportar view.Option en constructores de Presenter

En `view.go`:
Permitir que `NewSettingView`, `NewPlanView`, `NewConnectionView` y `NewApplyHistoryView` reciban `opts ...view.Option` para configurar descripciones u opciones adicionales.

Actualizar `view.go`:

```go
package networkmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	titleSettings    = "Red"
	titlePlan        = "Cambios pendientes"
	titleConnections = "Conectados"
	titleHistory     = "Historial de cambios"

	labelApply   = "Apply"
	confirmApply = "Apply these changes to the router?"

	registeredLabel   = "registrado"
	unregisteredLabel = "sin registrar"
	arrow             = " → "
	separator         = " · "
)

func (s *NetworkSetting) Item() view.Item {
	return view.Item{ID: s.Id, Label: s.DhcpServer, Description: s.FilterDns + separator + s.Unregistered}
}

func (p *PlanItem) Item() view.Item {
	desc := p.Reason
	if p.Category == PlanItemChange {
		desc = p.Before + arrow + p.After
	}
	label := p.Object
	if label == "" {
		label = p.Name
	}
	return view.Item{ID: p.Id, Label: label, Description: desc}
}

func (c *ConnectionRow) Item() view.Item {
	label := c.Name
	if label == "" {
		label = c.HostName
	}
	if label == "" {
		label = c.Mac
	}
	state := unregisteredLabel
	if c.Registered {
		state = registeredLabel
	}
	return view.Item{ID: c.Mac, Label: label, Description: c.Ip + separator + state}
}

func (a *NetworkApply) Item() view.Item {
	return view.Item{ID: a.Id, Label: a.Fingerprint, Description: a.TenantId}
}

// NewSettingView lists and saves the tenant's settings.
func NewSettingView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkSettings, Save: OpSaveNetworkSetting},
		func() model.ModelSlice { return &NetworkSettingList{} })
	allOpts := append([]view.Option{view.WithTitle(titleSettings)}, opts...)
	return view.New(b, &NetworkSetting{}, allOpts...)
}

// NewPlanView lists the pending plan and offers the Apply action.
func NewPlanView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpPlanNetwork, Actions: []view.Action{{
			Op:      OpApplyNetwork,
			Label:   labelApply,
			Confirm: confirmApply,
			Args: func(records []model.Model) model.Encodable {
				item := records[0].(*PlanItem)
				return &ApplyNetworkArgs{Fingerprint: item.Fingerprint}
			},
		}}},
		func() model.ModelSlice { return &PlanItemList{} })
	allOpts := append([]view.Option{view.WithTitle(titlePlan)}, opts...)
	return view.New(b, &PlanItem{}, allOpts...)
}

// NewConnectionView lists who is connected now.
func NewConnectionView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListConnections},
		func() model.ModelSlice { return &ConnectionRowList{} })
	allOpts := append([]view.Option{view.WithTitle(titleConnections)}, opts...)
	return view.New(b, &ConnectionRow{}, allOpts...)
}

// NewApplyHistoryView lists past applies, newest first.
func NewApplyHistoryView(caller router.Caller, opts ...view.Option) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, List: OpListNetworkApplies},
		func() model.ModelSlice { return &NetworkApplyList{} })
	allOpts := append([]view.Option{view.WithTitle(titleHistory)}, opts...)
	return view.New(b, &NetworkApply{}, allOpts...)
}
```

---

## Stage 2: Declarar ui.Screens sin dependencias de layout

En `ui/browser.go`:
Eliminar `Browser`, `PlanBrowser`, `ConnectionsBrowser`, `HistoryBrowser`, `screen`.
Implementar `Screens(caller router.Caller) []view.Screen` con las 4 pantallas.

Reescribir `ui/browser.go`:

```go
package ui

import (
	networkmanager "github.com/veltylabs/network_manager"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	descSettings    lang.Text = "Network configuration: DHCP server, DNS filtering, and domain controls."
	descPlan        lang.Text = "Pending network configuration changes awaiting router application."
	descConnections lang.Text = "Active client connections, IP and MAC address leases."
	descHistory     lang.Text = "Audit trail of previously applied network changes."
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       ID,
			Permission: networkmanager.ResourceNetworkSetting,
			Presenter:  networkmanager.NewSettingView(caller, view.WithDescription(descSettings)),
		},
		{
			ID:         PlanID,
			Label:      PlanLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewPlanView(caller, view.WithDescription(descPlan)),
		},
		{
			ID:         ConnectionsID,
			Label:      ConnectionsLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewConnectionView(caller, view.WithDescription(descConnections)),
		},
		{
			ID:         HistoryID,
			Label:      HistoryLabel,
			Icon:       ID,
			Permission: networkmanager.ResourceNetwork,
			Presenter:  networkmanager.NewApplyHistoryView(caller, view.WithDescription(descHistory)),
		},
	}
}
```

---

## Stage 3: Actualizar pruebas unitarias

Actualizar las pruebas en `tests/` para verificar `ui.Screens(caller)`:
- `len(screens) == 4`
- Cada pantalla valida con `s.Validate() == nil` y tiene `Permission != ""`.
- Asegurar que ningún archivo de prueba importe `webtyp.com/layout`.

---

## Stage 4: Limpieza de dependencias en go.mod y verificación

Ejecutar:
```bash
go mod tidy
```

Verificar que:
1. `grep "webtyp.com/layout" go.mod` retorna vacío.
2. `grep "webtyp.com/layout" go.sum` retorna vacío.
3. `grep -rn "webtyp.com/layout" .` retorna vacío.

Ejecutar la suite de pruebas completa:
```bash
gotest
```

---

## Stage 5: Documentación y regla de idioma en veltylabs

1. En `AGENTS.md` y `README.md`, reemplazar referencias a `ui.Browser(...) (platformd.UIModule, error)` por `ui.Screens(caller router.Caller) []view.Screen`.
2. Ejecutar `langc sync .` y verificar 0 errores.

---

## Stages table

| Etapa | Archivos modificados | Qué hace |
|---|---|---|
| Stage 1 | `view.go` | Agrega `opts ...view.Option` a constructores de Presenter |
| Stage 2 | `ui/browser.go` | Exporta `ui.Screens` y elimina imports de `layout` |
| Stage 3 | `tests/` | Valida `ui.Screens` y remueve dependencias de `layout` |
| Stage 4 | `go.mod`, `go.sum` | `go mod tidy` y ejecución de `gotest` |
| Stage 5 | `AGENTS.md`, `README.md`, `lang.json` | Actualiza documentación en español y sincroniza `langc` |

## Delivery check

1. `grep -rn "webtyp.com/layout" .` retorna vacío.
2. `gotest` pasa con 0 fallos de compilación, 0 fallos de tests y cobertura verificada.
3. `langc sync .` pasa sin errores.
4. El PR se abre inmediatamente al pasar estas comprobaciones.
