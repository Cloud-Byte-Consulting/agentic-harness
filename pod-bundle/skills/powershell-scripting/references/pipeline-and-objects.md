# Pipeline and Objects

The object pipeline, discovery, filtering/selecting/sorting/grouping, custom objects, and import/export.

## Contents
- Streams
- Discovering members
- Where-Object and ForEach-Object
- Select-Object
- Sort-Object, Group-Object, Measure-Object, Compare-Object
- Custom objects
- Parallel processing
- Import / export / convert

## Streams

A command's normal output goes to the **success** stream (1); it's what gets captured by `$x = cmd` or passed down a pipeline. Other streams are separate:

| Stream | Cmdlet | # | Preference variable |
|---|---|---|---|
| Success/Output | `Write-Output` | 1 | — |
| Error | `Write-Error` | 2 | `$ErrorActionPreference` |
| Warning | `Write-Warning` | 3 | `$WarningPreference` |
| Verbose | `Write-Verbose` | 4 | `$VerbosePreference` |
| Debug | `Write-Debug` | 5 | `$DebugPreference` |
| Information | `Write-Information` | 6 | `$InformationPreference` |

`Write-Host` writes to the information stream (since 5.1) — use it only for deliberate console UI (e.g. `-ForegroundColor`), never to emit data. `Write-Verbose`/`-Debug` are silent unless `-Verbose`/`-Debug` (or the preference variable) is set. Emit *data* by just leaving it on a line (implicit `Write-Output`).

## Discovering members

`Get-Member` is the primary discovery tool — pipe any object to it to see its type and members:

```powershell
Get-Process -Id $PID | Get-Member                      # all members + TypeName
Get-Process | Get-Member -MemberType Property          # just properties
$obj | Get-Member -Force                                # include hidden PS* members
```

Accessors `{get;set;}` (read/write) vs `{get;}` (read-only). Access properties with `.Name`, or on a command with `(Get-Process -Id $PID).StartTime`. Methods: `$obj.Method(args)`; call a method with no parens to see overloads: `(Get-Date).AddDays`. Chain because each call returns a new object: `(Get-Date).Date.AddSeconds(-1).AddDays(1.5)`.

## Where-Object and ForEach-Object

**Filter** with `Where-Object` (alias `?`, avoid in scripts). Simple comparison form vs script-block form:

```powershell
Get-Process | Where-Object WorkingSet64 -gt 50MB
Get-Service | Where-Object { $_.Status -eq 'Running' -and $_.StartType -eq 'Manual' }
```

`$_` (or `$PSItem`) is the current pipeline object. Use the script-block form for multiple/compound conditions or method calls.

**Iterate/transform** with `ForEach-Object` (aliases `%`, `foreach`):

```powershell
Get-Process | ForEach-Object { $_.Name.ToUpper() }
Get-Process | ForEach-Object -MemberName Path            # pull one property
@(d1, d2) | ForEach-Object ToString('yyyyMMdd')          # call a method on each
1..5 | ForEach-Object -Begin { $sum=0 } -Process { $sum+=$_ } -End { $sum }
```

`-Begin`/`-End` run once before/after the per-item `-Process` block. For pure speed over an in-memory collection, the `foreach` keyword or the `.ForEach{}` method beats the cmdlet (no pipeline overhead). The `.Where{}`/`.ForEach{}` methods also support modes: `.Where({ $_ -gt 5 }, 'First', 1)`.

## Select-Object

Shape, limit, and project objects:

```powershell
Get-Process | Select-Object Name, Id
Get-Process | Select-Object *                     # all properties (vs Get-Member)
Get-Process | Select-Object -First 5 / -Last 3 / -Skip 4 / -Index 1,3,5
1,1,2,3,3 | Select-Object -Unique                 # case-sensitive for strings
Get-Process | Select-Object -ExpandProperty Path  # unwrap one property to its values
```

**Calculated properties** add/rename properties via a hashtable with `Name`/`Label` (or `n`/`l`) and `Expression` (or `e`):

```powershell
Get-Process | Select-Object Name,
    @{ Name = 'ProcessId'; Expression = 'Id' },                 # rename
    @{ n = 'WS_MB'; e = { [math]::Round($_.WorkingSet64/1MB, 1) } }  # compute
```

**Note:** `Select-Object -Property` returns a new `PSCustomObject` and **loses the original type and methods**. If you need the object's methods later, use `Add-Member` instead, or don't project.

## Sort-Object, Group-Object, Measure-Object, Compare-Object

```powershell
Get-Process | Sort-Object WorkingSet64 -Descending
Get-ChildItem | Sort-Object LastWriteTime, Name          # multi-key
$data | Sort-Object { switch ($_.Result) { 'Pass'{1} 'Fail'{2} default{3} } }, Mark
$data | Sort-Object @{ Expression = 'Mark'; Descending = $true }   # per-key direction

Get-Process | Group-Object Company                        # groups with .Group + .Count
Get-Process | Group-Object Company -AsHashTable -AsString # fast lookup table
1..10 | Measure-Object -Sum -Average -Maximum             # stats
Compare-Object $ref $diff -Property Name                  # set difference (<= / =>)
Get-ChildItem | Sort-Object Length | Get-Unique           # Get-Unique needs sorted input
```

`Measure-Command { ... }` times a block.

## Custom objects

`[PSCustomObject]` is the idiomatic way to produce structured output. Property order is preserved:

```powershell
[PSCustomObject]@{
    ComputerName = $env:COMPUTERNAME
    OS           = (Get-CimInstance Win32_OperatingSystem).Caption
    FreeGB       = [math]::Round($disk.Free / 1GB, 1)
}
```

Build incrementally with `[ordered]@{}` when properties depend on conditions, then cast:

```powershell
$o = [ordered]@{ Name = $name }
if ($email) { $o.Email = $email }
[PSCustomObject]$o
```

`Add-Member` attaches members to an *existing* object while preserving its type:

```powershell
$item | Add-Member -NotePropertyName Source -NotePropertyValue $env:COMPUTERNAME -PassThru
```

`New-Object` and `Add-Member`-built objects are legacy patterns superseded by `[PSCustomObject]` (faster, ordered). Use `New-Object` for actual .NET/COM instantiation only.

## Parallel processing (7+)

```powershell
1..10 | ForEach-Object -Parallel { Start-Sleep 1; $_ } -ThrottleLimit 5
```

Each block runs in a separate runspace. Outside variables need the `using:` modifier (`$using:config`), and it's read-only — collections can be mutated via `($using:list).Add($_)` but aren't thread-safe; prefer returning output to the pipeline. Default `-ThrottleLimit` is 5. For CPU-bound work this helps; for trivial work the runspace overhead can make it slower than a plain loop. See also `Start-ThreadJob` and runspaces.

## Import / export / convert

```powershell
$data | Export-Csv data.csv -NoTypeInformation     # -NoType implicit in 7+
Import-Csv data.csv

$obj | ConvertTo-Json -Depth 5 | Set-Content api.json
Get-Content api.json -Raw | ConvertFrom-Json       # -Raw reads whole file as one string

$obj | Export-Clixml state.xml                      # full fidelity round-trip (PS types)
Import-Clixml state.xml
```

`ConvertTo-Json` defaults to `-Depth 2` — raise it for nested objects or data silently truncates. `Invoke-RestMethod` auto-parses JSON/XML responses; `Invoke-WebRequest` returns the raw response object.
