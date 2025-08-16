# Catalyst Engine
A prototype Go backend for accelerating the development of modern, data centric applications
`catalyst` is a boilerplate, an accelerator, and architectural pattern all rolled into one.  It's my answer to setting up new projects to do the same thing: ingest data, manage it, and expose it through an API. This is a config-driven, modular monolith designed to get you from zero to a deployed, valuable application at ludicrous speed. 
This platform was born from the need to build a powerful, specific application in public without exposing the sensitive business logic, which forced a hard separation between the generic "engine" and the application-specific "configuration". The result is a reusable core than you can wrap your own business logic around. 

## Core Principles
This isn't just a random pile of code. Its built on a few key ideas. 
- **Modular Monolith First:** We aren't creating microservices until its necessary. The backend is a single, deployable Go application, but it's build from highly modular, decoupled packages. This maximizes your initial velocity and keeps th local dev experience sane. When a module needs to scale independently or use a different tech stack, you have a clear path to extract it. 
- **Config-Driven Design:** The engine is generic on purpose. You define your application's datastructures, ingestion rules, and business logic in YAML configuration files. New data pipelines and even entire applications are onboarded by adding configs, not by forking the core platform code. 
- **API-First and Decoupled:** The Go backend exposes a clean, versioned API that serves a modern, client-side React frontend. Clear separation of concerns from day one. 

## How it Works: The Core Components
### The database Heart: A Generic `items` Table
At the center of the architecture is a single powerful `items` table designed for flexibility.
- `item_type`: A simple string that defines what kind of data a row represents (e.g., 'USER_PROFILE', 'KNOWLEDGE_CHUNK', 'INVOICE'). This is your primary application discriminator.
- `scope`: A generic, indexed column for your main business-level filter, like a region or a business line. 
- `custum_properties` **(JSONB): This is where the magic happens. All your application-specific data lives here, giving you a flexible, schema-on-read model without sacrificing the power of Postgres. 
- `embedding` **(vector)** Built with `pgvector` from the start, making your data AI-ready for semantic search and RAG applications out of the box. 

## The Ingestion Engine
`catalyst` includes a robust, config-driven ETL pipeline for getting data into the system.  You define a YAML file that maps your source data (like a CSV) to the `items` table structure, and the engine handles the "Stage, Validate, Load" process, complete with error logging and a triage queue for records that fail validation.

## Technology Stack
No exotic stuff. Just solid, modern tech that gets the job done
**Backend**
- **Language:** Go
- **Framework:** Echo
- **Database:** PostgreSQL with the pgvector extension
- **Migrations:** goose
- **Queries:** sqlc for type-safe, generated query code

**Frontend**
- Framework:** React with TypeScript
- **Build Tool:** Vite
- **Styling:** Tailwind CSS
- **Components:** `shadcn/ui` (built on Radix UI)

## Roadmap
This is a prototype, but the vision is clear.
1. **Foundation:** Solidify the core ingestion and API engine. Refine the database schema and add robust testing. 
2. **Acceleration:** Build our more "force-multiplier" features, like a generic admin UI for managing users and view triage queues
3. **Extensibility:** Improve the plugin architecture to make it even easier to add new application modules and custom logic without touching the core. 
