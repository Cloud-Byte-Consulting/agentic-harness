<#
.SYNOPSIS
  Packages markdown Pod bundles into zip deliverables.

.DESCRIPTION
  Creates one zip per Pod directory and a manifest.json in the output folder.
  A Pod is valid when required markdown files are present.
#>
[CmdletBinding()]
param(
  [string]$PodsRoot = '.\pods',
  [string]$OutputDir = '.\dist\pods',
  [string]$PodName,
  [switch]$Force
)

$ErrorActionPreference = 'Stop'

$requiredFiles = @('README.md', 'pod.md', 'behavior.md', 'sources.md', 'workflows.md')

if (-not (Test-Path -LiteralPath $PodsRoot -PathType Container)) {
  throw "Pods root not found: $PodsRoot"
}

New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$pods = Get-ChildItem -LiteralPath $PodsRoot -Directory
if ($PodName) {
  $pods = @($pods | Where-Object { $_.Name -eq $PodName })
}

if (-not $pods -or $pods.Count -eq 0) {
  throw "No pod directories found to package."
}

$manifest = New-Object System.Collections.Generic.List[object]

foreach ($pod in $pods) {
  $missing = @()
  foreach ($required in $requiredFiles) {
    $requiredPath = Join-Path $pod.FullName $required
    if (-not (Test-Path -LiteralPath $requiredPath -PathType Leaf)) {
      $missing += $required
    }
  }

  if ($missing.Count -gt 0) {
    throw "Pod '$($pod.Name)' is missing required files: $($missing -join ', ')"
  }

  $zipName = "$($pod.Name).zip"
  $zipPath = Join-Path $OutputDir $zipName
  if (Test-Path -LiteralPath $zipPath) {
    if (-not $Force) {
      throw "Zip already exists: $zipPath. Use -Force to overwrite."
    }
    Remove-Item -LiteralPath $zipPath -Force
  }

  Compress-Archive -Path (Join-Path $pod.FullName '*') -DestinationPath $zipPath -CompressionLevel Optimal

  $hash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash
  $manifest.Add([pscustomobject]@{
    pod = $pod.Name
    zip = (Resolve-Path -LiteralPath $zipPath).Path
    sha256 = $hash
    files = $requiredFiles
  }) | Out-Null
}

$manifestPath = Join-Path $OutputDir 'manifest.json'
$manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $manifestPath

[pscustomobject]@{
  outputDir = (Resolve-Path -LiteralPath $OutputDir).Path
  manifest = (Resolve-Path -LiteralPath $manifestPath).Path
  bundles = $manifest
} | ConvertTo-Json -Depth 6
