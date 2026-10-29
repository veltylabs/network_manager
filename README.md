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
nm.MountOperations(reg) // router.OperationRegistry
// Pantallas (ui/): Screens(caller router.Caller) []view.Screen
```

Esquema: `migrate.Migrate(conn, compiler)` desde el binario de migración de la aplicación.

## Operaciones

| Op | Acción | Recurso |
|---|---|---|
| `list_network_settings` | read | `network_setting` |
| `save_network_setting` | create/update | `network_setting` |
| `plan_network` | read | `network` |
| `apply_network` | update | `network` |
| `list_connections` | read | `network` |
| `discover_network` | read | `network` |
| `import_network` | create | `network` |
| `list_network_applies` | read | `network` |

## Demostración

Ejecuta `webtyp` en la raíz del repositorio: gateway en memoria y en el navegador con dos dispositivos conectados y una regla manual, sin inicio de sesión.

## Archivos clave

| Archivo | Propósito |
|---|---|
| `model.go` | Definiciones, constantes, errores, tema de eventos |
| `kinds/` | El radio `unregistered` (paquete propio, para ormc) |
| `model_orm.go` | Generado por `ormc` — no editar |
| `module.go` | `Module`, `New`, servicios (settings, plan, apply, connections, discover, import, history) |
| `ops.go` | `router.OperationModule` — `MountOperations` + handlers |
| `view.go` | Los cuatro presentadores; el presentador del plan declara la acción Aplicar |

## Documentación

| Documento | Descripción |
|---|---|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Alcance, entidades, flujo, decisiones, operaciones |
| [Diagrama de Base de Datos](docs/diagrams/database.md) | ERD en Mermaid |

Errores por `rpc.WriteError`; permisos por tabla (`model.ResourceOf`); guardar/borrar con `view.MountSave`/`MountDelete`; claves foráneas declaradas.
