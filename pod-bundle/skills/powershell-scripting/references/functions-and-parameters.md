# Functions and Parameters

Advanced functions, parameters, validation, pipeline input, parameter sets, named blocks, ShouldProcess, and help.

## Contents
- Scripts vs functions vs script blocks
- Basic and advanced functions
- The param block and parameter types
- The Parameter attribute
- Validation attributes
- Pipeline input
- Parameter sets
- begin / process / end / clean
- ShouldProcess (WhatIf / Confirm)
- Managing output
- Splatting
- Comment-based help

## Scripts vs functions vs script blocks

All three are commands and share most capabilities (parameters, pipeline input, common parameters, nesting). A **script** is a `.ps1` file (also supports `using` and `#Requires` statements). A **function** is named, defined with `function`. A **script block** `{ ... }` is an anonymous function used with `ForEach-Object`, `Where-Object`, `Invoke-Command`, `Start-Job`, etc. Script blocks capture parent variables by reference at run time; `.GetNewClosure()` snapshots them.

Keep functions small and single-purpose. Avoid nesting functions inside functions (they become untestable in isolation).

## Basic and advanced functions

```powershell
function Get-Thing {
    [CmdletBinding()]
    param (
        [string] $Name
    )
    "Hello $Name"
}
```

`[CmdletBinding()]` turns a basic function into an **advanced function**, which gains:
- common parameters (`-Verbose`, `-Debug`, `-ErrorAction`, `-ErrorVariable`, `-WarningAction`, `-OutVariable`, `-PipelineVariable`, ...);
- the `$PSCmdlet` automatic variable (for `ShouldProcess`, `ThrowTerminatingError`, `WriteError`, etc.);
- optional `SupportsShouldProcess`, `DefaultParameterSetName`, `PositionalBinding`, `ConfirmImpact`.

Using *any* `[Parameter()]` attribute also makes a function advanced even without `[CmdletBinding()]`. Add `[Alias('gt')]` above `param` to give the command an alias. The inline form `function f($a, $b) {}` works but prevents attributes — prefer a `param` block.

## The param block and parameter types

```powershell
param (
    [string] $Path,
    [int]    $Count = 10,                 # default value
    [switch] $Force,                       # presence = $true; pass -Force:$bool to vary
    [string] $End = ($Path.Length)         # default may reference earlier params
)
```

`param` must come first (after `using`/help). Typed params coerce input. Value types initialize to a default (`[int]`→0, `[bool]`→`$false`, `[string]`→`''`); reference types like `[datetime]`, `[hashtable]` default to `$null`. Command-derived defaults need parentheses.

## The Parameter attribute

```powershell
[Parameter(Mandatory)]                         # prompt if not supplied
[Parameter(Position = 0)]                      # positional
[Parameter(ValueFromPipeline)]                 # bind whole object from pipeline
[Parameter(ValueFromPipelineByPropertyName)]   # bind from a like-named property
[Parameter(ParameterSetName = 'ByName')]
[Parameter(Mandatory, HelpMessage = 'Enter a path')]
[Parameter(DontShow)]                          # hide from completion/IntelliSense
[Parameter(ValueFromRemainingArguments)]       # collect leftover args
```

Boolean properties may be written bare (`Mandatory`, not `Mandatory = $true`). Don't write `Mandatory = $false` — it's the default and adds noise.

## Validation attributes

Validate at bind time so the function body can trust its input:

```powershell
[ValidateNotNull()]                 [ValidateNotNullOrEmpty()]
[ValidateNotNullOrWhiteSpace()]     # 7.x
[ValidateCount(1, 5)]               # array element count
[ValidateLength(8, 64)]             # string length
[ValidateRange(1, 100)]             # numeric range; also 'Positive','NonNegative'
[ValidatePattern('^\d{4}$')]        # regex
[ValidateSet('Dev','Test','Prod')]  # fixed list (tab-completes)
[ValidateScript({ Test-Path $_ })]  # arbitrary; $_ is the value
```

Allow-attributes loosen mandatory/typed params: `[AllowNull()]`, `[AllowEmptyString()]`, `[AllowEmptyCollection()]`. `[PSTypeName('My.Type')]` requires a specific custom-object type name. Example:

```powershell
param (
    [Parameter(Mandatory)]
    [ValidateSet('Start','Stop','Restart')]
    [string] $Action,

    [ValidateRange(1, 3600)]
    [int] $TimeoutSeconds = 30
)
```

## Pipeline input

Two modes, set on the `[Parameter()]`:

