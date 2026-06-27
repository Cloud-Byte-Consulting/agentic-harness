# Classes, Enums, and Automation Patterns

Object-oriented PowerShell (`enum`/`class`) plus durable automation patterns.

## Contents
- Enumerations
- Flags enums
- Classes
- Properties, constructors, methods
- Hidden and static members
- Inheritance
- Interfaces
- When to use a class
- Automation patterns

## Enumerations

A named set of numeric constants. Defined with `enum`; parsed before run (position-independent), so `using module` is needed to consume an enum exported by a module.

```powershell
enum Environment {
    Dev   = 1
    Test  = 2
    Prod  = 3
}
[Environment]::Prod          # Prod
[int][Environment]::Prod     # 3
[Environment]::Prod.value__  # 3 (underlying numeric)
'Test' -as [Environment]     # Test (parse a name)
```

Names must start with a letter/underscore (no hyphens, no quotes); values may repeat (aliases). Default underlying type is `Int32`; 7+ allows `enum X : byte { ... }` (any integer type). Without explicit values, members auto-number from 0. Enums make great parameter types — self-documenting and they tab-complete, similar to `[ValidateSet()]` but reusable and type-checked.

## Flags enums

`[Flags]` lets values combine bitwise — assign powers of two:

```powershell
[Flags()] enum Permission {
    None    = 0
    Read    = 1
    Write   = 2
    Execute = 4
}
$p = [Permission]'Read, Write'         # Read, Write
$p -band [Permission]::Write           # test a bit
$p = $p -bor [Permission]::Execute     # add a bit
```

This mirrors .NET enums like `[System.Security.AccessControl.FileSystemRights]`.

## Classes

Define a type with `class`. Instantiate with `::new()`, `New-Object`, or by casting a hashtable/`[PSCustomObject]`:

```powershell
class Server {
    [string] $Name
    [int]    $Port = 443
}
[Server]::new()
[Server]@{ Name = 'web01'; Port = 8443 }   # cast a hashtable
```

## Properties, constructors, methods

```powershell
class Server {
    [string] $Name
    [int]    $Port

    Server() { $this.Port = 443 }                 # default constructor
    Server([string] $name) {                      # overloaded constructor
        $this.Name = $name
        $this.Port = 443
    }

    [string] ToString() { return "$($this.Name):$($this.Port)" }   # override
    [bool] Test() { return Test-Connection $this.Name -Quiet }
}
[Server]::new('web01').ToString()
```

- `$this` refers to the instance.
- Constructors share the class name; each overload needs a unique signature (argument count/types). There is no `this.base()` chaining — see inheritance.
- **Methods must use `return`** (unlike functions) and must return a value on every path if a return type is declared; they don't auto-emit. Methods can be overloaded.
- Property types are fixed; default values optional.

## Hidden and static members

```powershell
class Cache {
    static [hashtable] $Store = @{}            # shared across all instances
    hidden [datetime] $Created = (Get-Date)    # hidden from Get-Member/completion
    static [void] Clear() { [Cache]::Store.Clear() }
}
[Cache]::Store          # access static via the type
[Cache]::Clear()
```

`hidden` members still work, just don't show in `Get-Member` (use `-Force`) or tab completion. `static` members belong to the type, not instances.

## Inheritance

Single inheritance with `:`; child gets the parent's members and can override them:

```powershell
class Base {
    [string] $Kind = 'base'
    [string] Describe() { return 'base' }
}
class Derived : Base {
    [string] $Extra = 'x'
    [string] Describe() { return 'derived' }   # override (return type may change)
}
```

Constructors run **parent then child** and are not inherited automatically — each child declares its own and may chain to a base constructor:

```powershell
class Child : Parent {
    Child([string] $name) : base($name) { }    # call the base constructor
}
```

A method's return type may change when overridden; a **property's type cannot** (causes an "ambiguous match" error on instantiation). Inheritance can be arbitrarily deep.

## Interfaces

Implement .NET interfaces (also via `:`) to plug into framework behavior — `IComparable` (sorting), `IEquatable` (equality), `IDisposable` (cleanup):

```powershell
class Version : System.IComparable {
    [int] $Major
    [int] CompareTo([object] $other) { return $this.Major - $other.Major }
}
```

Classes can also back DSC resources (`[DscResource()]` with `Get`/`Set`/`Test` methods) — niche.

## When to use a class

Most PowerShell needs `[PSCustomObject]`, hashtables, and functions, not classes. Reach for a class when you need: a strong custom *type* (for parameters, `-is` checks, validation), methods that mutate internal state, inheritance, or a .NET interface implementation. Prefer functions + custom objects for ordinary scripting — classes have rough edges (runspace affinity, no easy `Export`, parse-time loading).

## Automation patterns

**Idempotency — check before change.** Make scripts safe to re-run: test current state, act only on drift.
```powershell
if (-not (Test-Path $dir)) { New-Item $dir -ItemType Directory }
$svc = Get-Service MyApp
if ($svc.Status -ne 'Running') { $svc | Start-Service }
```

**ShouldProcess for destructive ops.** Always gate changes behind `-WhatIf`/`-Confirm` (see functions-and-parameters.md) so operators can dry-run.

**Splatting for config-driven calls.** Build parameter hashtables, optionally from a config file, and splat — keeps long commands readable and parameterizable.

**Structured config.** Keep settings in JSON/PSD1, not hard-coded:
```powershell
$config = Get-Content config.json -Raw | ConvertFrom-Json
$config = Import-PowerShellDataFile config.psd1   # safe .psd1 loader, no code execution
```

**Logging via the right streams.** Use `Write-Verbose`/`Write-Warning`/`Write-Error` (controllable, capturable) instead of `Write-Host`. Add `[CmdletBinding()]` so callers get `-Verbose`. For an audit trail, append `[PSCustomObject]` records to a CSV (`Export-Csv -Append`) or use `Start-Transcript`.

**Robust loops over many targets.** Pair `try`/`catch` (or `-ErrorAction SilentlyContinue -ErrorVariable`) per item so one failure doesn't abort the batch; collect successes and failures separately. For remote fan-out use `Invoke-Command -ComputerName @(...)` (parallel, gather errors via `-ErrorVariable`).

**Parallelism.** `ForEach-Object -Parallel` (7+) for I/O-bound work; `Start-ThreadJob` (in-process, light) or `Start-Job` (separate process, heavier, fully isolated) for background work; remoting parallelizes by default. Mind thread-safety — return output to the pipeline rather than mutating shared collections.

**Scheduled/unattended scripts.** Add a `#Requires` header; start with `pwsh -NoProfile -File` so a user profile can't change behavior; never prompt (avoid `Read-Host`, `Get-Credential`, unconditional `ShouldContinue`); store secrets in a vault (`Microsoft.PowerShell.SecretManagement`) or a managed identity, never inline; set `$ErrorActionPreference = 'Stop'` at the top and wrap the body in `try`/`catch` to fail loudly and log; exit with a meaningful code (`exit 1`) so the scheduler detects failure.

**Reusability.** Promote repeated logic into advanced functions, then into a module (manifest + explicit `FunctionsToExport`). Validate with `Invoke-ScriptAnalyzer` and test with Pester.
