# De 0 a desarrollador con IA: código

Este repositorio tiene todo el código del libro "De 0 a desarrollador con IA": los ejemplos de cada capítulo y los cinco proyectos guiados del final. Las rutas que cita el libro, como `codigo/proyectos/47-api-reservas/`, son las mismas que ves aquí.

Todo lo que hay se ha ejecutado. Si algo no te funciona, abre una issue con el capítulo, tu sistema operativo y el error que ves.

## Qué necesitas

No hace falta instalarlo todo de golpe. Cada parte del libro te dice qué usa.

| Herramienta | Versión | Dónde se usa |
|---|---|---|
| Python | 3.13 | Partes II, V y VI, capítulos 29, 30, 41 y 45, proyecto 51 |
| Go | 1.27 | Partes III a IX y proyectos 47 y 48 |
| Node.js | 26 | Capítulo 36 y proyecto 49 |
| Docker | cualquiera reciente | Capítulos 37 y 38, proyecto 47 con PostgreSQL |
| JDK | 17 | Proyecto 50 (Gradle se descarga solo con `gradlew`) |

## Cómo ejecutar el código

- **Python.** Entra en la carpeta del capítulo y lanza el script, por ejemplo `python main.py`. Para los tests, `python -m unittest discover`.
- **Go.** Cada capítulo es un módulo. Dentro de su carpeta, `go test ./...` y `go run .` (o `go run ./cmd/api` en los proyectos).
- **Web (proyecto 49).** `npm ci` y después `npm run dev`. Con `npm run dev:simulada` funciona sin la API del proyecto 47.
- **App móvil (proyecto 50).** `./gradlew :shared:jvmTest` para los tests y Android Studio para abrir la app. La parte de iOS necesita un Mac con Xcode.

Los servidores escuchan en `127.0.0.1` por defecto. Solo dentro de un contenedor usan `0.0.0.0`.

## Claves de API y modo simulado

Los ejemplos que llaman a un modelo de IA leen la clave de una variable de entorno. Si no la defines, funcionan en **modo simulado**, con respuestas de ejemplo y sin red, así que puedes seguir el libro sin gastar nada.

Cada carpeta que lo necesita trae un `.env.example` con valores de mentira. Cópialo como `.env` y pon los tuyos. El `.env` está en `.gitignore` y no se sube nunca.

## Índice

### Parte II. Fundamentos de programación

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 4 | Pensamiento computacional | [`codigo/parte-02/04-pensamiento-computacional`](codigo/parte-02/04-pensamiento-computacional) | Python / pseudocódigo |
| 5 | El ecosistema del software moderno | [`codigo/parte-02/05-el-ecosistema-del-software`](codigo/parte-02/05-el-ecosistema-del-software) | Ejemplos de texto |
| 6 | Programar: los conceptos base | [`codigo/parte-02/06-programar-los-conceptos-base`](codigo/parte-02/06-programar-los-conceptos-base) | Python |
| 7 | Lenguajes de programación | [`codigo/parte-02/07-lenguajes-de-programacion`](codigo/parte-02/07-lenguajes-de-programacion) | Python (comparaciones breves) |
| 8 | Sistemas y redes | [`codigo/parte-02/08-sistemas-y-redes`](codigo/parte-02/08-sistemas-y-redes) | terminal |
| 9 | Buenas prácticas desde el principio | [`codigo/parte-02/09-buenas-practicas-desde-el-principio`](codigo/parte-02/09-buenas-practicas-desde-el-principio) | Python, Git |
| 10 | Tu primer proyecto | [`codigo/parte-02/10-tu-primer-proyecto`](codigo/parte-02/10-tu-primer-proyecto) | Python |

### Parte III. Ingeniería de software

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 12 | Requisitos: entender qué hay que construir | [`codigo/parte-03/12-requisitos`](codigo/parte-03/12-requisitos) | Ejemplos de texto |
| 13 | Desarrollo guiado por especificaciones | [`codigo/parte-03/13-spec-driven-development`](codigo/parte-03/13-spec-driven-development) | Markdown, Gherkin |
| 14 | Paradigmas de programación | [`codigo/parte-03/14-paradigmas-de-programacion`](codigo/parte-03/14-paradigmas-de-programacion) | Python, y Go como contraste |
| 15 | Principios de diseño | [`codigo/parte-03/15-principios-de-diseno`](codigo/parte-03/15-principios-de-diseno) | Python y Go |

### Parte IV. Arquitectura

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 16 | Por qué importa la arquitectura | [`codigo/parte-04/16-por-que-arquitectura`](codigo/parte-04/16-por-que-arquitectura) | Go |
| 17 | Arquitectura limpia y hexagonal en la práctica | [`codigo/parte-04/17-arquitectura-limpia-en-la-practica`](codigo/parte-04/17-arquitectura-limpia-en-la-practica) | Go |
| 18 | Un caso guiado y un reto | [`codigo/parte-04/18-caso-guiado-y-reto`](codigo/parte-04/18-caso-guiado-y-reto) | Go |
| 19 | Sistemas distribuidos y eventos | [`codigo/parte-04/19-sistemas-distribuidos-y-eventos`](codigo/parte-04/19-sistemas-distribuidos-y-eventos) | Go |

