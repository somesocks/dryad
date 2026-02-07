# root-build-11

This test verifies dependency invalidation during `dryad root build`.

Coverage:
- build `root-01` with a requirement on `root-02`, then capture root/dependency fingerprints
- mutate `root-02` trait content and rebuild `root-01`
- assert `root-01` fingerprint changes and `dyd/dependencies/root-02` moves to a new stem
- assert pinned requirements include updated dependency traits
- assert dependency stem basename still matches `dyd/fingerprint`
