---
name: powershell-scripting
description: >-
  Write robust, idiomatic PowerShell 7+ scripts, functions, and automation. Use for any .ps1 /
  .psm1 / .psd1 file, the pwsh shell, or tasks involving cmdlets, the object pipeline,
  Get-/Set-/New- verb-noun commands, splatting, advanced functions with
  param/CmdletBinding/parameter validation, modules, error handling (try/catch, throw,
  terminating vs non-terminating, ErrorAction, $Error), working with objects and .NET types,
  files/folders/registry providers, CIM/WMI (Get-CimInstance), regular expressions and text
  parsing, PowerShell remoting (Invoke-Command, PSSession, SSH), classes and enums, execution
  policy and script signing, and reusable automation. Triggers on PowerShell errors like
  "cannot bind argument", "parameter set cannot be resolved", "is not recognized as the name
  of a cmdlet", "execution policy", "term is not recognized", or questions about $_, $PSItem,
  hashtables, ConvertTo-Json, Where-Object, or cross-platform Windows PowerShell 5.1 vs
  PowerShell 7 differences.
---

# PowerShell Scripting

This skill equips Claude to write correct, maintainable PowerShell — from one-line pipelines to advanced functions, modules, and cross-platform automation — targeting PowerShell 7+ while flagging Windows PowerShell 5.1 differences.

## When to use this skill

- Authoring or debugging `.ps1`, `.psm1`, or `.psd1` files, or anything run by `pwsh`.
- Writing functions with parameters, validation, pipeline input, `CmdletBinding`, `WhatIf`/`Confirm`.
- Building or consuming modules; structuring a script into reusable commands.
- Designing error handling: try/catch/finally, `throw`, terminating vs non-terminating errors, `ErrorAction`.
- Working with the object pipeline: `Where-Object`, `ForEach-Object`, `Select-Object`, `Sort-Object`, `Group-Object`, custom objects.
- Files, folders, the registry, CIM/WMI, regular expressions, remoting (WinRM/SSH), classes/enums, and execution policy/signing.
- Any user message mentioning cmdlets, verb-noun commands, `$_`, splatting, hashtables, or PowerShell-specific errors.

## Core concepts

**Everything is an object.** Commands emit .NET objects, not text. The pipeline (`|`) passes objects between commands; each downstream command works on properties and methods, not parsed strings. Use `Get-Member` to discover an object's type and members, and `Get-Command` / `Get-Help -Full` to learn a command:

```powershell
Get-Process | Get-Member                 # what type? what properties/methods?
Get-Process | Where-Object WorkingSet64 -gt 50MB | Select-Object Name, Id
```

**Verb-Noun naming.** Commands are `Verb-Noun` (`Get-Item`, `New-Service`, `Stop-Process`). Use approved verbs (`Get-Verb`) — an unapproved verb warns on module import. This makes commands discoverable (`Get-Command -Verb Get -Noun *Firewall*`). Aliases (`ls`, `?`, `%`) are fine interactively but **never in scripts** — write the full name.

**Two PowerShell editions.** *PowerShell 7+* (the `pwsh` executable, built on modern .NET, cross-platform, open source) is the target. *Windows PowerShell 5.1* (`powershell.exe`, .NET Framework, Windows-only, frozen) is legacy. They install side by side and share most syntax. Key 7+ differences: `Get-WmiObject` removed (use `Get-CimInstance`), `ForEach-Object -Parallel`, ternary `? :`, pipeline chain `&&`/`||`, leading-pipeline line continuation, `Update-Help -Scope CurrentUser`. Note version-sensitive behavior when it matters.

**Pipeline rules of thumb:** filter as far *left* as possible (let the source command filter, or `Where-Object` early); format as far *right* as possible (`Format-*` only at the very end — formatting objects destroys them for further processing).

**Typing is dynamic but type-aware.** Values have .NET types; PowerShell coerces aggressively. `[type]` on the left of a variable/parameter pins the type for all future assignments; on the right it converts the current value only. `$x.GetType()` reveals the type.

## Workflow / how to approach PowerShell tasks

1. **Discover before guessing.** Reach for `Get-Command`, `Get-Help <cmd> -Full`, `Get-Member`, and `Get-Help about_<topic>`. PowerShell is highly self-documenting; do not invent parameter names — verify them.

2. **Prefer cmdlets and the pipeline over reinvention.** There is usually a built-in command. Use `Where-Object`/`ForEach-Object` (or `.Where{}`/`.ForEach{}` methods) rather than manual loops when a pipeline is clearer. Reach into .NET (`[System.IO.File]`, `[regex]`) only when no cmdlet fits.

3. **Build output as objects, not strings.** Emit `[PSCustomObject]@{ Name = ...; Value = ... }` so callers can sort, filter, and export. Reserve `Write-Host` for genuine console UI; use `Write-Output` (implicit) for data, `Write-Verbose`/`Write-Warning`/`Write-Error` for the right streams.

4. **Write advanced functions for anything reused.** Add `[CmdletBinding()]` and a `param()` block. This unlocks common parameters (`-Verbose`, `-ErrorAction`), `$PSCmdlet`, and `-WhatIf`/`-Confirm` support. Declare parameter types, mark mandatory params, validate input with attributes, and accept pipeline input via `ValueFromPipeline[ByPropertyName]` with a `process {}` block. See `references/functions-and-parameters.md`.

5. **Use splatting for commands with many parameters** — it keeps lines short and readable:
   ```powershell
   $params = @{ Path = $src; Destination = $dst; Recurse = $true; Force = $true }
   Copy-Item @params
   ```

