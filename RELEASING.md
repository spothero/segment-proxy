# Releasing

1. Verify the build by running `make build test docker_build`.
2. Update the `CHANGELOG.md` for the impending release.
3. Run `git tag -a X.Y.Z -m "Version X.Y.Z"` (where X.Y.Z is the new version).
4. Run `git push && git push --tags`.
5. Run `VERSION=X.Y.Z make docker`.