- **`ValueFromPipeline`** — binds the whole incoming object. The param type determines what's accepted.
- **`ValueFromPipelineByPropertyName`** — binds from a property of the same name (or an `[Alias()]`) on the incoming object.

Pipeline values are only available in the **`process` block**:

```powershell
function Get-Status {
    [CmdletBinding()]
    param (
        [Parameter(Mandatory, ValueFromPipelineByPropertyName)]
        [Alias('PSPath')]                  # so Get-Item/Get-ChildItem can pipe in
        [string] $Name
    )
    process { "Name: $Name" }
}
```

`[Alias('PSPath')]` on a `$Path` parameter is the standard way to accept piped `Get-Item`/`Get-ChildItem` output. To resolve a path that may not exist yet: `$PSCmdlet.GetUnresolvedProviderPathFromPSPath($Path)`.

## Parameter sets

Group parameters into mutually exclusive sets so a command offers alternative ways to be called:

```powershell
function Get-Item2 {
    [CmdletBinding(DefaultParameterSetName = 'ByName')]
    param (
        [Parameter(ParameterSetName = 'ByName', Position = 0)]
        [string] $Name,

        [Parameter(ParameterSetName = 'ById', Mandatory)]
        [int] $Id
    )
    "Set: $($PSCmdlet.ParameterSetName)"
}
```

Inspect the chosen set with `$PSCmdlet.ParameterSetName`. Set `DefaultParameterSetName` (or give one set a positional parameter) so PowerShell can resolve ambiguous calls — otherwise "Parameter set cannot be resolved."

## begin / process / end / clean

```powershell
function Measure-Item {
    [CmdletBinding()]
    param ([Parameter(ValueFromPipeline)] $InputObject)
    begin   { $count = 0 }              # once, before pipeline
    process { $count++ }                # once per piped item ($_ here)
    end     { $count }                  # once, after pipeline
    clean   { }                         # 7.3+: always runs, even on terminating error
}
```

If no named blocks are declared, all code is in `end` (functions) or `process` (filters). Use `clean` (7.3+) to dispose streams/connections reliably — `end` is skipped if a terminating error fires mid-`process`. `dynamicparam` defines parameters at runtime. `return` ends the current block early but does **not** constrain output — anything left on a line is emitted.

## ShouldProcess (WhatIf / Confirm)

Make state-changing commands safe:

```powershell
function Remove-Thing {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param ([Parameter(Mandatory)] [string] $Name)
    if ($PSCmdlet.ShouldProcess($Name, 'Delete')) {
        # perform the deletion
    }
}
```

`SupportsShouldProcess` adds `-WhatIf` and `-Confirm`. `ShouldProcess` returns `$false` under `-WhatIf` (and prints the "What if:" line) and prompts under `-Confirm` or when `ConfirmImpact` ≥ `$ConfirmPreference` (default `High`). Use `ShouldContinue` only when a prompt must be unconditional (rare); pair it with a `-Force` switch to allow bypass.

## Managing output

PowerShell auto-emits anything on a line; suppress unwanted output explicitly. Best options in 7+:

```powershell
$null = $sb.Append('x')      # clear, fast
[void]$list.Add('x')         # clean for method calls
$sb.Append('x') | Out-Null   # fastest in 7+ (parser-optimized), pipeline-friendly
```

Avoid `return $value` thinking it limits output — it doesn't. Build results as `[PSCustomObject]`.

## Splatting

Pass a hashtable (named) or array (positional) of arguments — keeps lines short and code DRY:

```powershell
$params = @{
    Path        = $src
    Destination = $dst
    Recurse     = $true
    ErrorAction = 'Stop'
}
Copy-Item @params
Copy-Item @params -Force        # combine with explicit params
```

Use `@params` (splat, `@`) not `$params`. Build the hashtable conditionally to include optional parameters only when needed.

## Comment-based help

Place a `<# ... #>` block before `param` (or at the top of a script). Keywords (case-insensitive, but UPPERCASE by convention; spelling matters or help vanishes):

```powershell
function Get-Thing {
    <#
    .SYNOPSIS
        One-line summary.
    .DESCRIPTION
        Longer description.
    .PARAMETER Name
        What Name does.
    .EXAMPLE
        Get-Thing -Name foo
        Explains the example.
    .INPUTS
        System.String
    .OUTPUTS
        PSCustomObject
    .NOTES
        Author, version.
    .LINK
        https://example.com
    #>
    [CmdletBinding()]
    param ([string] $Name)
}
```

Then `Get-Help Get-Thing -Full` shows it. Validate style with PSScriptAnalyzer (`Invoke-ScriptAnalyzer`), which the VS Code PowerShell extension runs automatically.
