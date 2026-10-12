# Sprout run fixtures

Shared shell fixtures for sprout execution stdin and concurrency e2e tests. Each
test creates and builds a disposable garden through the CLI; workers coordinate
through marker files and a locked active-process counter.
