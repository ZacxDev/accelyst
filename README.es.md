# Accelyst

Una interfaz de iteración de proyectos nativa para IA.

## Descripción

Accelyst transforma la estructura de tu proyecto y el backlog de funcionalidades en prompts listos para ejecución con el contexto completo del código. Cada prompt incluye los directorios relevantes, rutas de documentación, orden de dependencias y contexto de artefactos—todo lo que un agente de IA necesita para implementar funcionalidades de forma autónoma.

- **Contexto enriquecido**: Los prompts referencian directorios exactos y documentación
- **Consciente de dependencias**: Hitos y pasos se ordenan topológicamente via DAG
- **Listo para paralelismo**: El trabajo independiente se marca para ejecución concurrente con subagentes
- **Persistencia de conocimiento**: Los artefactos de análisis se acumulan entre iteraciones
- **Múltiples flujos de trabajo**: Soporte para épicas concurrentes

## Instalación

```bash
go build -o accelyst
```

## Inicio Rápido

```bash
# Inicializar un nuevo proyecto
./accelyst init

# Listar todas las épicas
./accelyst epics list

# Ejecutar un hito
./accelyst run <epica>/<hito>

# Ejecutar un paso específico
./accelyst run <epica>/<hito>/<paso>
```

## Arquitectura

Accelyst utiliza un modelo de configuración dividido:

```
proyecto/
├── .accelyst/
│   ├── registry.yaml      # Persistente: partes del proyecto + artefactos
│   ├── epics/             # Efímero: colecciones de hitos
│   │   ├── feature-a.yaml
│   │   └── feature-b.yaml
│   └── artifacts/         # Salidas de conocimiento
│       └── analisis.md
```

- **Registro** (`.accelyst/registry.yaml`): Conocimiento persistente del proyecto incluyendo componentes y artefactos de análisis
- **Épicas** (`.accelyst/epics/*.yaml`): Colecciones de hitos desechables que pueden regenerarse
- **Artefactos** (`.accelyst/artifacts/*.md`): Salidas de análisis que persisten entre iteraciones

## Formato del Registro

```yaml
version: 1

projectParts:
  - name: client
    directoryAbs: /ruta/al/cliente
    description: "Aplicación frontend"
  - name: api
    directoryAbs: /ruta/a/api
    documentationAbs: /ruta/a/docs/api.md
    description: "API backend"
  - name: test
    directoryAbs: /ruta/a/tests
    description: "Suites de pruebas"

artifacts:
  - id: auth_flow
    description: "Análisis del flujo de autenticación"
    producedBy: auth-feature/jwt_auth/analyze_auth
    createdAt: "2024-01-08"
```

## Formato de Épica

```yaml
name: "Nombre de la Funcionalidad"
description: "Lo que logra esta épica"
priority: ready  # ready | in_progress | blocked | ideas

milestones:
  - id: auth_api
    name: "API de Autenticación"
    dependsOn: []
    steps:
      - id: analyze_auth
        instruction: "analizar patrones de autenticación existentes"
        produces: auth_analysis  # Crea artefacto
        projectParts:
          - api

      - id: implement
        instruction: "implementar endpoints de autenticación"
        requires:
          - auth_analysis  # Usa artefacto
        dependsOn:
          - analyze_auth
        projectParts:
          - api

      - id: test_auth
        instruction: "escribir tests de autenticación"
        dependsOn:
          - implement
        projectParts:
          - test

  - id: auth_ui
    name: "UI de Autenticación"
    dependsOn:
      - auth_api
    steps:
      - id: build_form
        instruction: "implementar formulario de login"
        requires:
          - auth_analysis  # Reutiliza artefacto del hito anterior
        projectParts:
          - client
```

### Campos de Step

| Campo | Requerido | Descripción |
|-------|-----------|-------------|
| `id` | Sí | Identificador único dentro del hito |
| `instruction` | Sí | Acción a realizar |
| `dependsOn` | No | IDs de pasos que deben completarse primero |
| `projectParts` | No | Partes del proyecto relevantes para este paso |
| `produces` | No | ID del artefacto que este paso crea |
| `requires` | No | IDs de artefactos que este paso necesita como contexto |