### Parte V. Fundamentos de IA

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 21 | IA generativa y modelos de lenguaje | [`codigo/parte-05/21-ia-generativa-y-llm`](codigo/parte-05/21-ia-generativa-y-llm) | Python |
| 22 | Prompting y multimodalidad | [`codigo/parte-05/22-prompting-y-multimodalidad`](codigo/parte-05/22-prompting-y-multimodalidad) | Python |
| 23 | Frameworks y formas de adaptar un modelo | [`codigo/parte-05/23-frameworks-de-ia`](codigo/parte-05/23-frameworks-de-ia) | Python |

### Parte VI. Desarrollar con agentes

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 26 | Contexto, capacidades y extensiones (MCP, skills, reglas) | [`codigo/parte-06/26-contexto-capacidades-y-extensiones`](codigo/parte-06/26-contexto-capacidades-y-extensiones) | Markdown, Python |
| 27 | Integrar APIs de IA en tus programas | [`codigo/parte-06/27-integrar-apis-de-ia`](codigo/parte-06/27-integrar-apis-de-ia) | Python |
| 28 | Orquestar varios agentes | [`codigo/parte-06/28-orquestacion-de-agentes`](codigo/parte-06/28-orquestacion-de-agentes) | Python |

### Parte VII. Calidad

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 29 | Testing, con y sin IA | [`codigo/parte-07/29-testing`](codigo/parte-07/29-testing) | Go y Python |
| 30 | Code smells, refactorización y deuda técnica | [`codigo/parte-07/30-code-smells-refactor-y-deuda`](codigo/parte-07/30-code-smells-refactor-y-deuda) | Go y Python |
| 31 | Métricas: cobertura y complejidad | [`codigo/parte-07/31-metricas-de-calidad`](codigo/parte-07/31-metricas-de-calidad) | Go |
| 32 | Revisión de código con IA | [`codigo/parte-07/32-revision-de-codigo-con-ia`](codigo/parte-07/32-revision-de-codigo-con-ia) | Ejemplos de texto |
| 33 | Documentación con IA | [`codigo/parte-07/33-documentacion`](codigo/parte-07/33-documentacion) | Markdown, Mermaid |
| 34 | Observabilidad | [`codigo/parte-07/34-observabilidad`](codigo/parte-07/34-observabilidad) | Go |
| 35 | Usabilidad | [`codigo/parte-07/35-usabilidad`](codigo/parte-07/35-usabilidad) | Ejemplos de texto |
| 36 | Dar ojos a los agentes (y calidad en tiempos de agentes) | [`codigo/parte-07/36-ojos-para-los-agentes`](codigo/parte-07/36-ojos-para-los-agentes) | TypeScript (Playwright) |

### Parte VIII. Infraestructura

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 37 | Bases de datos | [`codigo/parte-08/37-bases-de-datos`](codigo/parte-08/37-bases-de-datos) | SQL, Go |
| 38 | Contenedores | [`codigo/parte-08/38-contenedores`](codigo/parte-08/38-contenedores) | Docker |
| 39 | Integración y despliegue continuos | [`codigo/parte-08/39-ci-cd`](codigo/parte-08/39-ci-cd) | YAML |
| 41 | Llevar la IA a producción (LLMOps) | [`codigo/parte-08/41-llmops`](codigo/parte-08/41-llmops) | Python |

### Parte IX. Seguridad

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 42 | Desarrollo seguro | [`codigo/parte-09/42-desarrollo-seguro`](codigo/parte-09/42-desarrollo-seguro) | Go |
| 43 | Las vulnerabilidades web más comunes (OWASP Top 10) | [`codigo/parte-09/43-owasp-web`](codigo/parte-09/43-owasp-web) | Go |
| 44 | Seguridad en APIs | [`codigo/parte-09/44-owasp-api`](codigo/parte-09/44-owasp-api) | Go |
| 45 | Seguridad en aplicaciones con LLM | [`codigo/parte-09/45-owasp-llm`](codigo/parte-09/45-owasp-llm) | Python |

### Parte X. Proyectos guiados

| Cap. | Título | Carpeta | Lenguaje |
|---|---|---|---|
| 47 | Proyecto: la API de reservas | [`codigo/proyectos/47-api-reservas`](codigo/proyectos/47-api-reservas) | Go |
| 48 | Proyecto: un asistente de terminal con IA | [`codigo/proyectos/48-cli-ia`](codigo/proyectos/48-cli-ia) | Go |
| 49 | Proyecto: la web de reservas | [`codigo/proyectos/49-web-reservas`](codigo/proyectos/49-web-reservas) | TypeScript |
| 50 | Proyecto: la app móvil con Kotlin Multiplatform | [`codigo/proyectos/50-app-movil`](codigo/proyectos/50-app-movil) | Kotlin |
| 51 | Proyecto: un asistente que responde con tus documentos | [`codigo/proyectos/51-asistente-rag`](codigo/proyectos/51-asistente-rag) | Python |

## Ediciones

Cada edición del libro tiene una etiqueta en este repositorio. Si estás leyendo una edición concreta, usa su etiqueta para que el código coincida con lo que lees.

## Licencia

El código se publica con licencia MIT (mira `LICENSE`). El texto del libro no forma parte de este repositorio.
