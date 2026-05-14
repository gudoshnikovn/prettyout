"""Utility to format Basedpyright JSON output into a human-readable summary."""

import json
import sys
from collections import defaultdict
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from collections.abc import Sequence


def format_basedpyright() -> None:
    """Parse Basedpyright JSON from stdin and print a colorized, grouped summary."""
    try:
        if sys.stdin.isatty():
            print("Usage: basedpyright --outputjson . | python basedpyright_pretty.py")
            return
        content = sys.stdin.read()
        if not content:
            return
        data = json.loads(content)
    except json.JSONDecodeError:
        print("Error: Invalid JSON.")
        return

    diagnostics: Sequence[dict[str, Any]] = data.get("generalDiagnostics", [])
    rules: dict[str, dict[str, str | set[tuple[str, str]]]] = defaultdict(lambda: {"message": "", "items": set()})

    for entry in diagnostics:
        rule_code = str(entry.get("rule", "StaticAnalysis"))
        file_path = str(entry.get("file", "unknown"))
        filepath = file_path.split("/")[-1]
        if "backend/" in file_path:
            filepath = file_path.split("backend/")[-1]

        diag_range = entry.get("range", {})
        start_pos = diag_range.get("start", {})
        end_pos = diag_range.get("end", {})
        start = (start_pos.get("line", 0) if isinstance(start_pos, dict) else 0) + 1
        end = (end_pos.get("line", 0) if isinstance(end_pos, dict) else 0) + 1
        line_info = f"line {start}" if start == end else f"lines {start}-{end}"

        if not rules[rule_code]["message"]:
            rules[rule_code]["message"] = str(entry.get("message", ""))

        items = rules[rule_code]["items"]
        if isinstance(items, set):
            items.add((filepath, line_info))

    for code, info in sorted(rules.items()):
        # Blue bold code
        message = str(info["message"])
        print(f"\033[1;34m{code}\033[0m - \033[1m{message[:100]}...\033[0m")
        print("Affected files:")
        items = info["items"]
        if isinstance(items, set):
            for file, line in sorted(items):
                print(f"  - {file} — {line}")
        print("-" * 50)


if __name__ == "__main__":
    format_basedpyright()
