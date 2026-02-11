### Release Process

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

#### Creating a Release

1. **Create and push a tag:**

   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Run tests
   - Build cross-platform binaries
   - Create release archives
   - Generate checksums
   - Create a GitHub release
   - Upload artifacts
