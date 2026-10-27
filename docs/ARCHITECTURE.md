# network_manager — arquitectura

## Alcance del dominio

`network_manager` hace que la red de un sitio **siga su inventario de dispositivos**. El inventario (`device_manager`) dice qué dispositivos existen, sus tarjetas de red (MAC, IP) y cuánto de la red puede usar cada uno. Este módulo convierte eso en un plan para el router, lo muestra a un administrador junto con quién está conectado en este momento, y aplica exactamente lo que el administrador revisó.

Lo encuentras cuando un administrador registra una PC o cambia el nivel de acceso de un dispositivo y luego abre la pantalla "Red" para revisar y aplicar el cambio, o quiere saber quién está conectado antes de tocar el router.

No conoce **marcas de routers**: el router llega como un `webtyp.com/network.Gateway` (`veltylabs/mikrotik` en producción, `network/mem` en tests).

## Entidades

- **NetworkSetting** (una por inquilino): los valores de todo el sitio que cambian sin recompilar — nombre del servidor DHCP del router, el pool dinámico, el resolutor DNS de filtrado, y qué obtienen los dispositivos sin registrar (`no_address` o `local`). Mapea 1:1 a `network.Settings`.
- **NetworkApply** (historial): una fila por aplicación exitosa — cuándo, el fingerprint, y cuántos cambios. Publicado como `network_manager.network.applied`.
- Filas de solo salida (nunca se almacenan): **PlanItem** (un cambio, conflicto o advertencia del plan actual), **ConnectionRow** (un dispositivo en línea, marcado como registrado o no, con su nombre de inventario), **DiscoveredRow** (configuración de router hecha a mano, para importar).

## Flujo

```
device_manager ──HostSource.Hosts()──┐
NetworkSetting ──────────────────────┼─► network.Desired ─► Gateway.Plan ─► PlanItems (el operador revisa)
                                     │                                          │ fingerprint
operador: apply_network(fingerprint) ┴─► Gateway.Apply(desired, fingerprint) ◄──┘
                                          ├─ ErrPlanStale → 409 (re-planificar)
                                          ├─ ErrConflicts → 409
                                          └─ ok → fila NetworkApply + evento
```

## Decisiones

- **Planificar, revisar, aplicar** — nunca aplicar al guardar o con un temporizador. Justificación en la ARQUITECTURA de `webtyp.com/network` ("Planificar, luego aplicar").
- **La configuración vive en la base de datos**, por lo que un sitio cambia su filtro DNS o el nombre del servidor DHCP sin una nueva compilación.
- **Tipos de opción cerrada** (el radio `unregistered`) viven en el subpaquete `kinds/`: ormc solo importa un constructor de tipo de un paquete propio.
- **Sin configuración ⇒ sin plan.** Hasta que el inquilino guarda un `NetworkSetting`, `plan_network` y `apply_network` responden 400 con `la configuración de red no está configurada`: el módulo nunca adivina los nombres de los objetos del router.
- **Quién está conectado** viene del gateway, cruzado con el inventario por MAC, para que un dispositivo sin registrar se destaque antes de cerrar la red a dispositivos sin registrar.

## Operaciones

| Op | Acción | Recurso | Descripción |
|---|---|---|---|
| `list_network_settings` | `r` | `network_setting` | La configuración del inquilino como una lista de 0 o 1 filas (una vista enumera; vacío = no configurado) |
| `save_network_setting` | `c`/`u` | `network_setting` | Crear o reemplazar la configuración del inquilino |
| `plan_network` | `r` | `network` | Plan actual: cambios, conflictos, advertencias (+ fingerprint en cada fila) |
| `apply_network` | `u` | `network` | Aplicar el plan con el fingerprint dado |
| `list_connections` | `r` | `network` | Dispositivos en línea ahora, marcados registrados / sin registrar |
| `discover_network` | `r` | `network` | Configuración del router hecha a mano (vista previa de importación) |
| `import_network` | `c` | `network` | Crear hosts de inventario desde la configuración descubierta |
| `list_network_applies` | `r` | `network` | Historial de aplicaciones |

## Fuera de alcance (fases posteriores)

Puntos de acceso centralizados (CAPsMAN), red de invitados (VLAN), integración QR de teléfonos, y mostrar dispositivos en el plano de planta de `room_layout` — cada uno extiende `webtyp.com/network` primero.
