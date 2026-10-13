#!/usr/bin/env python3
"""Exercise the KMX front door with synthetic README pages, not archived prose."""
import importlib.util
import subprocess
import sys
import tempfile
from pathlib import Path

CHECKER = Path(__file__).with_name("check-readme-front-door.py")
spec = importlib.util.spec_from_file_location("front_door", CHECKER)
front_door = importlib.util.module_from_spec(spec)
spec.loader.exec_module(front_door)

INSTALL = """brew install kaimahi-agents/tap/kmx
# or
curl -fsSL https://raw.githubusercontent.com/kaimahi-agents/kaimahi/main/install.sh | sh
# or
go install github.com/kaimahi-agents/kaimahi/cmd/kmx@v0.5.1
"""
WARNING = """> [!NOTE]
> The table below describes the current development build, which is unreleased.
> The latest tagged release is v0.5.1 and still uses `kmx agent ...`.
"""
JOURNEY = WARNING + """| Stage | Current development build |
|---|---|
| **Create** | `kmx create NAME --workspace DIR` authors source. |
| **Prove** | `kmx eval --workspace DIR` stops later cases after the first non-pass and writes evidence only when all cases pass. |
| **Publish** | `kmx push --workspace DIR --evaluation RECEIPT REF` copies the suite. Plain `kmx push REF` remains explicitly unevaluated. Neither form deploys. |
| **Lift** | `kmx lift REF@sha256:DIGEST --to-context CTX --inference provider:NAME --evidence RECEIPT` deploys without building, repacking, or publishing. `--plan` is GET-only; legacy bundle lift is separate. |
"""
GOOD = """<img src="brand/ketu.svg" alt="Kaimahi ketu mark">
**Agent Builder CLI for Kubernetes.**
[![CI](https://github.com/kaimahi-agents/kaimahi/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/kaimahi-agents/kaimahi/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kaimahi-agents/kaimahi)](https://github.com/kaimahi-agents/kaimahi/releases)
[![License](https://img.shields.io/github/license/kaimahi-agents/kaimahi)](LICENSE)
## Create, Prove, Lift
Root Lift is not currently qualified against a real cluster.
""" + JOURNEY + """## Get Your First Answer
```bash
""" + INSTALL + """```
```bash
kmx version
kmx quickstart
```
## Workspace And Live Paths
Workspace and live paths differ.
## Runtime Boundaries
- **Orka** is the default runtime and owns workspace evaluation/evaluated-publication/digest-lift, live create, run, chat and the legacy bundle lifecycle.
- **kagent** support is exact-version: v0.10.2 live create; v0.10.3 Go workspace authoring/publication and legacy bundle lift/status/evaluation. Public suite evaluation, lift and status are not available.
- **agentsessions** is evaluation and replay integration for the legacy bundle path, not a KMX deployment target.
## Safety And Status
> [!IMPORTANT]
> **Kaimahi is experimental and under active development.** Interfaces may change.
> [!NOTE]
> kmx is the Agent Builder and lifecycle layer, not a runtime or generic
> governance control plane.
## Guides
[Legacy PortableAgent bundle lifecycle](docs/agent-lift.md)
## Development
[developer guide](docs/development.md)
[codebase overview](docs/codebase-overview.md)
[interface baseline](docs/interface-baseline.md)
[repository map](docs/repository-map.md)
"""

CASES = [("valid front door", GOOD, None)]
for label, literal in [
    ("ketu icon", '<img src="brand/ketu.svg" alt="Kaimahi ketu mark">\n'),
    ("product line", "**Agent Builder CLI for Kubernetes.**\n"),
    ("journey heading", "## Create, Prove, Lift\n"),
    ("first-answer heading", "## Get Your First Answer\n"),
    ("workspace/live heading", "## Workspace And Live Paths\n"),
    ("runtime boundaries heading", "## Runtime Boundaries\n"),
    ("safety/status heading", "## Safety And Status\n"),
    ("guides heading", "## Guides\n"),
    ("development heading", "## Development\n"),
    ("experimental notice", "> [!IMPORTANT]\n> **Kaimahi is experimental and under active development.** Interfaces may change.\n"),
    ("runtime ownership note", "> [!NOTE]\n> kmx is the Agent Builder and lifecycle layer, not a runtime or generic\n> governance control plane.\n"),
    ("CI badge", "actions/workflows/ci.yml/badge.svg?branch=main"),
    ("release badge", "img.shields.io/github/v/release/kaimahi-agents/kaimahi"),
    ("license badge", "img.shields.io/github/license/kaimahi-agents/kaimahi"),
]:
    CASES.append((f"missing {label}", GOOD.replace(literal, ""), f"{label} is missing"))

for label, literal in [
    ("Homebrew install", "brew install kaimahi-agents/tap/kmx\n"),
    ("release installer", "curl -fsSL https://raw.githubusercontent.com/kaimahi-agents/kaimahi/main/install.sh | sh\n"),
    ("pinned Go install", "go install github.com/kaimahi-agents/kaimahi/cmd/kmx@v0.5.1\n"),
    ("kmx version", "kmx version\n"),
    ("kmx quickstart", "kmx quickstart\n"),
]:
    CASES.append((f"missing {label}", GOOD.replace(literal, ""), f"{label} is missing"))

