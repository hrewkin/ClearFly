#!/usr/bin/env python3
"""
ClearFly demo data seeder.

Generates a realistic slice of demo data by calling the public API gateway:
  * 15+ passengers with varied loyalty tiers, meal preferences, and special needs
  * bookings on the upcoming flights so the analytics dashboard shows all three
    pricing tiers (≤50%, 50–80%, >80% load factor)
  * baggage tags at different stages of the CHECKED_IN → CLAIMED journey
  * one FLIGHT_DELAYED incident so the operations centre and notifications
    feed are not empty when the demo starts

Runs against docker-compose's gateway at http://localhost:8080 by default.
Stdlib-only (urllib) so no pip install required.

Usage:
    python3 scripts/seed_demo.py
    python3 scripts/seed_demo.py --base-url http://host.docker.internal:8080
    python3 scripts/seed_demo.py --passengers 30 --seed 1337
"""
from __future__ import annotations

import argparse
import json
import random
import sys
import time
import urllib.error
import urllib.request
from typing import Any

# ---------------------------------------------------------------------------
# Deterministic demo fixtures
# ---------------------------------------------------------------------------

RU_FIRST_NAMES = [
    "Анна", "Михаил", "Елена", "Дмитрий", "Ольга", "Алексей", "Мария",
    "Иван", "София", "Николай", "Екатерина", "Сергей", "Виктория",
    "Артём", "Полина", "Павел", "Алиса", "Денис", "Ксения", "Роман",
    "Юлия", "Игорь", "Дарья", "Фёдор",
]
RU_LAST_NAMES = [
    "Иванов", "Смирнов", "Кузнецов", "Попов", "Васильев", "Петров",
    "Соколов", "Михайлов", "Новиков", "Фёдоров", "Морозов", "Волков",
    "Алексеев", "Лебедев", "Семёнов", "Егоров", "Павлов", "Козлов",
    "Степанов", "Николаев", "Орлов", "Макаров", "Никитин", "Захаров",
]

LOYALTY_TIERS = ["STANDARD", "STANDARD", "SILVER", "SILVER", "GOLD", "PLATINUM"]
MEAL_PREFS = [
    "STANDARD", "VEGETARIAN", "VEGAN", "HALAL", "KOSHER",
    "GLUTEN_FREE", "DIABETIC",
]
SPECIAL_NEEDS = [
    "NONE", "NONE", "NONE",  # bias toward none
    "WHEELCHAIR", "EXTRA_LEGROOM", "INFANT", "UNACCOMPANIED_MINOR",
]

# ---------------------------------------------------------------------------
# Tiny HTTP client on top of urllib so the script has no third-party deps.
# ---------------------------------------------------------------------------

class ApiError(RuntimeError):
    pass