6. **Handle errors deliberately.** Decide per operation whether a failure is terminating (stop) or non-terminating (continue processing other items). Wrap risky calls in `try/catch`; set `-ErrorAction Stop` to make a cmdlet's non-terminating error catchable. In advanced functions prefer `$PSCmdlet.ThrowTerminatingError()` over bare `throw` (which is affected by `$ErrorActionPreference`). See `references/error-handling.md`.

7. **Make state-changing commands safe.** Support `-WhatIf`/`-Confirm` via `[CmdletBinding(SupportsShouldProcess)]` + `if ($PSCmdlet.ShouldProcess(...))`. Test destructive operations with `-WhatIf` first.

8. **Mind the platform and version.** On Windows, registry/WMI/`-UseWindowsPowerShell` compatibility apply; on Linux/macOS they do not. Use `[System.IO.Path]::PathSeparator`, `Join-Path`, and `$env:` rather than hard-coded separators. Note 5.1-vs-7 differences when relevant.

## Common pitfalls & anti-patterns

- **`$null` on the wrong side of a comparison.** Write `if ($null -eq $x)`, never `if ($x -eq $null)` — when `$x` is an array the latter *filters* and misbehaves. Comparison operators on an array return matching elements, not a boolean.
- **`Format-*` mid-pipeline.** `Format-Table`/`Format-List` emit formatting objects. Anything after them (export, `Select-Object`) breaks. Format last, or use `Select-Object`/calculated properties to shape data.
- **Aliases and positional args in scripts.** `gci`, `?`, `%`, unnamed positional values hurt readability and break across PowerShell versions/platforms. Use full cmdlet and parameter names.
- **`throw` for control flow in functions.** Bare `throw` is silenced by `-ErrorAction SilentlyContinue` and tears down all parent scopes. Use `throw` only inside `try`; use `$PSCmdlet.ThrowTerminatingError()` to stop cleanly.
- **`Write-Host` for data.** It writes to the information stream and can't be captured or piped as data. Emit objects; use `Write-Host` only for deliberate colored console output.
- **Building arrays with `+=` in a loop.** Arrays are fixed-size; `+=` recreates the whole array each time (O(n²)). Assign the loop's output directly (`$a = foreach (...) { ... }`) or use `[System.Collections.Generic.List[object]]`.
- **Unsuppressed method output.** `$list.Add(x)` and `[StringBuilder].Append(...)` return values that leak into the pipeline. Use `$null = ...`, `[void]...`, or `| Out-Null`.
- **`Get-WmiObject` in PowerShell 7.** Removed. Use `Get-CimInstance` (cross-platform-aware, auto date conversion, WSMan by default).
- **Assuming `-WhatIf` is honored.** It only works if the command author implemented `ShouldProcess` correctly. Test on small data first.
- **Single-quoted vs double-quoted strings.** Variables and subexpressions expand only in `"double"` quotes. Use `'single'` for literals (and for `-replace` substitution patterns like `'$1'`).

## Reference files

- **`references/language-fundamentals.md`** — variables, scopes, types and coercion, arrays, hashtables/ordered dictionaries, operators (comparison, logical, arithmetic, type, redirection), conditionals and loops, switch. Read when writing core logic or debugging type/comparison/array behavior.
- **`references/pipeline-and-objects.md`** — streams, `Get-Member`, selecting/filtering/sorting/grouping, calculated properties, `[PSCustomObject]`, `ForEach-Object -Parallel`, importing/exporting (CSV/JSON), `.Where{}`/`.ForEach{}`. Read for pipeline-heavy work and shaping output.
- **`references/functions-and-parameters.md`** — advanced functions, `CmdletBinding`, `param` blocks, the Parameter attribute, validation attributes, pipeline input, parameter sets, `begin/process/end/clean`, `ShouldProcess`, comment-based help, splatting. Read when authoring any function or script with parameters.
- **`references/error-handling.md`** — terminating vs non-terminating, `try/catch/finally`, `throw` vs `ThrowTerminatingError`, `Write-Error`/`WriteError`, `ErrorAction`/`$ErrorActionPreference`, `$Error`, `Get-Error`, error records, `trap`. Read for any robust error strategy.
- **`references/modules-and-structure.md`** — module anatomy (.psd1/.psm1), `Get-/Import-/Remove-Module`, `PSModulePath`, the Gallery and `Install-Module`/`PSResourceGet`, repositories, plus security: execution policy, signing, `#Requires`. Read when packaging code or dealing with module/policy/signing issues.
- **`references/remoting.md`** — `Invoke-Command`, `Enter-PSSession`, `New-PSSession`, WinRM vs SSH transport, `ArgumentList`/`using:`, parallelism, capturing failures, CIM sessions, the double-hop problem, JEA. Read for remote management tasks.
- **`references/regex-and-text.md`** — `-match`/`-notmatch`/`-replace`/`-split`, `$matches`, capture groups and named groups, `switch -Regex`, `[regex]` class, common parsing recipes. Read for text parsing and pattern matching.
- **`references/cim-wmi-and-providers.md`** — CIM/WMI cmdlets, WQL, methods/instances, plus the provider model: files, folders, registry, `Get-/Set-/Test-Path`, content cmdlets, encoding, ACLs. Read for system inventory, registry, or filesystem-permission work (Windows-centric).
- **`references/classes-and-automation-patterns.md`** — `class`/`enum`, properties/constructors/methods, inheritance, interfaces; and reusable automation patterns (idempotency, logging, config, scheduled scripts, parallelism). Read for object-oriented code or designing durable automation.
