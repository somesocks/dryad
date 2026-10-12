# Run fixtures

Shared shell fixtures for stdin and sprout concurrency e2e tests. Each test creates
and builds a disposable garden through the CLI; workers coordinate through marker
files and a locked active-process counter.