## Sistema de Artefactos

Los pasos pueden producir y consumir artefactos para transferencia de conocimiento:

**Producción**: Un paso con `produces: auth_analysis`:
1. Incluye una instrucción de guardado en el prompt
2. El agente escribe en `.accelyst/artifacts/auth_analysis.md`
3. El artefacto se registra para uso futuro

**Consumo**: Un paso con `requires: [auth_analysis]`:
1. Carga el contenido del artefacto desde disco
2. Lo inyecta en el contexto del prompt
3. El agente recibe el análisis previo como contexto

Esto permite la acumulación de conocimiento entre hitos e iteraciones.

## Referencia de CLI

```bash
# Inicialización del proyecto
accelyst init              # Crear estructura .accelyst/
accelyst migrate           # Migrar desde accelyst.yaml legado

# Gestión de épicas
accelyst epics list        # Listar todas las épicas y hitos

# Ejecución
accelyst run <epica>/<hito>         # Generar prompt del hito
accelyst run <epica>/<hito>/<paso>  # Generar prompt de un paso

# Inspección del registro
accelyst registry parts         # Listar partes del proyecto
accelyst registry artifacts     # Listar artefactos registrados
accelyst registry add-artifact  # Registrar un artefacto
```

## Formato de Salida

```
# Milestone: API de Autenticación

## Step: analyze_auth

use a subagent to Analyze api (/ruta/a/api) then analizar patrones de autenticación existentes

---
Save analysis to: .accelyst/artifacts/auth_analysis.md

---

## Step: implement

### Context: auth_analysis

[contenido de auth_analysis.md]

---

Analyze api (/ruta/a/api) then read results from steps analyze_auth then implementar endpoints de autenticación
```

**Componentes de salida:**
- Prefijo `use a subagent to` marca pasos paralelizables (sin dependencias)
- `Analyze <nombre> (<ruta>)` proporciona contexto del código
- Secciones `### Context:` inyectan artefactos requeridos
- `read results from steps X, Y` encadena pasos dependientes
- `Save analysis to:` instruye la creación de artefactos

## Migración

Los proyectos que usan el formato de archivo único legado pueden migrar:

```bash
./accelyst migrate
```

Esto:
1. Crea la estructura del directorio `.accelyst/`
2. Extrae `projectParts` a `.accelyst/registry.yaml`
3. Mueve los hitos a `.accelyst/epics/default.yaml`
4. Respalda el archivo original a `accelyst.yaml.bak`

## Cómo Funciona

1. Carga el registro y escanea el directorio de épicas
2. Valida todas las referencias (partes del proyecto, artefactos, dependencias)
3. Ordena topológicamente los hitos por dependencias (algoritmo de Kahn)
4. Para cada hito:
   - Construye un DAG a partir de las dependencias de los pasos
   - Ordena topológicamente los pasos
   - Inyecta contenido de artefactos para `requires`
   - Genera fragmentos de prompt con referencias resueltas
   - Agrega instrucciones de guardado para `produces`
5. Devuelve los prompts concatenados por hito

## Skills de Claude Code

Accelyst incluye skills de Claude Code para gestión automatizada de hitos:

| Skill | Propósito |
|-------|-----------|
| `/milestone-orchestrate` | Ciclo completo: backlog → funcionalidades implementadas |
| `/milestone-plan` | Descubrimiento interactivo → brief de planificación |
| `/milestone-convert` | Convertir brief → YAML de épica |
| `/milestone-execute` | Ejecutar épica via subagentes |
| `/accelyst-guide` | Operaciones CLI y solución de problemas |

```bash
# Automatización completa desde archivo scratch
/milestone-orchestrate scratch.txt

# O ejecutar pasos individuales
/milestone-plan scratch.txt
/milestone-convert planning-brief.yaml
/milestone-execute mi-epica/mi-hito
```

## Documentación

- [SPEC.md](SPEC.md) - Especificación completa con esquemas JSON
- [.claude/skills/](.claude/skills/) - Definiciones de skills de Claude Code
