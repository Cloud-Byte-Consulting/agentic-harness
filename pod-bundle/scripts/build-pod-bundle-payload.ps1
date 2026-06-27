<#
.SYNOPSIS
  Builds the pod-bundle plugin payload zip.

.DESCRIPTION
  Creates `.github/extensions/pod-bundle/payload/pod-bundle-payload.zip` containing
  all files that must be installed into a workspace when the plugin loads.
#>
[CmdletBinding()]
param(
  [switch]$Force
)

$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$payloadDir = Join-Path $repoRoot '.github\extensions\pod-bundle\payload'
$payloadZip = Join-Path $payloadDir 'pod-bundle-payload.zip'
$staging = Join-Path $env:TEMP ('pod-bundle-payload-' + [guid]::NewGuid().ToString())

$sources = @(
  'templates\pod-bundle',
  'scripts\new-pod-bundle.ps1',
  'scripts\package-pod-bundles.ps1',
  'scripts\build-pod-bundle-package.ps1',
  'scripts\build-pod-bundle-payload.ps1',
  'knowledge-base\43-user-defined-pods-and-zips.md',
  'knowledge-base\41-skill-routing-architecture.md',
  'contracts\skill-routing.yaml',
  '.claude\skills',
  'router-skills'
)

New-Item -ItemType Directory -Path $payloadDir -Force | Out-Null
if (Test-Path -LiteralPath $payloadZip) {
  if (-not $Force) {
    throw "Payload zip already exists: $payloadZip. Use -Force to overwrite."
  }
  Remove-Item -LiteralPath $payloadZip -Force
}

New-Item -ItemType Directory -Path $staging -Force | Out-Null

foreach ($relative in $sources) {
  $src = Join-Path $repoRoot $relative
  if (-not (Test-Path -LiteralPath $src)) {
    throw "Missing source for payload: $src"
  }
  $dst = Join-Path $staging $relative
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

Compress-Archive -Path (Join-Path $staging '*') -DestinationPath $payloadZip -CompressionLevel Optimal
$hash = (Get-FileHash -LiteralPath $payloadZip -Algorithm SHA256).Hash

Remove-Item -LiteralPath $staging -Recurse -Force

[pscustomobject]@{
  payloadZip = (Resolve-Path -LiteralPath $payloadZip).Path
  sha256 = $hash
} | ConvertTo-Json -Depth 3
