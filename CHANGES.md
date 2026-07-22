# Changes

Changes made on this branch relative to `main`.

## Unreleased

### Containers

- Split the development and production deployments into
  `docker-compose.dev.yaml` and `docker-compose.prod.yml`.
- Added a hardened production deployment with non-root users, read-only
  filesystems, dropped capabilities, isolated networking, resource limits,
  bounded logs, authenticated logins, and a dedicated database volume.
- Hardened the server image with a static, trimmed Go binary and an
  unprivileged runtime user.
- Added Bash, curl, and the BART import and ICQ user population utilities to
  the server image for fresh-server setup.
- Added the `oscar-admin` management API CLI to the server image with a
  password reset command accepting the password as an argument or on stdin.
- Updated the Go and Debian base images while retaining stunnel 5.76, OpenSSL
  1.0.2u, and Alpine 3.16's NSS tooling for retro client compatibility;
  verified the stunnel source checksum and added a restricted production TLS
  configuration on unprivileged ports.
- Added `.dockerignore` rules for repository metadata, build output,
  certificates, and local databases.
- Updated Make targets, setup scripts, documentation, and the standalone
  stunnel runner for the new Compose filenames and image tags.
- Pointed Compose, local builds, and runtime scripts at the repository's
  `ghcr.io/sharkyrawr/open-oscar-server` package family.
- Stored OSCAR's SQLite database in the host's `./data` directory for both
  development and production Compose deployments.

### CI

- Updated all workflow push and pull-request triggers for the new
  `lunar-build` primary branch.
- Added GitHub Actions workflows that publish the server, certgen, and stunnel
  images to this repository's GHCR packages with generated tags, OCI labels
  and annotations, isolated build caches, SBOMs, and provenance attestations.
- Removed the Codecov upload step from the Go workflow.

### Dependencies

- Updated all Go modules with `go get -u ./...` and normalized `go.mod` and
  `go.sum` with `go mod tidy`.

### Fixes

- Checked session-manager shutdown errors in lifecycle tests.
- Replaced quadratic feedbag ID allocation with a linear occupancy scan,
  eliminating the multi-billion-comparison full-ID test case.
