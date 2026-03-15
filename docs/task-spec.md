# Task Specification Standard

This document defines how to write a backlog of agile implementation tasks for a software project. Its goal is to produce tasks that are well-ordered, self-contained, testable, and that result in clean, well-structured code — without over-prescribing how that code is written.

---

## Structure of a backlog

### Milestones

Group tasks into milestones. Each milestone represents a state of the project that is demonstrably more complete than the one before it. A milestone should have a one-line description of what works when all its tasks are done. Milestones provide orientation — a developer should be able to read the milestone list and understand the shape of the project at a glance.

A milestone contains between 1 and 5 tasks. If a milestone would contain more, consider splitting it.

### MVP scope

Define the MVP scope before writing any tasks. State explicitly what is in scope and what is out of scope (post-MVP). This prevents scope creep and gives tasks a clear stop condition.

### Story points

Assign story points to each task. Use a simple 1–3 scale:
- **1 point:** straightforward wiring, thin task, little ambiguity
- **2 points:** moderate scope, some design decisions
- **3 points:** significant scope, multiple moving parts

Tasks should rarely exceed 3 points. If a task feels like 4 or more, split it. Story points must include the cost of any refactoring noted in the task — not as a separate estimate.

---

## Structure of a task

Every task must include the following sections, in order:

### Story points

A single number.

### Description

One or more paragraphs describing what this task accomplishes and why. Written in terms of behavior and outcomes, not implementation. A developer should be able to read this section and understand what they are building and what problem it solves, without needing to read any other document.

**Self-contained rule:** The description must not say "see the spec" or "see the README." All context needed to implement the task must be inline. If the task depends on a schema, a file format, or a specific behavior defined elsewhere, reproduce the relevant details here.

### Refactoring note *(optional)*

When a task will likely require looking back at previously written code and improving its structure, say so explicitly. Keep it short and specific: name the earlier task, describe what the structural problem will probably be, and state clearly that the refactoring is expected, costed into the estimate, and part of the Definition of Done.

Do not add a refactoring note speculatively. Only add one when the structural tension is predictable — for example, when two commands share significant logic, or when adding a cross-cutting concern (like preflight) will require separating validation from execution in an earlier command.

**Inline vs. separate ticket:** Keep the refactoring note inside the implementation ticket unless the refactoring would touch code from two or more *previous* tickets. In that case, a dedicated refactoring ticket may be warranted. Separate refactoring tickets must state which earlier tasks they touch, what the structural goal is, and that all previously passing tests must continue to pass — no new tests are required.

### Required tests

Specify the tests that must pass for this task to be considered done. Tests must be concrete: specific inputs, specific expected outputs, specific assertions. Vague requirements like "write tests for the happy path" are not acceptable.

Two categories (see the Testing section below for details):

- **Txtar script tests:** for anything involving the compiled binary
- **Unit tests:** for business logic that does not involve running the binary

### Definition of Done

A short bulleted list of verifiable statements. Each item must be objectively true or false — no subjective criteria. Typically includes: specific test commands pass, specific behaviors are observable, `go test ./...` passes.

---

## Ordering principles

### Foundation before features

The project must build and parse its config file before any command is implemented. These are always the first one or two tasks.

### Shared infrastructure before dependent commands

If two or more commands share a significant piece of logic (host resolution, secret collection, a file writer), implement that shared logic as its own task before implementing either command. This prevents the logic from being written twice and then merged awkwardly.

### Simple commands before orchestrators

Commands that do one thing come before commands that coordinate other commands. `stc setup`, which calls `inventory generate`, `ansible bootstrap-host`, etc., comes last.

### Core behavior before cross-cutting concerns

Implement each command's core behavior (the happy path and error cases) before adding cross-cutting concerns like `--preflight`, `--dry-run`, or diagnostic output. This keeps early tasks focused and avoids forcing the implementor to design two behaviors simultaneously. Cross-cutting concerns get their own milestone after all commands exist.

### Anticipate calling conventions

When a later task will need to call into the logic of an earlier task (e.g., `stc setup` calling `inventory generate` logic), the earlier task cannot be written in isolation. State the global architectural rule in a "Notes for all tasks" section so every implementor knows it from the start. Do not bury this expectation only in the later task.

---

## Global notes