class Api:
    def __init__(self, base_url: str, timeout: float = 5.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def _request(self, method: str, path: str, body: Any = None) -> Any:
        url = f"{self.base_url}/api/v1{path}"
        data = None
        headers = {"Accept": "application/json"}
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=data, method=method, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read()
                if not raw:
                    return None
                return json.loads(raw)
        except urllib.error.HTTPError as exc:
            raise ApiError(f"{method} {url} → {exc.code}: {exc.read().decode(errors='ignore')}")
        except urllib.error.URLError as exc:
            raise ApiError(f"{method} {url} → {exc.reason}")

    def get(self, path: str) -> Any:
        return self._request("GET", path)

    def post(self, path: str, body: Any) -> Any:
        return self._request("POST", path, body)

    def patch(self, path: str, body: Any) -> Any:
        return self._request("PATCH", path, body)


# ---------------------------------------------------------------------------
# Seeders
# ---------------------------------------------------------------------------

def wait_for_gateway(api: Api, attempts: int = 30) -> None:
    for i in range(attempts):
        try:
            # /health is exposed on the gateway root, not /api/v1.
            with urllib.request.urlopen(f"{api.base_url}/health", timeout=2) as resp:
                if resp.status == 200:
                    return
        except Exception:
            pass
        time.sleep(1)
    raise ApiError(f"gateway at {api.base_url} is not responding after {attempts}s")


def seed_passengers(api: Api, count: int, rnd: random.Random) -> list[dict]:
    print(f"→ Creating {count} passengers…")
    passengers: list[dict] = []
    used_emails: set[str] = set()

    for i in range(count):
        first = rnd.choice(RU_FIRST_NAMES)
        last = rnd.choice(RU_LAST_NAMES)
        # Disambiguate the email so repeat runs don't collide.
        suffix = f"{i:02d}"
        email = f"{_translit(first).lower()}.{_translit(last).lower()}{suffix}@clearfly.demo"
        while email in used_emails:
            suffix = f"{rnd.randint(0, 9999):04d}"
            email = f"{_translit(first).lower()}.{_translit(last).lower()}{suffix}@clearfly.demo"
        used_emails.add(email)

        try:
            p = api.post("/passengers", {
                "name": f"{last} {first}",
                "email": email,
                "phone": f"+7 9{rnd.randint(10, 99)} {rnd.randint(100, 999)} {rnd.randint(10, 99)} {rnd.randint(10, 99)}",
                "passport_number": f"{rnd.randint(4000, 4999)} {rnd.randint(100000, 999999)}",
            })
        except ApiError as exc:
            print(f"  ! failed to create {last} {first}: {exc}")
            continue

        pref_payload = {
            "loyalty_tier": rnd.choice(LOYALTY_TIERS),
            "meal_preference": rnd.choice(MEAL_PREFS),
            "special_needs": rnd.choice(SPECIAL_NEEDS),
        }
        try:
            api.patch(f"/passengers/{p['id']}/preferences", pref_payload)
            p.update(pref_payload)
        except ApiError as exc:
            print(f"  ! failed to set prefs on {p['id']}: {exc}")
        passengers.append(p)

    print(f"  ✓ {len(passengers)} passengers created")
    return passengers


def book_seats_on_flight(api: Api, flight: dict, passengers: list[dict], target_pct: float, rnd: random.Random) -> int:
    """Book up to `target_pct` of the flight's seats. Returns bookings made."""
    seats = api.get(f"/flights/{flight['id']}/seats") or []
    available = [s for s in seats if s.get("status") == "AVAILABLE"]
    want = int(len(seats) * target_pct / 100)
    to_book = available[:want]
    rnd.shuffle(to_book)
    made = 0
    for seat in to_book:
        passenger = rnd.choice(passengers)
        try:
            api.post("/bookings/book", {
                "flight_id": flight["id"],
                "passenger_id": passenger["id"],
                "seat_id": seat["id"],
            })
            made += 1
        except ApiError:
            # Seat was grabbed by a parallel request or has a tariff mismatch;
            # just keep going.
            continue
    return made


def seed_bookings(api: Api, passengers: list[dict], rnd: random.Random) -> None:
    print("→ Generating bookings across upcoming flights…")
    flights = api.get("/flights/upcoming") or []
    if not flights:
        print("  ! no upcoming flights returned; is booking service seeded?")
        return

    # One flight per pricing tier, so the analytics page shows all three
    # multipliers (×1.0 / ×1.2 / ×1.5).
    tiers = [20, 65, 95]
    for i, flight in enumerate(flights[:3]):
        pct = tiers[i % len(tiers)]
        made = book_seats_on_flight(api, flight, passengers, pct, rnd)
        print(f"  ✓ {flight['flight_number']} → {made} bookings (~{pct}% target)")

    # Remaining flights get a lightweight 10–30% load so they're not empty.
    for flight in flights[3:]:
        made = book_seats_on_flight(api, flight, passengers, rnd.randint(10, 30), rnd)
        print(f"  ✓ {flight['flight_number']} → {made} bookings (light)")


def seed_baggage(api: Api, passengers: list[dict], rnd: random.Random) -> None:
    print("→ Registering baggage tags at different stages…")
    flights = api.get("/flights/upcoming") or []
    if not flights or not passengers:
        return

    # (target_stage_index, count) — 0 is CHECKED_IN, 5 is CLAIMED.
    buckets = [(0, 3), (1, 2), (2, 2), (3, 1), (4, 1)]
    total = 0
    for stage, count in buckets:
        for _ in range(count):
            flight = rnd.choice(flights[:4])
            passenger = rnd.choice(passengers)
            try:
                bag = api.post("/baggage", {
                    "passenger_id": passenger["id"],
                    "flight_id": flight["id"],
                })
            except ApiError as exc:
                print(f"  ! failed to create bag: {exc}")
                continue
            # Scan forward the right number of times.
            for _ in range(stage):
                try:
                    api.post(f"/baggage/{bag['id']}/scan", {})
                except ApiError:
                    break
            total += 1
    print(f"  ✓ {total} baggage tags created")


def seed_incident(api: Api) -> None:
    print("→ Posting one FLIGHT_DELAYED incident…")
    flights = api.get("/flights/upcoming") or []
    if not flights:
        print("  ! skipped — no upcoming flights")
        return
    flight = flights[0]
    try:
        api.post("/incidents", {
            "type": "FLIGHT_DELAYED",
            "flight_id": flight["id"],
            "reason": "Метеоусловия в аэропорту назначения, задержка 40 минут",
        })
        print(f"  ✓ FLIGHT_DELAYED for {flight['flight_number']}")
    except ApiError as exc:
        print(f"  ! incident post failed: {exc}")


def _translit(text: str) -> str:
    """Very small Russian→Latin transliteration so generated emails are ASCII."""
    table = {
        "а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "е": "e", "ё": "e",
        "ж": "zh", "з": "z", "и": "i", "й": "i", "к": "k", "л": "l", "м": "m",
        "н": "n", "о": "o", "п": "p", "р": "r", "с": "s", "т": "t", "у": "u",
        "ф": "f", "х": "h", "ц": "ts", "ч": "ch", "ш": "sh", "щ": "sch",
        "ъ": "", "ы": "y", "ь": "", "э": "e", "ю": "yu", "я": "ya",
    }
    out = []
    for ch in text:
        lower = ch.lower()
        mapped = table.get(lower, ch)
        if ch.isupper():
            mapped = mapped.capitalize()
        out.append(mapped)
    return "".join(out)


# ---------------------------------------------------------------------------
# Entrypoint
# ---------------------------------------------------------------------------

def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="Seed the ClearFly demo with realistic data.")
    parser.add_argument("--base-url", default="http://localhost:8080",
                        help="API gateway base URL (default: %(default)s)")
    parser.add_argument("--passengers", type=int, default=18,
                        help="Number of passengers to create (default: %(default)s)")
    parser.add_argument("--seed", type=int, default=42,
                        help="Random seed for reproducible runs (default: %(default)s)")
    parser.add_argument("--skip-baggage", action="store_true",
                        help="Do not create baggage tags")
    parser.add_argument("--skip-incident", action="store_true",
                        help="Do not post a demo incident")
    args = parser.parse_args(argv)

    rnd = random.Random(args.seed)
    api = Api(args.base_url)

    print(f"ClearFly demo seeder · {args.base_url}")
    try:
        wait_for_gateway(api)
    except ApiError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    passengers = seed_passengers(api, args.passengers, rnd)
    if not passengers:
        print("error: no passengers were created; aborting.", file=sys.stderr)
        return 1

    seed_bookings(api, passengers, rnd)
    if not args.skip_baggage:
        seed_baggage(api, passengers, rnd)
    if not args.skip_incident:
        seed_incident(api)

    print("\nAll done. Open http://localhost:3000 to see the demo.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
