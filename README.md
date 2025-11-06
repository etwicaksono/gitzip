# 🧩 GitZip

A simple Go CLI tool (built with [spf13/cobra](https://github.com/spf13/cobra)) to **collect and zip all changed files between two Git commits** — similar to a PowerShell automation script, but cross-platform and portable.

---

## 🚀 Features

* ✅ Retrieve changed files between two Git commits
* 🚫 Excludes deleted files automatically (`--diff-filter=d`)
* 🗂️ Preserves folder structure inside the ZIP
* 💻 Works on Linux, macOS, and Windows
* ⚙️ Lightweight — no external dependencies beyond Git CLI

---

## 🧱 Installation

### Clone & Build

```bash
git clone https://github.com/etwicaksono/gitzip.git
cd gitzip
go mod tidy
go build -o gitzip .
```

Or install directly (if you publish it later):

```bash
go install github.com/etwicaksono/gitzip@latest
```

---

## ⚙️ Command Syntax

```bash
gitzip <commit1> <commit2> <zipfile>
```

| Argument    | Description                                                  |
| ----------- | ------------------------------------------------------------ |
| `<commit1>` | Starting commit hash or reference (e.g., `HEAD~1`, `abc123`) |
| `<commit2>` | Ending commit hash or reference (e.g., `HEAD`, `main`)       |
| `<zipfile>` | Output ZIP filename (e.g., `changes.zip`)                    |

---

## 🧩 Examples

### 1️⃣ Compare last commit with the current one

```bash
gitzip HEAD~1 HEAD changes.zip
```

Zips all modified and newly added files between the previous commit and the current `HEAD`.

---

### 2️⃣ Compare two specific commits

```bash
gitzip 123abc 789def feature_update.zip
```

Creates `feature_update.zip` containing all changes between commits `123abc` and `789def`.

---

### 3️⃣ Compare branches

```bash
gitzip main develop diff.zip
```

Packages files that differ between `main` and `develop` branches.

---

### 4️⃣ Compare by date references

Git allows you to use date-based references like:

```bash
gitzip main@{yesterday} main@{now} daily_update.zip
```

This zips all changes committed from **yesterday to today** on the `main` branch.

---

## 📚 Usage Guide

Here’s how you can effectively use `gitzip` in real workflows:

### 🧮 A. Before deployment

Want to send only changed files to QA or DevOps?

```bash
gitzip HEAD~5 HEAD release_patch.zip
```

→ Zips the last 5 commits’ worth of changes.

### 🥪 B. Before creating a Pull Request

Generate a patch file bundle:

```bash
gitzip origin/main HEAD pr-changes.zip
```

→ Helps reviewers see exactly what’s changed since you branched off.

### 🛋 C. For code backup or transfer

You can export changed source files to share with another developer without the entire repo:

```bash
gitzip HEAD~2 HEAD my_changes.zip
```

### 🧮 D. Inspect what changed

If you only want to list files without zipping:

```bash
git diff --name-only HEAD~1 HEAD
```

`gitzip` simply automates this step and zips the result for portability.

---

## 💾 Requirements

* [Go 1.21+](https://go.dev/dl/)
* [Git](https://git-scm.com/downloads) must be installed and available in your PATH

---

## 📁 Project Structure

```
.
├── main.go          # CLI implementation
├── go.mod
├── go.sum
└── README.md
```

---

## 🧑‍💻 Development

To add more subcommands or options:

```bash
go install github.com/spf13/cobra-cli@latest
cobra-cli add <subcommand-name>
```

---

## 🧮 Optional: Makefile Usage

You can add a simple `Makefile` to streamline commands:

```makefile
build:
	go build -o gitzip .

run:
	go run . HEAD~1 HEAD changes.zip

clean:
	rm -f gitzip *.zip
```

Run:

```bash
make build
make run
```

---

### ❤️ Credits

Built with [Go](https://go.dev/) and [Cobra](https://github.com/spf13/cobra).
Inspired by a PowerShell automation script for Git diff packaging.
