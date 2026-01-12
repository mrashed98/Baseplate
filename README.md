# Baseplate

**Baseplate** is a "Headless" Backend Engine that allows defining data entities via **Blueprints** (schemas) rather than hard-coded structs.

Inspired by the internal architecture of tools like **Port.io** and **n8n**, this engine allows users to register new resources (e.g., "Service", "Cluster", "Employee") at runtime and instantly get validated CRUD APIs for them without database migrations or code changes.

Built with **Go (Gin)** and **PostgreSQL (JSONB)**.

## 🚀 Getting Started

### Prerequisites
- [Go 1.23+](https://go.dev/dl/)
- [Docker & Docker Compose](https://www.docker.com/)

### Running the Project
1. **Clone the repo:**
   ```bash
   git clone <repo_url>
   cd Baseplate
   ```
   