# WinGet Packaging

This folder contains a ready-to-submit WinGet manifest set for `JGador.GitSemantic` version `0.1.0`.

## Current Package Metadata

- Package identifier: `JGador.GitSemantic`
- Package version: `0.1.0`
- Publisher: `JGador`
- Release repository: `https://github.com/jgador/gitsemantic`
- Release tag: `v0.1.0`
- Windows asset: `gitsemantic_0.1.0_windows_amd64.zip`
- SHA256: `CD097D89A498FAA52AE5FB77561F545B87DCEBE7C1CB8038BF5AE77CC977346A`
- Nested portable command: `gitsemantic`

The `v0.1.0` Windows archive contains `gitsemantic.exe` at the archive root, so the installer manifest uses `InstallerType: zip` with a nested portable executable.

## Local Validation Workflow

1. Install the Windows Package Manager tooling on a Windows machine with `winget`.
2. Run the local-manifest feature enablement once from an elevated PowerShell session:

   ```powershell
   winget settings --enable LocalManifestFiles
   ```

3. Optionally download the Windows release asset and verify its SHA256:

   ```powershell
   winget hash .\gitsemantic_0.1.0_windows_amd64.zip
   ```

4. Validate the manifest directory:

   ```powershell
   winget validate .\packaging\winget\manifests\j\JGador\GitSemantic\0.1.0
   ```

5. Test install from the local manifest:

   ```powershell
   winget install --manifest .\packaging\winget\manifests\j\JGador\GitSemantic\0.1.0
   ```

6. Verify the command:

   ```powershell
   gitsemantic version
   ```

## Public Submission Workflow

1. Clone `microsoft/winget-pkgs`.
2. Copy the files from `manifests/j/JGador/GitSemantic/0.1.0/` into the matching path in `winget-pkgs`.
3. Run `winget validate` again in that repo if your local tooling supports it.
4. Open a pull request to `microsoft/winget-pkgs`.
5. After the PR is merged, users can install with:

   ```powershell
   winget install --id JGador.GitSemantic
   ```

## Updating For A New Release

1. Push a new tag such as `v0.1.1` to publish GitHub Release assets.
2. Update `PackageVersion` in all three manifest files.
3. Update the Windows asset URL and SHA256 in the installer manifest.
4. Update this document if the package identity, publisher, or release repository changes.
5. Re-run local validation and install checks before opening the next `winget-pkgs` PR.

## Metadata Notes

The public GitHub repository currently does not declare a license through GitHub's repository metadata. The manifest therefore uses `License: Proprietary` until the upstream release repository publishes an explicit license that should be reflected in WinGet.
