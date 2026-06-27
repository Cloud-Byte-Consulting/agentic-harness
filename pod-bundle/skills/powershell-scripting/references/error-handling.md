# Error Handling

Terminating vs non-terminating errors, try/catch/finally, raising errors, ErrorAction, and the quirks.

## Contents
- Two kinds of error
- Error actions and preferences
- Raising non-terminating errors
- Raising terminating errors
- Catching: try / catch / finally
- Catching: ErrorVariable
- The throw gotcha and recommendations
- Error records and Get-Error
- trap

## Two kinds of error

- **Non-terminating** — informational; the command writes the error and keeps going. Right choice when processing many items and one item's failure shouldn't stop the rest (e.g. a pipeline). Raised with `Write-Error` or `$PSCmdlet.WriteError()`.
- **Terminating** — stops execution. Right choice when the operation genuinely cannot continue (failed to read a critical config, validation failed hard). Raised with `throw` or `$PSCmdlet.ThrowTerminatingError()`.

```powershell
function Update-Value {
    [CmdletBinding()]
    param ([Parameter(Mandatory, ValueFromPipeline)] [string] $Value)
    process {
        if ($Value.Length -lt 5) {
            Write-Error "Value '$Value' is too short"   # non-terminating: continue
        } else { "Updated: $Value" }
    }
}
'value','val','longvalue' | Update-Value   # error on 'val', others still processed
```

## Error actions and preferences

`-ErrorAction` (per command) and `$ErrorActionPreference` (scoped variable) control non-terminating-error behavior:

| Value | Effect |
|---|---|
| `Continue` (default) | display, keep going |
| `Stop` | turn it into a *terminating* error (catchable) |
| `SilentlyContinue` | suppress display, still record in `$Error` |
| `Ignore` | suppress and **don't** record |
| `Inquire` | prompt |

`-ErrorAction` is available on any advanced function/cmdlet. Setting `-ErrorAction Stop` on a cmdlet is the standard way to make its non-terminating error catchable by `try/catch`. `$Error` is an auto array of every error this session, newest at `$Error[0]`.

## Raising non-terminating errors

```powershell
Write-Error -Message 'Something failed'
Write-Error -Message 'Bad input' -Category InvalidArgument -ErrorId 'BadInput'
```

In an advanced function, `$PSCmdlet.WriteError($errorRecord)` is preferable — unlike `Write-Error` it correctly sets `$?` to `$false` and reports the function (not `Write-Error`) as the activity. Build an `ErrorRecord` for rich, debuggable errors (see below).

## Raising terminating errors

```powershell
throw 'Error message'
throw [System.ArgumentException]::new('Unsupported value')
```

In an advanced function prefer the method, which is **not** affected by `$ErrorActionPreference` and behaves like a cmdlet error (stops the function, not all callers):

```powershell
$PSCmdlet.ThrowTerminatingError($errorRecord)
```

## Catching: try / catch / finally

```powershell
try {
    Get-Content $path -ErrorAction Stop      # Stop makes a cmdlet error catchable
    1 / $divisor
} catch [System.IO.FileNotFoundException] {
    Write-Warning "Missing: $path"
} catch [System.DivideByZeroException], [System.ArgumentException] {
    Write-Error -ErrorRecord $_              # rich rethrow preserving position
} catch {
    throw                                    # rethrow anything else
} finally {
    if ($conn.State -eq 'Open') { $conn.Close() }   # always runs
}
```

- `try` must pair with `catch` and/or `finally`.
- Inside `catch`, `$_` (`$PSItem`) is the `ErrorRecord`. `$_.Exception.Message`, `$_.Exception.GetType().Name`, `$_.ScriptStackTrace`, `$_.FullyQualifiedErrorId`, `$_.TargetObject`.
- Order catches **most-specific first**.
- Only the *outermost* exception type can be matched, except `MethodInvocationException` (from a .NET method call), where PowerShell also lets you catch the **inner** type, e.g. `catch [System.ArgumentOutOfRangeException]`.
- `finally` runs even on a thrown error — ideal for closing streams/connections.
- **Rethrow** with bare `throw` in `catch`. To add context, wrap the original as an inner exception: `[InvalidOperationException]::new($_.Exception.Message, $_.Exception)`.

## Catching: ErrorVariable

For non-terminating errors, capture without try/catch:

```powershell
Get-Thing -ErrorAction SilentlyContinue -ErrorVariable failures
if ($failures.Count) { ... }
Get-Thing -ErrorVariable +failures   # '+' appends instead of overwriting
```

This is the idiom for `Invoke-Command` against many hosts — let it run, collect per-host failures in `-ErrorVariable`.

## The throw gotcha and recommendations

Despite the docs, **bare `throw` is affected by `-ErrorAction`/`$ErrorActionPreference`**: under `SilentlyContinue` or `Ignore` it is silently swallowed and the script continues. `throw` is also *script-terminating* — it tears down all parent scopes, unlike `ThrowTerminatingError` (statement-terminating, stops only the function), which matches how cmdlet/.NET-method errors behave.

Recommendations for consistent behavior:

1. Use `[CmdletBinding()]` on functions/scripts.
2. Use `throw` **only inside a `try` block** (so `catch` always fires regardless of error action).
3. Use `$PSCmdlet.ThrowTerminatingError($_)` to actually stop a command — reliable, ignores `-ErrorAction`.
4. Prefer `$PSCmdlet.WriteError()` for non-terminating errors (sets `$?`).
5. Any caller that invokes a command which may fail should wrap it in `try` — don't rely on global preferences.

```powershell
function Invoke-Something {
    [CmdletBinding()]
    param ( )
    try {
        throw 'failed'                       # caught locally, error action can't skip it
    } catch {
        $PSCmdlet.ThrowTerminatingError($_)  # reliably surfaces a terminating error
    }
}
```

## Error records and Get-Error

`Get-Error` (7+) shows the latest error in full detail (exception chain, stack trace, category). `$Error[4] | Get-Error` inspects a specific one.

Build a bespoke `ErrorRecord` to communicate clearly and attach diagnostic context:

```powershell
using namespace System.Management.Automation
$record = [ErrorRecord]::new(
    [InvalidOperationException]::new('Division failed'),
    'InvalidDivision',                       # ErrorId (stable, for matching)
    [ErrorCategory]::InvalidOperation,       # category
    [PSCustomObject]@{ Numerator=$n; Denominator=$d }   # TargetObject for debugging
)
Write-Error -ErrorRecord $record             # or $PSCmdlet.WriteError($record)
```

Match on `FullyQualifiedErrorId` or exception status codes rather than message text, which may be localized.

## trap

Legacy (PS 1.0) terminating-error handler, parsed before run so its position doesn't matter. Largely superseded by `try/catch`. Occasionally used to log unhandled errors. `continue` inside `trap` resumes at the next statement **in the trap's own scope**.

```powershell
trap { Write-Warning "Unhandled: $_"; continue }
```
