# API Security Scanner

A small Go-based tool to test API endpoints for common issues (IDOR, JWT misconfigurations) and run Nuclei checks.

**Features**
- CLI flags:
  - `-api` : path to OpenAPI JSON (default: `endpoints_and_passwords/API.json`).
  - `-url` : override base URL from the OpenAPI file (e.g. `http://localhost:8888`).
  - `-type`: vulnerability type (1 = IDOR, 2 = JWT). If omitted the tool will prompt for selection.
- IDOR testing:
  - Attempts to detect login and obtain an auth token. If found, runs IDOR checks with Authorization header.
  - If no token is found, runs unauthenticated IDOR checks.
- JWT testing:
  - Logs in (using `emails.txt` and `passwords.txt`) and tests for `alg: none` and JWT confusion attacks.
- Report generation:
  - Results are saved to `report.json` in the repository root.
- Nuclei integration:
  - The project will invoke `nuclei` after scanning (requires `nuclei` installed and on `PATH`). Output saved to `nuclei-report.json`.
- File fallbacks:
  - Credential and supporting files are expected under `endpoints_and_passwords/` (the code also falls back to that folder if files are not found in CWD).

**Required files (examples)**
- `endpoints_and_passwords/API.json` — OpenAPI file used to read `servers[0].url` and `paths`.
- `endpoints_and_passwords/emails.txt` — candidate email(s) for login attempts (one per line).
- `endpoints_and_passwords/passwords.txt` — candidate password(s) (one per line).
- `endpoints_and_passwords/apiendpoints_read.txt` — additional endpoints read by discovery.
- `endpoints_and_passwords/JWKSet_endpoints.txt` — list of JWK set endpoint paths to check.

**Requirements**
- Go 1.24+ (to build/run)
- Optionally: `nuclei` (CLI) if you want CVE/tech checks

**Quick usage**
- Interactive mode (prompts for vuln type):

```bash
go run .
```

- Non-interactive examples:

```bash
# Run JWT tests using API.json default servers URL
go run . -type 2

# Run IDOR tests against a specific host/port
go run . -type 1 -url http://localhost:8888

# Use a custom API file
go run . -api path/to/my_api.json -type 2
```

**Output**
- `report.json` — JSON array of `issue`, `endpoint`, and optional `token` fields.
- `nuclei-report.json` — output from `nuclei` (if `nuclei` ran).

**Notes & tips**
- If your service runs on `localhost` with a specific port, either set that port in `endpoints_and_passwords/API.json` `servers[0].url` or pass `-url`.
- The tool tries credentials from `emails.txt`/`passwords.txt`; if you have a single test account add it there.
- To skip or customize Nuclei runs, edit `main.go` and modify or remove the `nuclei.RunNuclei(...)` call.

## Tested Environment

This project was tested in a controlled lab environment using:

- OWASP crAPI (for setup refer : https://github.com/OWASP/crAPI/blob/develop/docs/setup.md )
- Local Spring Boot applications
- Authenticated REST APIs

## Example Findings


1. [JWT]
    Algorithm: RS256
    Issue : Possible JWT Confusion Attack Vulnerability
    Endpoint : http://localhost:8888/identity/api/v2/user/dashboard 

2. [IDOR]
    Isuue:Possible authorization issue detected on:
    Endpoint : http://localhost:8888/api/orders/{id}

3. [TECH]
    Info : OpenResty detected


