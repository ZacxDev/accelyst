# Accelyst

Una interfaz de iteración de proyectos nativa para IA.

## Descripción

Accelyst transforma la estructura de tu proyecto y el backlog de funcionalidades en prompts listos para ejecución con el contexto completo del código. Cada prompt incluye los directorios relevantes, rutas de documentación, orden de dependencias y criterios de completitud—todo lo que un agente de IA necesita para implementar funcionalidades de forma autónoma.

- **Contexto enriquecido**: Los prompts referencian directorios exactos y documentación
- **Consciente de dependencias**: Hitos y pasos se ordenan topológicamente via DAG
- **Listo para paralelismo**: El trabajo independiente se marca para ejecución concurrente con subagentes
- **Consciente de niveles**: Distingue pasos ejecutables por IA de acciones que requieren humanos
- **Criterios de completitud**: Los criterios de aceptación definen cuándo los pasos están completos

## Instalación

```bash
go build -o accelyst
```

## Uso

```bash
./accelyst                      # usa accelyst.yaml
./accelyst -config custom.yaml  # archivo de configuración personalizado
```

## Formato de Configuración

```yaml
projectParts:
  - name: client
    directoryAbs: /ruta/al/cliente
    documentationAbs: /ruta/a/docs/cliente.md  # opcional
  - name: server
    directoryAbs: /ruta/al/servidor

milestones:
  - id: auth_api
    name: "API de Autenticación"
    dependsOn: []
    steps:
      - id: research
        instruction: "investigar mejores prácticas de JWT"
        dependsOn: []
        projectParts: []
        acceptanceCriteria:
          - "estrategia de expiración de tokens documentada"
      - id: implement
        instruction: "implementar endpoints de autenticación"
        dependsOn:
          - research
        projectParts:
          - server
        acceptanceCriteria:
          - "endpoints de login y logout funcionando"
          - "tests unitarios pasando"

  - id: auth_ui
    name: "UI de Autenticación"
    dependsOn:
      - auth_api
    steps:
      - id: build_form
        instruction: "implementar formulario de login"
        dependsOn: []
        projectParts:
          - client
      - id: configure_oauth
        instruction: "configurar proveedor OAuth en producción"
        dependsOn:
          - build_form
        tier: human
        acceptanceCriteria:
          - "credenciales OAuth configuradas"
```

### Campos

**Milestone (Hito):**
| Campo | Requerido | Descripción |
|-------|-----------|-------------|
| `id` | Sí | Identificador único entre todos los hitos |
| `name` | Sí | Nombre legible |
| `dependsOn` | No | IDs de hitos que deben completarse primero |
| `steps` | Sí | Lista ordenada de pasos |

**Step (Paso):**
| Campo | Requerido | Descripción |
|-------|-----------|-------------|
| `id` | Sí | Identificador único dentro del hito |
| `instruction` | Sí | Acción a realizar |
| `dependsOn` | No | IDs de pasos que deben completarse primero |
| `projectParts` | No | Partes del proyecto relevantes para este paso |
| `acceptanceCriteria` | No | Condiciones que definen "completado" |
| `tier` | No | `ai` (por defecto) o `human` |

### Valores de Tier

- **ai**: El paso puede ser ejecutado completamente por un agente de IA de forma autónoma
- **human**: El paso requiere acción humana (aprobaciones, configuración externa, tareas físicas)

## Formato de Salida

```
# Milestone: API de Autenticación

research: use a subagent to investigar mejores prácticas de JWT [done when: estrategia de expiración de tokens documentada]
implement: Analyze server (/ruta/al/servidor) then read results from steps research then implementar endpoints de autenticación [done when: endpoints de login y logout funcionando; tests unitarios pasando]

---

# Milestone: UI de Autenticación (depends on: auth_api)

build_form: use a subagent to Analyze client (/ruta/al/cliente) then implementar formulario de login
configure_oauth: [HUMAN] read results from steps build_form then configurar proveedor OAuth en producción [done when: credenciales OAuth configuradas]
```

**Componentes de salida:**
- Prefijo `[HUMAN]` marca pasos que requieren acción humana
- Prefijo `use a subagent to` marca pasos de IA paralelizables
- `Analyze <nombre> (<ruta>)` proporciona contexto del código
- `read results from steps X, Y` encadena pasos dependientes
- `[done when: ...]` define criterios de completitud

## Cómo Funciona

1. Analiza `projectParts` y `milestones` desde YAML
2. Ordena topológicamente los hitos por dependencias (algoritmo de Kahn)
3. Para cada hito:
   - Construye un DAG a partir de las dependencias de los pasos
   - Ordena topológicamente los pasos
   - Genera fragmentos de prompt con referencias resueltas
4. Devuelve los prompts concatenados por hito
