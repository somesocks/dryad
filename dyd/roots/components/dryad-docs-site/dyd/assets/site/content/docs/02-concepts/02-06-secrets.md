---
title: "02.06 - Secrets"
description: "Handling sensitive values in a package."
type: docs
layout: single
---

# Secrets

If a stem needs access to credentials, keys, or other temporary secrets, they will be mounted in the `dyd/secrets/` directory at install or run-time.  The secrets a stem _must be available_ during the stem build process, and will be included (indirectly) as part of the stem fingerprint.  If a stem should have access to secrets at run-time, a **secrets fingerprint** will be generated and stored in the stem at `dyd/secrets-fingerprint`.

Secrets are _not_ stored as part of a stem, and _not_ packed into a stem when it is bundled.  Think of a secret as an "addon" mounted into a stem at runtime.

Secrets _are_ stored in the heap, and one stem can easily access secrets used by other stems just by exploring outside of its directory.  They are _not_ meant to be a secure-at-rest storage mechanism for secrets, just a secrets-injection mechanism during the build process, which is useful for things like keys for package signing, or downloading assets from sources that require authentication.

If you need high-security secrets for an application to access at runtime, you should consider using another mechanism, such as environment variables, or pre-filling the execution context for the application before running it.
