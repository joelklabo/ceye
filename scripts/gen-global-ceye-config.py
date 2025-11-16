#!/usr/bin/env python3
"""Build a global `ceye.yaml` by querying GitHub and Azure DevOps for your orgs."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

DEFAULT_GITHUB_ORG = "joelklabo"
DEFAULT_AZURE_ORG = "joelklabo"
DEFAULT_AZURE_PROJECT = "Big Timer"
DEFAULT_OUTPUT = Path.home() / ".config" / "ceye" / "ceye.yaml"


def run_command(cmd: list[str]) -> str:
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as exc:
        stderr = exc.stderr.strip()
        hint = f"\n{stderr}" if stderr else ""
        raise SystemExit(f"{cmd!r} failed: {exc.returncode}{hint}")
    return result.stdout.strip()


def gather_github_repos(org: str) -> list[str]:
    cmd = [
        "gh",
        "repo",
        "list",
        org,
        "--limit",
        "1000",
        "--json",
        "nameWithOwner",
    ]
    output = run_command(cmd)
    try:
        records = json.loads(output)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"failed to decode GH output: {exc}")
    repos = []
    for record in records:
        name = record.get("nameWithOwner")
        if not name or "/" not in name:
            continue
        owner, repo = name.split("/", 1)
        if owner != org:
            continue
        repos.append(repo)
    if not repos:
        raise SystemExit(f"no repositories found for GitHub org {org}")
    return sorted(dict.fromkeys(repos))


def gather_azure_pipelines(org: str, project: str) -> list[int]:
    org_url = org if org.startswith("http://") or org.startswith("https://") else f"https://dev.azure.com/{org}"
    cmd = [
        "az",
        "pipelines",
        "list",
        "--org",
        org_url,
        "--project",
        project,
        "--query",
        "[].id",
        "-o",
        "tsv",
    ]
    output = run_command(cmd)
    pipelines = []
    for line in output.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            pipelines.append(int(line))
        except ValueError:
            raise SystemExit(f"unexpected pipeline id: {line}")
    if not pipelines:
        raise SystemExit(f"no pipelines found for {project} in {org_url}")
    return sorted(dict.fromkeys(pipelines))


def build_config(repo_list: list[str], github_org: str, azure_org: str, azure_project: str, pipelines: list[int]) -> str:
    lines = [
        "# Auto-generated global ceye configuration",
        "# Run scripts/gen-global-ceye-config.py (with --help) whenever your orgs change.",
        "providers:",
        "  - type: github",
        "    repos:",
    ]
    for repo in repo_list:
        lines.append(f"      - owner: {github_org}")
        lines.append(f"        repo: {repo}")
    lines.append("  - type: azure")
    lines.append(f"    org: {azure_org}")
    lines.append(f"    project: {azure_project}")
    lines.append("    pipelines:")
    for pipeline_id in pipelines:
        lines.append(f"      - {pipeline_id}")
    lines.append("")
    return "\n".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate a global ceye.yaml from GH/Azure metadata")
    parser.add_argument("--github-org", default=DEFAULT_GITHUB_ORG, help="GitHub organization to monitor")
    parser.add_argument("--azure-org", default=DEFAULT_AZURE_ORG, help="Azure DevOps organization name or URL")
    parser.add_argument("--azure-project", default=DEFAULT_AZURE_PROJECT, help="Azure DevOps project to inspect")
    parser.add_argument("--output", default=str(DEFAULT_OUTPUT), help="Output path for the generated config")
    args = parser.parse_args()

    repo_list = gather_github_repos(args.github_org)
    pipelines = gather_azure_pipelines(args.azure_org, args.azure_project)
    config_text = build_config(repo_list, args.github_org, args.azure_org, args.azure_project, pipelines)

    output_path = Path(args.output).expanduser()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(config_text, encoding="utf-8")
    print(f"wrote global config to {output_path}")


if __name__ == "__main__":
    main()
