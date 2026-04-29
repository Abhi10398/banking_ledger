#!/usr/bin/env python3
"""
Concurrency test: fire N concurrent transfers between two accounts and verify
that the final balances are consistent — no negative balance, no money created
or destroyed.

Usage:
    python3 scripts/concurrency_test.py <account_a_id> <account_b_id> [n] [amount]

    account_a_id  UUID of the source account
    account_b_id  UUID of the destination account
    n             number of concurrent transfers (default: 50)
    amount        amount per transfer in paise (default: 100)

Example:
    python3 scripts/concurrency_test.py \
        aaaaaaaa-0000-0000-0000-000000000001 \
        bbbbbbbb-0000-0000-0000-000000000002
"""
import concurrent.futures
import json
import sys
import urllib.error
import urllib.request

BASE = "http://localhost:8080/api"


def get_account(account_id: str) -> dict:
    try:
        with urllib.request.urlopen(f"{BASE}/accounts/{account_id}", timeout=10) as r:
            return json.loads(r.read())
    except urllib.error.HTTPError as e:
        if e.code == 404:
            print(f"Error: account {account_id} not found. Run 'make seed' and use the printed UUIDs.", file=sys.stderr)
        else:
            print(f"Error: GET /accounts/{account_id} returned HTTP {e.code}.", file=sys.stderr)
        sys.exit(1)
    except urllib.error.URLError as e:
        print(f"Error: could not reach {BASE} — is the app running? ({e.reason})", file=sys.stderr)
        sys.exit(1)


def transfer(from_id: str, to_id: str, amount: int) -> tuple[int, dict]:
    body = json.dumps({"from_account_id": from_id, "to_account_id": to_id, "amount": amount}).encode()
    req = urllib.request.Request(f"{BASE}/transfers", data=body, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            return r.status, json.loads(r.read())
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read())


def test_concurrent_transfers(account_a: str, account_b: str, n: int = 50, amount: int = 100):
    a_before = get_account(account_a)
    b_before = get_account(account_b)
    initial_total = a_before["balance"] + b_before["balance"]

    print(f"Initial state:")
    print(f"  A ({account_a[:8]}…)  balance={a_before['balance']}")
    print(f"  B ({account_b[:8]}…)  balance={b_before['balance']}")
    print(f"  Total={initial_total}")
    print(f"\nFiring {n} concurrent transfers of {amount} each…")

    with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
        futures = [executor.submit(transfer, account_a, account_b, amount) for _ in range(n)]
        results = [f.result() for f in concurrent.futures.as_completed(futures)]

    success = sum(1 for status, _ in results if status == 200)
    failed  = sum(1 for status, _ in results if status != 200)

    a_after = get_account(account_a)
    b_after = get_account(account_b)
    final_total = a_after["balance"] + b_after["balance"]

    print(f"\nResults: {success} succeeded, {failed} failed/rejected")
    print(f"\nFinal state:")
    print(f"  A ({account_a[:8]}…)  balance={a_after['balance']}")
    print(f"  B ({account_b[:8]}…)  balance={b_after['balance']}")
    print(f"  Total={final_total}")

    # Invariant checks
    a_ok    = a_after["balance"] >= 0
    b_ok    = b_after["balance"] >= 0
    total_ok = final_total == initial_total

    print(f"\nInvariant checks:")
    print(f"  A balance non-negative : {'PASS' if a_ok    else 'FAIL'}")
    print(f"  B balance non-negative : {'PASS' if b_ok    else 'FAIL'}")
    print(f"  Total conserved        : {'PASS' if total_ok else 'FAIL'}")

    if not (a_ok and b_ok and total_ok):
        print("\nINVARIANT VIOLATED", file=sys.stderr)
        sys.exit(1)

    print("\n All invariants hold.")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(__doc__)
        sys.exit(1)

    account_a = sys.argv[1]
    account_b = sys.argv[2]
    n         = int(sys.argv[3]) if len(sys.argv) > 3 else 50
    amount    = int(sys.argv[4]) if len(sys.argv) > 4 else 100

    test_concurrent_transfers(account_a, account_b, n=n, amount=amount)