At the top of the backlog, before any milestones, include a "Notes for all tasks" section. This is where project-wide rules and conventions go — things every task must follow that would be redundant to repeat in every ticket.

Candidates for global notes:
- **Architectural principles** (e.g., business logic in `internal/`, CLI wiring in `cmd/`, callable without going through the CLI layer)
- **Mandated libraries** (e.g., a specific YAML library, a specific test framework)
- **Mandated patterns** (e.g., use `text/template` for all text generation, use txtar for command-level tests)
- **Test framework setup** (describe the test format and available verbs once, so individual tasks can reference it without re-explaining)

Keep global notes to genuine cross-cutting rules. Per-task details belong in the task.

---

## Testing requirements

### Command-level tests: txtar script tests

Any behavior observable through the compiled binary must be tested with txtar script tests. The test framework compiles the binary, runs it inside a temporary directory populated with the files defined in the test, and asserts on stdout, stderr, exit code, and generated file contents.

Txtar tests should cover:
- The happy path (command succeeds, expected output and files produced)
- The most important error cases (missing config, unknown flag values, missing required fields)
- Exit codes

Txtar test specs in task descriptions should be written out in full — not described in prose. The developer should be able to copy the test directly into a file.

### Unit tests: business logic

Logic that does not involve running the binary (parsing, rendering, resolving, building argument lists) must be covered by standard unit tests. Unit test specs must name the test function, specify the exact input, and specify the exact expected output or behavior. Vague descriptions like "test with a valid input" are not acceptable.

Unit tests are particularly important for:
- Parsers and serializers
- Template renderers
- Resolvers (e.g., mapping flag values to config records)
- Argument builders
- Any pure function with non-trivial logic

### What not to test

Do not require automated tests for behaviors that depend on external systems (running Ansible, running Terraform, writing to a real server). State explicitly that these are verified manually. A note in the Definition of Done like "manual verification expected for the full playbook run" is sufficient.

---

## Implementation vs. prescription

### Describe behavior, not implementation

Tasks specify what the code must do and how it must be tested. They do not specify how to implement it. Do not name structs, methods, or function signatures unless there is a compelling reason. Do not describe the internal architecture of a function. Let the implementor make those decisions.

**Wrong:** "Implement a `Diagnostics` struct with an `Append` method and a `HasErrors` method that returns true if any entry has status `not found`."

**Right:** "When any required config value is missing, the command must print a two-column table showing which values are missing. The table must auto-size its column widths to the widest entry."

### Exceptions

Sometimes a specific library or pattern must be mandated project-wide for good reason (consistency, maintainability, avoiding a known bad option). These mandates belong in the global notes, not in individual tasks. State the mandate and the reason once. Examples:
- "Use `github.com/goccy/go-yaml` — the previously standard library was archived."
- "Use `text/template` for all generated text output — this keeps the output format readable and modifiable without touching Go code."
- "Use txtar script tests for command-level behavior — this compiles the real binary and tests it end-to-end."

---

## Anti-patterns to avoid

**Forward references in early tasks.** Do not write "this will be called by task 7 later, so structure it accordingly." The implementor of task 3 should not need to know about task 7. Instead, state the architectural principle globally (business logic must be callable without the CLI layer), and let each implementor follow it naturally.

**Vague testing requirements.** "Write appropriate tests" is not a testing requirement. Specify inputs, expected outputs, and test names.

**Implementation details in the description.** If the task says "use a `map[string]string` to store the diagnostic results," it is over-prescribing. Say what the output must look like, not how to store it internally.

**Tasks that are too large.** If a task requires more than one PR's worth of review, split it. A task should be completable and reviewable in isolation.

**Tasks that cannot be tested in isolation.** If a task's correctness cannot be verified until three tasks later, the task boundaries are wrong. Reorder or re-scope so that each task has at least one test that can be run and passes on its own.

**Refactoring notes that are separate tickets for small, localized cleanup.** If the refactoring only touches code written in the current task, it belongs inside the current task. Separate tickets are only warranted when the refactoring spans multiple prior tasks.

**Milestones with too many tasks.** More than five tasks in one milestone makes the milestone feel like a feature dump. Split it.

**Referencing external documents.** Every task must be self-contained. If a task says "refer to the schema in `cli-spec.md`," the developer has to context-switch. Reproduce the relevant details inline.
