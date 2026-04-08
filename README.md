# dogear

A CLI-based tool for you to remember where you stopped reading! 

## What It Does

You add titles from the shell, drop checkpoints as you go, and ask `where` when you pick something up again. List active and finished work, skim what you touched lately, browse history, tag sources, search with fuzzy matching, and export everything as JSON if you want a backup. Everything stays in a single SQLite file on your machine. Set `DOGEAR_DB` or pass `--db` if you want that file somewhere else, for example next to a sync folder.

## Installation

You need [Go](https://go.dev/dl/) 1.22 or newer.

```bash
git clone https://github.com/example/dogear.git
cd dogear
go build -o dogear ./cmd/dogear
```

Move the binary onto your `PATH` if you like, for example into `~/bin` or `/usr/local/bin`.

Run tests from the repo root:

```bash
go test ./...
```

## Usage

Add a book with a first checkpoint:

```bash
dogear add "The Hobbit" --page 47 --total-pages 310 --format paperback --note "arrived at Rivendell"
dogear update "The Hobbit" --page 61
dogear update "Distributed Systems Notes" --section "3.4" --note "start of leader election"
dogear update "Paper on BFT" --loc 1832
dogear where "The Hobbit"
dogear list
dogear list --active
dogear list --finished
dogear list --tag research
dogear lately
dogear history "The Hobbit"
dogear finish "The Hobbit"
dogear remove "The Hobbit" --yes
dogear search hobbit
dogear stats
```

Other commands: `tag`, `untag`, `stale`, `export`, `import`, `doctor`.

Import overwrites the database: use `dogear import backup.json --replace`.

Sample `where` output:

```text
The Hobbit
page 61 / 310 • 20%
updated today
note: arrived at Rivendell
```

Sample `list` output:

```text
The Hobbit                      page 61 / 310 • 20%    active    updated today
Distributed Systems Notes       section 3.4            active    updated 2d ago
Paper on BFT                    loc 1832               active    updated 5h ago
```

## Reference

**Commands:** `add`, `update`, `where`, `list`, `lately`, `history`, `finish`, `remove`, `search`, `stats`, `tag`, `untag`, `stale`, `export`, `import`, `doctor`.

**Global:** `--db` path to the SQLite file. Overrides `DOGEAR_DB` when set.

**add:** `--author`, `--format`, `--page`, `--chapter`, `--section`, `--loc`, `--percent`, `--note`, `--total-pages`, `--total-chapters`, `--tag` (repeat or comma-separated), `--allow-duplicate` when the same title and format already exists as active.

**update:** `--page`, `--chapter`, `--section`, `--loc`, `--percent`, `--note`. You may combine fields when it makes sense; the tool picks a primary position type for the checkpoint.

**list:** `--active`, `--finished`, `--tag`.

**lately:** `--n` limits how many rows appear (default 20).

**remove:** `--yes` / `-y` or `--force` skips the confirmation prompt.

**export:** `--out` / `-o` writes JSON to a file instead of stdout.

**import:** `--replace` is required so imports never happen by accident.

**stale:** `--days` defaults to 30; lists active titles older than that threshold.

Titles are resolved by fuzzy match when the exact string is not passed. If several titles tie, `dogear` prints a short disambiguation list.

**Storage:** By default the database is created at a path under your OS config directory, in a `dogear` folder, file name `dogear.db`. On many Unix systems this follows `os.UserConfigDir()` (for example `~/Library/Application Support` on macOS). Override with `DOGEAR_DB` or `--db`.

**Data model:** `titles` hold `title`, optional `author_or_source`, `format`, optional `total_pages` / `total_chapters`, `status` (`active` or `finished`), timestamps, and optional `finished_at`. Tags live in `title_tags`. Each `update` appends a `checkpoints` row (`position_type` plus optional `page`, `chapter`, `section`, `loc`, `percent`, `note`, `created_at`). The latest checkpoint drives `where` and list summaries; `history` shows the full chain.

## Roadmap

Ideas that could arrive later, only if they stay small: optional `DOGEAR_CONFIG` for default flags, read-only `dogear show` alias for `where`, and perhaps a `--json` mode on `list` for scripting. None of this is required for daily use.

## Limitations

There is no built-in sync layer. Point `DOGEAR_DB` at a folder your tools already sync if you want the same file on several machines. Fuzzy matching is intentionally simple; very short queries can match more than you meant. Import replaces all rows when you pass `--replace`; keep backups.

## Contributing

Pull requests and issues are very welcome! :-)

Please run `go test ./...` before sending a patch.

## License

MIT
