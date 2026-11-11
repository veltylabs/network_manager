# network_manager
<img src="docs/img/badges.svg">

Módulo de gestión de red para el ecosistema Velty: hace que la red de un sitio **siga su inventario de dispositivos**. Almacena la configuración de red del sitio, pide al router un plan, lo muestra junto con quién está conectado en este momento, aplica exactamente el plan que el operador revisó (botón Aplicar con confirmación) y mantiene un historial. No conoce marcas de routers: el router llega como un `webtyp.com/network.Gateway` (`veltylabs/mikrotik` en producción, `network/mem` en tests).

## Inicio Rápido

```go
nm, err := networkmanager.New(db, networkmanager.Deps{
    IDs:       idGenerator,                                           // model.IDGenerator
    Hosts:     devicemanager.NetworkHosts{Devices: dm, TenantID: t},  // network.HostSource
    Importer:  devicemanager.NetworkHosts{Devices: dm, TenantID: t},  // network.HostImporter
    Gateway:   gw,                                                    // network.Gateway (mikrotik.Open(...))
    Publisher: eventPublisher,                                        // opcional
    TenantID:  t,
})
// Pantallas (ui/): Screens(caller router.Caller) []view.Screen
```

## Operaciones

| Sin framework (hoy) | Con framework (después) |
|---|---|
| `list_network_settings (+save)` | `network_setting.read`, `network_setting.create`, `network_setting.update` |
| `plan_network` | read model `plan_item` |
| `apply_network` | `network_apply.create` con `fingerprint` |
| `list_connections` | read model `connection_row` |
| `discover_network` | read model `discovered_row` |
| `import_network` | `network_import.create` |
| `list_network_applies` | `network_apply.read` |

## Demostración

Ejecuta `webtyp` en la raíz del repositorio: gateway en memoria y en el navegador con dos dispositivos conectados y una regla manual, sin inicio de sesión.

## Archivos clave

| Archivo | Propósito |
|---|---|
| `definition.go` | Definiciones, constantes, errores, tema de eventos |
| `kinds/` | El radio `unregistered` (paquete propio, para ormc) |
| `definition_gen.go` | Generado por `definitionc` — no editar |
| `module.go` | `Module`, `New`, servicios de lectura |
| `rule_*.go` | Reglas de negocio del dominio por tabla |
| `call_*.go` | Llamadas externas integradas a crud |
| `ui/` | Presentadores de navegador |

## Documentación

| Documento | Descripción |
|---|---|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Alcance, entidades, reglas de negocio frente a persistencia |
| [Diagrama de Base de Datos](docs/diagrams/database.md) | ERD en Mermaid |
