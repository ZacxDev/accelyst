# Accelyst

Una herramienta estructurada de evolución de proyectos que genera prompts de ejecución para que agentes de IA implementen funcionalidades de forma autónoma a partir de descripciones de alto nivel.

## Descripción

Accelyst analiza la estructura del proyecto y las definiciones de hitos, luego genera prompts ordenados topológicamente. Los pasos sin dependencias se marcan para ejecución paralela mediante subagentes.

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
  - name: "Nombre de la Funcionalidad"
    steps:
      - id: 1
        instruction: "investigar la documentación"
        dependsOnStepIds: []
        projectParts: []
      - id: 2
        instruction: "analizar la implementación existente"
        dependsOnStepIds: []
        projectParts:
          - client
      - id: 3
        instruction: "implementar la funcionalidad"
        dependsOnStepIds:
          - 1
          - 2
        projectParts:
          - client
          - server
```

## Formato de Salida

```
# Milestone: Nombre de la Funcionalidad

1: use a subagent to investigar la documentación
2: use a subagent to Analyze client (/ruta/al/cliente) then analizar la implementación existente
3: Analyze client (/ruta/al/cliente), server (/ruta/al/servidor) then read results from steps 1, 2 then implementar la funcionalidad
```

- Pasos sin dependencias reciben el prefijo `use a subagent to` (paralelizables)
- Pasos con dependencias incluyen `read results from steps X, Y`
- Pasos con partes del proyecto incluyen `Analyze <nombre> (<ruta>) (docs: <ruta>)`

## Cómo Funciona

1. Analiza `projectParts` y `milestones` desde YAML
2. Para cada hito:
   - Construye un DAG a partir de las dependencias de los pasos
   - Ordena topológicamente los pasos (algoritmo de Kahn)
   - Genera fragmentos de prompt con referencias resueltas
3. Devuelve los prompts concatenados por hito
