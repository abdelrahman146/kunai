# Kunai CLI
Your Development Swiss Knife

<!-- right-floated logo -->
<p align="center">
<img src="./kunai-logo.png" alt="Kunai Logo"
     width="250"
     style="margin: 0 0 1em 1em;" />
</p>
<hr />

# Description

🗡️ Kunai CLI is your development swiss knife — a compact command-line toolchain designed to streamline repetitive workflows and supercharge productivity for polyglot teams.

## Features

- Rapid project scaffolding with language-aware templates
- Intelligent task runners that surface context-aware shortcuts
- Git helpers for linting, formatting, and changelog hygiene
- Extensible plugin system with simple YAML-based configuration
- Built-in AI hooks for summarizing diffs and drafting release notes

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/dhruv/kunai/main/install.sh | bash
```

- Alternatively, clone the repository and run `make install`.
- Kunai supports Linux and macOS; Windows support is tracked in the roadmap.

## Usage

```bash
kunai init my-app        # Scaffold a new application
kunai run lint           # Execute the default lint pipeline
kunai git smart-commit   # Generate a commit message draft
kunai doctor             # Inspect environment issues and suggestions
```

- Run `kunai help` to see all commands and flags.
- Configuration lives in `.kunai.yml`; use `kunai config edit` to generate a starter file.
- Commands can be chained, e.g. `kunai run lint format test`.

## Roadmap

- [ ] AI auto generate commits
- [ ] AI auto generate PR descriptions
- [ ] Windows port
- [ ] Plugin marketplace

## Contributing

1. Fork the repository and create a feature branch.
2. Run `kunai run lint test` before opening a PR.
3. Describe your change succinctly and include screenshots or terminal output when relevant.

## Support

- File issues or feature requests in the GitHub tracker.
- Join the discussion on Discord: `https://discord.gg/kunai-cli`.
- Follow release notes in `CHANGELOG.md`.