CASES += [
    ("command in purpose column only", GOOD.replace("| **Create** | `kmx create", "| Author | **Create** with `kmx create"), "Create row is missing"),
    ("suite rows out of order", GOOD.replace("| **Create** |", "| **TEMP** |").replace("| **Prove** |", "| **Create** |").replace("| **TEMP** |", "| **Prove** |"), "Create row is missing"),
    ("missing eval", GOOD.replace("`kmx eval --workspace DIR`", "`kmx run`"), "Prove row is missing"),
    ("partial evidence accepted", GOOD.replace("only when all cases pass", "after any case"), "Prove row must require a full pass"),
    ("later cases continue after failure", GOOD.replace("stops later cases after the first non-pass", "continues through every case"), "Prove row must preserve fail-fast execution"),
    ("push implies deployment", GOOD.replace("Neither form deploys", "Then it deploys"), "Publish row must say it does not deploy"),
    ("plain push treated as evaluated", GOOD.replace("Plain `kmx push REF` remains explicitly unevaluated. ", ""), "Publish row must keep plain push unevaluated"),
    ("lift builds", GOOD.replace("without building, repacking, or publishing", "after rebuilding and publishing"), "Lift row must stay deploy-only"),
    ("lift plan writes", GOOD.replace("`--plan` is GET-only", "`--plan` applies resources"), "Lift row must preserve the GET-only plan boundary"),
    ("lift qualification claimed", GOOD.replace("not currently qualified against a real cluster", "qualified against a real cluster"), "root Lift must not claim real-cluster qualification"),
    ("late release warning", GOOD.replace(WARNING, "").replace("| **Lift** | `kmx lift REF@sha256:DIGEST --to-context CTX --inference provider:NAME --evidence RECEIPT` deploys without building, repacking, or publishing. `--plan` is GET-only; legacy bundle lift is separate. |", "| **Lift** | `kmx lift REF@sha256:DIGEST --to-context CTX --inference provider:NAME --evidence RECEIPT` deploys without building, repacking, or publishing. `--plan` is GET-only; legacy bundle lift is separate. |\n" + WARNING.rstrip()), "release/main warning must precede"),
    ("stable table label", GOOD.replace("| Stage | Current development build |", "| Stage | Latest stable release |"), "exact two-column header"),
    ("malformed table separator", GOOD.replace("|---|---|", "|---|"), "separator is missing or malformed"),
    ("extra suite row", GOOD.replace("| **Lift** |", "| **Inspect** | Future status. |\n| **Lift** |"), "exactly four rows"),
    ("extra table column", GOOD.replace("| **Create** | `kmx create NAME --workspace DIR` authors source. |", "| **Create** | `kmx create NAME --workspace DIR` authors source. | extra |"), "exactly two columns"),
    ("generic legacy guide", GOOD.replace("Legacy PortableAgent bundle lifecycle", "Lift operation guide"), "agent-lift links must be labeled"),
    ("Orka suite support broadened", GOOD.replace("owns workspace evaluation/evaluated-publication/digest-lift", "owns every suite operation"), "Orka default is missing"),
    ("wrong kagent version", GOOD.replace("v0.10.3 Go", "v1.0 Python"), "kagent version boundary is missing"),
    ("kagent suite lift claimed", GOOD.replace("Public suite evaluation, lift and status are not available.", "Public suite lift is available."), "kagent version boundary is missing"),
    ("agentsessions deployment claimed", GOOD.replace("not a KMX deployment target", "a KMX deployment target"), "agentsessions boundary is missing"),
    ("wrong tap", GOOD.replace("kaimahi-agents/tap/kmx", "other/tap/kmx"), "Homebrew install is missing"),
    ("moving Go source", GOOD.replace("cmd/kmx@v0.5.1", "cmd/kmx@main"), "pinned Go install is missing"),
    ("old Go release", GOOD.replace("cmd/kmx@v0.5.1", "cmd/kmx@v0.5.0"), "pinned Go install is missing"),
    ("wrong installer", GOOD.replace("kaimahi/main/install.sh", "other/main/install.sh"), "release installer is missing"),
    ("install routes out of order", GOOD.replace("brew install kaimahi-agents/tap/kmx\n# or\ncurl", "curl").replace("install.sh | sh\n# or\ngo install", "install.sh | sh\n# or\nbrew install kaimahi-agents/tap/kmx\n# or\ngo install"), "release installer is missing"),
    ("first answer out of order", GOOD.replace("kmx version\nkmx quickstart", "kmx quickstart\nkmx version"), "kmx quickstart is missing"),
    ("missing developer guide", GOOD.replace("[developer guide](docs/development.md)\n", ""), "developer guide link is missing"),
    ("missing codebase overview", GOOD.replace("[codebase overview](docs/codebase-overview.md)\n", ""), "codebase overview link is missing"),
    ("missing interface baseline", GOOD.replace("[interface baseline](docs/interface-baseline.md)\n", ""), "interface baseline link is missing"),
    ("missing repository map", GOOD.replace("[repository map](docs/repository-map.md)\n", ""), "repository map link is missing"),
    ("empty document", "", "ketu icon is missing"),
]

failed = 0
for name, text, expected in CASES:
    got = front_door.check(text)
    ok = (got is None) if expected is None else (got is not None and expected in got)
    print(("ok  " if ok else "FAIL") + f" [{name}] -> {got or 'valid'}")
    failed += not ok

with tempfile.TemporaryDirectory() as tmp:
    for name, text, want in [("valid", GOOD, 0), ("invalid", "# Empty README\n", 1)]:
        readme = Path(tmp) / "README.md"
        readme.write_text(text)
        got = subprocess.run([sys.executable, str(CHECKER), str(readme)], capture_output=True, text=True)
        ok = got.returncode == want
        print(("ok  " if ok else "FAIL") + f" [as a script: {name}] -> exit {got.returncode}, want {want}")
        failed += not ok

print(f"check-readme-front-door self-test: {len(CASES) + 2} case(s), {failed} failure(s)")
sys.exit(1 if failed else 0)
