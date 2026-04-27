# ClearFly UI Testing Skill

## Overview
How to test the ClearFly React frontend (services/webui) against the full microservices backend.

## Devin Secrets Needed
None required — demo credentials are built into the app.

## Environment Setup

### Start Full Stack
```bash
cd /home/ubuntu/repos/ClearFly/docker
docker compose up -d --build
# First build takes ~5 min (8 Go services + Vite frontend)
# Subsequent builds are faster due to Docker cache
```

### Seed Demo Data
Wait for gateway to be healthy first:
```bash
curl -s http://localhost:8080/health  # should return {"status":"ok"}
python3 scripts/seed_demo.py
```
This creates 18 passengers, 5 flights with varied load factors, 9 baggage tags, and 1 incident.

### Verify Services
| Service | URL | Purpose |
|---------|-----|--------|
| Frontend | http://localhost:3000 | React SPA (nginx) |
| Gateway | http://localhost:8080 | API gateway |
| RabbitMQ | http://localhost:15672 | Management UI (guest/guest) |

## Login Credentials
- **Admin**: `admin` / `admin` — sees all pages including Operations and Analytics
- **Passenger**: Register via /register — sees passenger-specific pages only
- Login hint is shown on the login page itself

## Key Pages to Test
1. **Login** (`/login`) — auth card, gradient logo, cyan button
2. **Dashboard** (`/`) — KPI cards, flight list, event feed, sidebar navigation
3. **Flight Search** (`/search`) — search form, flight cards grid, hover effects
4. **Analytics** (`/analytics`) — gauge, KPI cards, flight list, pricing rules (click different flights to see gauge change)
5. **Notifications** (`/notifications`) — notification cards with tone-colored borders
6. **Baggage** (`/baggage`) — 6-stage timeline, baggage list, "Следующий скан" button
7. **Operations** (`/operations`) — incident form, flight list (admin only)
8. **Profile** (`/profile`) — passenger search/registration form (admin view)

## Architecture Notes
- Frontend connects to gateway at `localhost:8080` when served from port 3000 (see `api.js` line 11)
- No vite proxy configured — full docker compose stack required for API calls
- The webui container builds and serves via nginx (port 3000 → container port 80)
- CSS design tokens are in `services/webui/src/index.css` as `:root` custom properties

## Teardown
```bash
cd /home/ubuntu/repos/ClearFly/docker
docker compose down     # stop (keeps data)
docker compose down -v  # stop and wipe DB
```

## Tips
- Gateway health check may take 15-30s after containers start — wait before seeding
- The `make reseed` command combines down -v + up + seed in one step (Linux/macOS)
- All flight data is date-relative, so demo data works on any day
- Analytics page is most visually interesting when clicking between low-load (CN101 ~20%) and high-load (CN318 ~93%) flights
