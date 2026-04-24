# dryad-sh code guide

`dryad-sh` is the shell reference implementation of Dryad. The source is
organized for correctness first: shell makes it easy to mix parsing, reads,
writes, and execution in one place, so the code should keep those concerns
separate on purpose.

The generated `dyd-stem-run` may be a single shell script. That is a packaging
format, not the source architecture.

## Layers

Code should be grouped by layer. Dependencies point downward only.

- `lib`: generic shell primitives. These functions do not know what a Dryad
  root, stem, sprout, scope, or garden is.
- `model`: read, resolve, validate, and describe Dryad filesystem objects.
  Model functions may read the garden and may use ephemeral memo caches, but
  they must not mutate Dryad state.
- `ops`: perform state transitions. These functions create, delete, publish,
  chmod, symlink, execute, or otherwise mutate the garden or a temporary build
  workspace.
- `cmd`: parse CLI arguments, call model/ops functions, and format user output.
- `main`: command dispatch and process-level setup only.

Examples:

- `dryad_model_variant_descriptors "$root"` belongs in `model`.
- `dryad_model_heap_fingerprint_path "$garden" stems "$fingerprint"` belongs in
  `model`.
- `dryad_op_heap_publish_stem "$src" "$fingerprint"` belongs in `ops`.
- `dryad_op_root_build "$root" "$descriptor"` belongs in `ops`.
- `dryad_cmd_root_build "$@"` belongs in `cmd`.

## Naming

Use names that make the layer visible during review.

- `dryad_lib_*`: generic primitives.
- `dryad_model_*`: read/resolve/validate/describe Dryad state.
- `dryad_op_*`: mutate state or execute commands.
- `dryad_cmd_*`: parse CLI arguments and print user-facing output.
- `dryad_build_phase_*`: root build pipeline phases.

If a function name and its behavior disagree, fix the boundary instead of
explaining the exception.

## Model Rules

Model functions answer questions such as "what is this path?", "which variant
matches?", or "is this object valid?" They should be deterministic for the same
filesystem state.

Allowed in model code:

- Reading files and directories.
- Resolving symlinks.
- Validating sentinel files, descriptors, requirement names, and selectors.
- Printing machine-readable results.
- Writing transparent memo cache entries.

Not allowed in model code:

- Creating, deleting, moving, chmodding, or publishing Dryad state.
- Running build commands or sprout commands.
- Parsing CLI flags.
- Formatting rich user output beyond concise diagnostics.

Deleting the memo cache must never change behavior beyond speed.

## Ops Rules

Ops functions change the world. They should have explicit input paths and
explicit output paths, and should avoid hidden dependency on the caller's
current directory.

Allowed in ops code:

- Creating temporary workspaces.
- Materializing roots, stems, and sprouts.
- Publishing heap files, stems, sprouts, and derivations.
- Updating `dyd/sprouts`.
- Running package commands.

Required in ops code:

- Use atomic publication patterns for heap state.
- Validate existing heap objects enough to avoid silently trusting corrupt
  artifacts.
- Keep mutation local to the declared output path unless the operation name says
  otherwise.
- Clean up temporary files on success; leave useful diagnostics on failure when
  practical.

## Performance Rules

Readable organization is not the main performance risk. Repeated process
launches and repeated filesystem walks are.

Guidelines:

- Keep top-level script load cheap: function definitions, constants, and final
  dispatch only.
- Avoid `find`, `sort`, `awk`, `sed`, `dirname`, `basename`, `wc`, and `tr`
  inside hot inner loops unless the cost is intentional.
- Prefer one manifest-producing walk over rediscovering the same tree in
  several phases.
- Prefer shell parameter expansion for simple string work.
- Prefer a single well-scoped `awk` program for bulk manifest processing over
  many tiny subprocesses.
- Memoization is an optimization, not an ownership model or correctness
  mechanism.

## Review Checklist

When reviewing a change, ask:

- Is this function in the right layer?
- Does it mutate state despite being model/lib/cmd code?
- Does it parse CLI flags despite not being cmd code?
- Does it depend on ambient current directory or ambient environment?
- Does it repeatedly walk or sort the same tree?
- Could deleting the memo cache change behavior?
- Do fingerprint, publish, and execution phases agree on the same package
  contents?
