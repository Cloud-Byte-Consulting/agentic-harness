<#
.SYNOPSIS
  Build a single zip package for the pod-bundle toolkit.

.DESCRIPTION
  Produces one deliverable zip containing:
  - markdown pod templates
  - scaffold + packaging scripts
  - full skills payload (shared + router-skills)
  - user guide docs
  - pod-bundle extension plugin
  - quickstart README
#>
[CmdletBinding()]
param(
  [string]$OutputDir = '.\dist',
  [switch]$Force
)

$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$packageRoot = Join-Path $OutputDir 'pod-bundle-package'
$zipPath = Join-Path $OutputDir 'pod-bundle-package.zip'

# Refresh plugin payload so the extension can install all required files on load.
& (Join-Path $PSScriptRoot 'build-pod-bundle-payload.ps1') -Force | Out-Null

New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

if (Test-Path -LiteralPath $packageRoot) {
  Remove-Item -LiteralPath $packageRoot -Recurse -Force
}

if (Test-Path -LiteralPath $zipPath) {
  if (-not $Force) {
    throw "Output zip already exists: $zipPath. Use -Force to overwrite."
  }
  Remove-Item -LiteralPath $zipPath -Force
}

New-Item -ItemType Directory -Path $packageRoot -Force | Out-Null

$map = @(
  @{ Source = 'templates\pod-bundle'; Target = 'templates\pod-bundle' },
  @{ Source = 'scripts\new-pod-bundle.ps1'; Target = 'scripts\new-pod-bundle.ps1' },
  @{ Source = 'scripts\package-pod-bundles.ps1'; Target = 'scripts\package-pod-bundles.ps1' },
  @{ Source = 'scripts\build-pod-bundle-package.ps1'; Target = 'scripts\build-pod-bundle-package.ps1' },
  @{ Source = 'scripts\build-pod-bundle-payload.ps1'; Target = 'scripts\build-pod-bundle-payload.ps1' },
  @{ Source = '.github\extensions\pod-bundle\extension.mjs'; Target = '.github\extensions\pod-bundle\extension.mjs' },
  @{ Source = '.github\extensions\pod-bundle\README.md'; Target = '.github\extensions\pod-bundle\README.md' },
  @{ Source = '.github\extensions\pod-bundle\payload'; Target = '.github\extensions\pod-bundle\payload' },
  @{ Source = 'knowledge-base\43-user-defined-pods-and-zips.md'; Target = 'docs\43-user-defined-pods-and-zips.md' },
  @{ Source = 'knowledge-base\41-skill-routing-architecture.md'; Target = 'docs\41-skill-routing-architecture.md' },
  @{ Source = 'contracts\skill-routing.yaml'; Target = 'contracts\skill-routing.yaml' },
  @{ Source = '.claude\skills'; Target = '.claude\skills' },
  @{ Source = 'router-skills'; Target = 'router-skills' }
)

foreach ($entry in $map) {
  $src = Join-Path $repoRoot $entry.Source
  if (-not (Test-Path -LiteralPath $src)) {
    throw "Required source not found: $src"
  }
  $dst = Join-Path $packageRoot $entry.Target
  $dstParent = Split-Path -Parent $dst
  if ($dstParent) {
    New-Item -ItemType Directory -Path $dstParent -Force | Out-Null
  }
  if (Test-Path -LiteralPath $src -PathType Container) {
    Copy-Item -LiteralPath $src -Destination $dst -Recurse -Force
  } else {
    Copy-Item -LiteralPath $src -Destination $dst -Force
  }
}

$readme = @"
# Pod Bundle Toolkit (Quickstart)

This package lets you define Pods in markdown, includes all Pod/skill assets, and
ships each Pod as a single zip.

## 1) Create a new Pod

```powershell
.\scripts\new-pod-bundle.ps1 -Name "My Operations Pod"
```

This creates:

```text
pods\my-operations-pod\
  README.md
  pod.md
  behavior.md
  sources.md
  workflows.md
```

## 2) Customize behavior in markdown

Edit:

- `pod.md` (identity, owners, routing hints)
- `behavior.md` (prioritization and escalation rules)
- `sources.md` (data connectors and evidence policy)
- `workflows.md` (playbooks)

## 3) Package a Pod as one zip

Single Pod:

```powershell
.\scripts\package-pod-bundles.ps1 -PodsRoot .\pods -PodName my-operations-pod -Force
```

All Pods:

```powershell
.\scripts\package-pod-bundles.ps1 -PodsRoot .\pods -Force
```

Output:

- `dist\pods\<pod-name>.zip`
- `dist\pods\manifest.json` (SHA256 hashes)

## Plugin mode

This package also includes a Copilot extension plugin at:

`.github\extensions\pod-bundle\extension.mjs`

When the plugin loads, it auto-installs the bundled payload (including skills)
into the correct workspace locations once.

Tool names:

- `pod_bundle_scaffold`
- `pod_bundle_package`
- `pod_bundle_package_toolkit`
- `pod_bundle_install_payload`

Reload extensions after copying into a repo:

```powershell
extensions_reload
```
"@

Set-Content -LiteralPath (Join-Path $packageRoot 'README.md') -Value $readme -NoNewline

Compress-Archive -Path (Join-Path $packageRoot '*') -DestinationPath $zipPath -CompressionLevel Optimal

$hash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash
[pscustomobject]@{
  packageRoot = (Resolve-Path -LiteralPath $packageRoot).Path
  zip = (Resolve-Path -LiteralPath $zipPath).Path
  sha256 = $hash
} | ConvertTo-Json -Depth 4
