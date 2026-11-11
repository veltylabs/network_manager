# network_manager — arquitectura

## Alcance del dominio

`network_manager` hace que la red de un sitio **siga su inventario de dispositivos**. El inventario (`device_manager`) dice qué dispositivos existen, sus tarjetas de red (MAC, IP) y cuánto de la red puede usar cada uno. Este módulo convierte eso en un plan para el router, lo muestra a un administrador junto con quién está conectado en este momento, y aplica exactamente lo que el administrador revisó.

Lo encuentras cuando un administrador registra una PC o cambia el nivel de acceso de un dispositivo y luego abre la pantalla "Red" para revisar y aplicar el cambio, o quiere saber quién está conectado antes de tocar el router.

No conoce **marcas de routers**: el router llega como un `webtyp.com/network.Gateway` (`veltylabs/mikrotik` en producción, `network/mem` en tests).

## Entidades (Reglas de negocio frente a Persistencia)

- **NetworkSetting** (tabla): los valores del sitio. El CRUD lo persiste. Las reglas de dominio validan la configuración al crear o actualizar (`rule_network_setting_*.go`).
- **NetworkApply** (tabla): historial de aplicaciones. La regla de creación (`rule_network_apply_create.go`) verifica el plan actual y encola la llamada externa (`call_network_apply_ApplyGateway.go`) para contactar al router mediante `RetryInBackground`. El CRUD orquesta la persistencia, la llamada al router, y publica `network_manager.network.applied`.
- **NetworkImport** (tabla): progreso de la importación. La regla de creación encola la llamada externa (`call_network_import_RunImport.go`) para leer del router y registrar en el inventario.
- Filas de solo lectura y modelos de presentación (sin tabla, proveídos al vuelo): **PlanItem**, **ConnectionRow**, **DiscoveredRow**.

## Flujo

El framework CRUD persiste las tablas. El módulo declara sus reglas de dominio:
1. Petición HTTP -> CRUD -> Validación de esquema
2. Operación de lectura/escritura -> `rule_*.go` -> Módulo procesa dominio
3. Retorno a CRUD -> Encola `call_*.go` (si lo hay) y almacena -> Retorna HTTP

## Operaciones

Se listan en el README. Cada modelo define si es lectura o creación.

## Fuera de alcance (fases posteriores)

Puntos de acceso centralizados (CAPsMAN), red de invitados (VLAN), integración QR de teléfonos, y mostrar dispositivos en el plano de planta de `room_layout` — cada uno extiende `webtyp.com/network` primero.
