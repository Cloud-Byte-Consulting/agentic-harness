# Modules, Script Structure, and Security

Finding/using/building modules, `PSModulePath`, the Gallery, plus `#Requires`, execution policy, and signing.

## Contents
- Module anatomy
- Discovering and importing modules
- PSModulePath
- The PowerShell Gallery and installing modules
- PSResourceGet (PowerShellGet 3)
- Private repositories
- Building a module
- Script structure and #Requires
- Execution policy
- Script signing

## Module anatomy

A module is a packaged set of commands. Files typically in a versioned directory `Modules\MyModule\1.2.0\`:

- **`.psd1` manifest** — metadata: `ModuleVersion`, `RootModule`, `Author`, `FunctionsToExport`, `RequiredModules`, `PowerShellVersion`, `CompatiblePSEditions`, GUID.
- **`.psm1` root module** — the code (or a `.dll` for a binary module).
- `.xml` help/format files, plus any supporting files.

Most modules are self-contained and portable (just copy the folder). Exceptions: OS-component modules (`NetAdapter`, `Storage`, `ActiveDirectory`) depend on Windows features/WMI and aren't portable.

## Discovering and importing modules

```powershell
Get-Module                       # loaded in this session
Get-Module -ListAvailable        # installed and discoverable
Get-Command -Module Pester       # commands a module exports
Import-Module Pester
Import-Module Pester -MinimumVersion 5.4.0
Remove-Module Pester             # unload from session (DLLs stay loaded until restart)
```

**Autoloading** (PS 3+): running a command from a discoverable module imports it automatically — explicit `Import-Module` is rarely needed. Disable with `$PSModuleAutoLoadingPreference = 'None'`.

**Windows PowerShell compatibility (7+, Windows):** a module marked Desktop-only loads in a background WinPS compatibility session (`Import-Module X -UseWindowsPowerShell`). Objects come back *deserialized* — properties work, methods may not. `-SkipEditionCheck` forces a native load (may fail). Prefer native 7-compatible modules.

## PSModulePath

`$env:PSModulePath` is a delimited list of module search roots, split by `[System.IO.Path]::PathSeparator` (`;` Windows, `:` Linux/macOS). PowerShell searches paths **in order** and uses the highest version from the **first** path that contains the module — it does *not* scan all paths for the newest. Default 7+ roots: `$HOME\Documents\PowerShell\Modules` (CurrentUser), `$env:ProgramFiles\PowerShell\Modules` (AllUsers), `$PSHOME\Modules` (shipped). Edits to `$env:PSModulePath` are process-scoped (set them in a profile).

## The PowerShell Gallery and installing modules

The Gallery (powershellgallery.com) is a public NuGet-based repository.

```powershell
Find-Module -Name Carbon
Find-Module -Filter IIS                      # search tags/description
Install-Module Carbon                         # CurrentUser scope by default
Install-Module Carbon -Scope AllUsers         # needs admin
Update-Module Carbon
Save-Module Carbon -Path C:\Modules           # download without installing
```

`Install-Module` cannot install under `$PSHOME` (reserved for shipped modules). Newer versions install side by side; `-Force` reinstalls.

## PSResourceGet (PowerShellGet 3)

`Microsoft.PowerShell.PSResourceGet` ships with 7.4+ and replaces PowerShellGet 2. It treats modules and scripts uniformly as "resources" and drops the PackageManagement/NuGet bootstrap. New command names let it coexist with the old module:

```powershell
Register-PSResourceRepository -PSGallery
Find-PSResource -Name Indented.Net.IP
Install-PSResource Pester
Install-PSResource Pester -Version '[5.0,6.0)'   # NuGet range syntax
Find-PSResource -Name PowerShellGet -Version * -Prerelease
```

## Private repositories

For curated/internal content. Simplest is an SMB share registered as a repository:

```powershell
Register-PSRepository -Name Internal -SourceLocation '\\server\share\modules' -InstallationPolicy Trusted
Publish-Module -Name Pester -RequiredVersion 5.4.0 -Repository Internal
```

For richer auth/lifecycle, host a NuGet feed (Sonatype Nexus, ProGet, or Chocolatey.Server on IIS) and authenticate with an API key when publishing.

## Building a module

1. Create `MyModule.psm1` with your functions; `Export-ModuleMember -Function Verb-Noun, ...` (or list `FunctionsToExport` in the manifest — explicit exports beat `*`).
2. Generate a manifest: `New-ModuleManifest -Path MyModule.psd1 -RootModule MyModule.psm1 -ModuleVersion 1.0.0 -Author '...' -FunctionsToExport @('Get-Thing')`.
3. Place under a `PSModulePath` root in a version folder.
4. Validate: `Test-ModuleManifest MyModule.psd1`, then `Invoke-ScriptAnalyzer` for style/correctness.

A common pattern is one `.ps1` file per function under `Public/` and `Private/`, dot-sourced from the `.psm1`. Store shared module state in `$script:` variables. Snap-ins are legacy (Windows PowerShell only) — never author them.

## Script structure and #Requires

Order in a `.ps1`: `#Requires` comments → `using` statements → comment-based help → `[CmdletBinding()]`/`param` → code. `#Requires` is parsed before the script runs and blocks execution if unmet:

```powershell
#Requires -Version 7.4
#Requires -Modules @{ ModuleName = 'Pester'; ModuleVersion = '5.4.0' }
#Requires -RunAsAdministrator
#Requires -PSEdition Core
```

`using namespace System.IO` shortens type names; `using module MyModule` imports PowerShell classes at parse time (needed to use a module's `class` types). `using assembly 'C:\path\lib.dll'` loads a DLL (7+ needs a full literal path).

## Execution policy (Windows only)

Controls whether/which `.ps1` files run. It is a safety feature, **not a security boundary** (trivially bypassed) — never rely on it to stop a determined user.

```powershell
Get-ExecutionPolicy -List                        # effective policy per scope
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
pwsh -ExecutionPolicy Bypass -File .\script.ps1   # per-invocation override
```

Values: `Restricted` (no scripts), `AllSigned` (all must be signed), `RemoteSigned` (downloaded scripts must be signed; local ones run — common dev default), `Unrestricted`, `Bypass`, `Undefined`. Scopes (most→least specific): `MachinePolicy`, `UserPolicy` (Group Policy, win), `Process`, `CurrentUser`, `LocalMachine`. On Linux/macOS execution policy is `Unrestricted` and not enforced.

## Script signing (Windows)

Under `AllSigned`/`RemoteSigned`, sign scripts with an Authenticode code-signing certificate:

```powershell
$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Select-Object -First 1
Set-AuthenticodeSignature -FilePath .\script.ps1 -Certificate $cert `
    -TimestampServer 'http://timestamp.digicert.com'
Get-AuthenticodeSignature .\script.ps1            # verify status: Valid / NotSigned / etc.
```

Timestamping keeps the signature valid after the cert expires. The signing cert's CA must be trusted on machines that run the script (a self-signed dev cert must be imported into Trusted Publishers / Root). Code-signing certs live in the `Cert:` provider (`Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert`).
