<#
.SYNOPSIS
  Scaffolds a markdown-only Pod bundle from templates.

.DESCRIPTION
  Creates a Pod folder containing markdown files users can edit to define Pod
  behavior without changing code defaults.
#>
[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)]
  [ValidatePattern('^[a-zA-Z0-9][a-zA-Z0-9\- ]*$')]
  [string]$Name,
  [string]$PodsRoot = '.\pods',
  [switch]$Force
)

$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$templateRoot = Join-Path $repoRoot 'templates\pod-bundle'
if (-not (Test-Path -LiteralPath $templateRoot -PathType Container)) {
  throw "Template folder not found: $templateRoot"
}

$safeName = ($Name -replace '\s+', '-').ToLowerInvariant()
$podPath = Join-Path $PodsRoot $safeName

if (Test-Path -LiteralPath $podPath) {
  if (-not $Force) {
    throw "Pod path already exists: $podPath. Use -Force to overwrite."
  }
  Remove-Item -LiteralPath $podPath -Recurse -Force
}

New-Item -ItemType Directory -Path $podPath -Force | Out-Null
Copy-Item -Path (Join-Path $templateRoot '*') -Destination $podPath -Recurse -Force

$podId = $safeName
$files = Get-ChildItem -LiteralPath $podPath -File -Filter '*.md'
foreach ($file in $files) {
  $content = Get-Content -LiteralPath $file.FullName -Raw
  $content = $content.Replace('__POD_NAME__', $Name)
  $content = $content.Replace('__POD_ID__', $podId)
  Set-Content -LiteralPath $file.FullName -Value $content -NoNewline
}

[pscustomobject]@{
  podName = $Name
  podId = $podId
  podPath = (Resolve-Path -LiteralPath $podPath).Path
  files = @($files.Name)
} | ConvertTo-Json -Depth 4
