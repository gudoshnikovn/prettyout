"""Utility to format Ruff JSON output into a human-readable summary."""

import json
import sys
from collections import defaultdict
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from collections.abc import Sequence


def format_ruff_json() -> None:
    """Parse Ruff JSON from stdin and print a colorized, grouped summary."""
    try:
        if sys.stdin.isatty():
            print("Usage: ruff check . --output-format=json | python ruff_pretty.py")
            return
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        print("Error: Invalid JSON.")
        return

    rules: dict[str, dict[str, str | set[tuple[str, str]]]] = defaultdict(lambda: {"message": "", "items": set()})

    if not isinstance(data, list):
        print("Error: Expected a list of issues.")
        return

    actual_data: Sequence[dict[str, Any]] = data

    for entry in actual_data:
        code = str(entry.get("code", "UNKNOWN"))
        full_path = str(entry.get("filename", ""))
        filename = full_path.split("/")[-1]
        if "backend/" in full_path:
            filename = full_path.split("backend/")[-1]

        location = entry.get("location", {})
        end_location = entry.get("end_location", {})
        start_row = location.get("row")
        end_row = end_location.get("row")

        line_info = f"line {start_row}" if start_row == end_row else f"lines {start_row}-{end_row}"

        rules[code]["message"] = str(entry.get("message", ""))
        items = rules[code]["items"]
        if isinstance(items, set):
            items.add((filename, line_info))

    for code, info in sorted(rules.items()):
        # Yellow bold code
        print(f"\033[1;33m{code}\033[0m - \033[1m{info['message']}\033[0m")
        print("Affected files:")

        items = info["items"]
        if isinstance(items, set):
            # Sort the unique entries
            for file, line in sorted(items):
                print(f"  - {file} — {line}")
        print("-" * 40)


if __name__ == "__main__":
    from typing import Any

    format_ruff_json()
