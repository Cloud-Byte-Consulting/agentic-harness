# Language Fundamentals

Variables, types, collections, operators, and control flow. PowerShell 7+ with 5.1 notes.

## Contents
- Variables and scope
- Types and coercion
- Strings and expansion
- Arrays and collections
- Hashtables and ordered dictionaries
- Operators (comparison, logical, arithmetic, type, redirection)
- Conditionals and loops
- switch

## Variables and scope

Variables start with `$`. Assignment is `=`; multiple assignment works (`$a = $b = 0`, or destructuring `$first, $rest = 1, 2, 3` — extras land in the last variable, `$null` discards).

```powershell
$count = 5
$first, $null, $last = -split 'First A. Last'   # discard middle
```

Names with odd characters go in braces: `${My Variable}`. A colon makes it a *provider* path, not a name: `${env:ProgramFiles(x86)}`, `${function:mkdir}`.

**Scopes** layer parent→child. A child reads parent variables but assigning *creates a new local variable* shadowing the parent. Named scopes: `Global` (console top level), `Script` (a .ps1 or module), `Local` (current). Modifiers: `$global:x`, `$script:x`, `$private:x` (hidden from children), `$using:x` (read a parent-session variable inside `Invoke-Command`/`Start-Job`/`ForEach-Object -Parallel`). Avoid reaching across scopes for user variables — it makes code untestable. Prefer parameters and return values.

The `*-Variable` cmdlets (`Get-/Set-/New-/Remove-/Clear-Variable`) exist but are rarely needed; direct assignment is idiomatic and faster (the cmdlets disable some engine optimizations). `New-Variable -Option Constant`/`ReadOnly` is the exception.

## Types and coercion

Values are .NET objects. `$x.GetType().FullName` shows the type. PowerShell is dynamically typed but converts aggressively (a string `'1/1/2020'` becomes a `[datetime]`, `'4'` becomes an `[int]` when added to a number).

- **Type on the right** converts only that value: `$x = [string]1`.
- **Type on the left of a variable/param** pins the type for all future assignments: `[int]$n = '42'` — any later `$n = ...` is coerced to int (or errors).

```powershell
[datetime]$d = Get-Date
$d = '1/1/1970'        # coerced via [datetime]::Parse
```

Common type accelerators: `[int] [long] [double] [decimal] [string] [bool] [datetime] [timespan] [guid] [regex] [xml] [hashtable] [pscustomobject] [ordered] [scriptblock] [version] [ipaddress]`. Arrays use `[]`: `[string[]]`, `[int[]]`.

**Value vs reference types.** Integers/structs copy on assignment; hashtables/arrays/objects share a reference (`$b = $a; $b.Key = 'x'` mutates `$a` too). Strings are immutable, so they behave like value types.

Numeric multipliers: `1KB 1MB 1GB 1TB 1PB` (binary, 1024-based).

## Strings and expansion

- `'single'` — literal, no expansion.
- `"double"` — expands `$var` and subexpressions `$(...)`.

```powershell
"There are $count items"
"Length: $($word.Length)"          # subexpression for property/method/expr
"${ComputerName}: running"          # braces delimit the name before a literal ':'
```

A `:` right after a `$var` is read as a provider drive — use `${name}` or `$($name)`. Here-strings span lines: `@"..."@` (expanding) and `@'...'@` (literal); the closing token must be at column 0.

Format operator: `'Name: {0}, Status: {1}' -f $name, $status`.

## Arrays and collections

Multiple results from a command become a `[object[]]` array; a single result is a scalar. Force an array with `@(...)`:

```powershell
$procs = @(Get-Process -Id $PID)   # always an array, even with 0 or 1 result
```

Create: `$a = 1, 2, 3` or `@(1, 2, 3)` or multiline inside `@( )`. Index with `$a[0]`, `$a[-1]` (last), ranges `$a[2..4]`, `$a[-1..-3]`, combined `$a[@(0) + 6..8 + -1]`.

**Arrays are fixed-size.** `+=` recreates the whole array — O(n²) in a loop. Two better patterns:

```powershell
# Best: assign the loop output directly
$result = foreach ($i in 1..1000) { [PSCustomObject]@{ N = $i } }

# When you must append incrementally, use a List
$list = [System.Collections.Generic.List[object]]::new()
$list.Add($item)            # returns nothing; no suppression needed
$list.AddRange([string[]]$arr)
```

`[System.Collections.ArrayList]` is older; `.Add()` returns an index you must suppress (`$null = $list.Add(x)`). Filter arrays with comparison operators (`$a -gt 5` returns matches), `Where-Object`, or the faster `.Where{ ... }` / `.ForEach{ ... }` methods (these require an actual collection).

## Hashtables and ordered dictionaries

```powershell
$h = @{ Key1 = 'v1'; Key2 = 'v2' }
$h['Key3'] = 'v3'      # add/overwrite
$h.Key4 = 'v4'         # dot also works
$h.Add('K5', 'v5')     # errors if key exists
$h.ContainsKey('Key1') # test (Contains works on hashtables too)
$h.Remove('Key2')
```

Hashtable key order is **not guaranteed**. For predictable order use `[ordered]@{ ... }` — essential when building a `[PSCustomObject]` field by field. `[ordered]` and `[pscustomobject]` are parser instructions, not ordinary types.

**Hashtables as fast lookups / joins.** Membership lookup is O(1). To inner-join two large sets, index one side into a hashtable and test the other — orders of magnitude faster than `Where-Object -in`:

```powershell
$lookup = @{}
$right | ForEach-Object { $lookup[$_.UserID] = $_ }
$left | Where-Object { $lookup.ContainsKey($_.UserID) }
```

Hashtables are case-insensitive by default; `[System.Collections.Generic.HashSet[string]]` is not unless given `[StringComparer]::OrdinalIgnoreCase`.

Hashtables also drive splatting (`cmd @params`) and parameters for `Select-Object`/`Sort-Object` calculated properties.

## Operators

**Comparison** (case-insensitive by default; `-c` prefix = case-sensitive, `-i` = explicit insensitive). They return a boolean for scalars but **return matching elements when the left side is an array**:

| Operator | Meaning |
|---|---|
| `-eq` `-ne` | equal / not equal |
| `-gt` `-ge` `-lt` `-le` | greater/less (and-or-equal) |
| `-like` `-notlike` | wildcard (`*` `?` `[a-z]`) |
| `-match` `-notmatch` | regex (sets `$matches`) |
| `-contains` `-notcontains` | array (left) contains scalar (right) |
| `-in` `-notin` | scalar (left) in array (right) |

PowerShell coerces the right side to the left side's type. **Always put `$null` on the left:** `if ($null -eq $x)`. To test "no array element matches a wildcard," negate a positive test: `if (-not ($a -like 't*'))`.

**Logical:** `-and -or -xor -not` (`!`). Short-circuits, so order conditions to avoid errors (`(Test-Path $p) -and (Get-Item $p).Length`).

**Arithmetic:** `+ - * / %`. `+` concatenates strings and joins arrays/hashtables; `*` repeats strings/arrays (`'ab' * 3`). `++`/`--` increment/decrement.

**Ternary (7+):** `$cond ? $ifTrue : $ifFalse`. **Pipeline chain (7+):** `cmd1 && cmd2` (run on success), `cmd1 || cmd2` (run on failure). **Null-coalescing (7+):** `$x ?? 'default'`, assign `$x ??= 'default'`.

**Type:** `-is` `-isnot` `-as` (`'1' -as [int]`). **Format:** `-f`. **Join/split:** `-join`, `-split` (regex; unary `-split $s` splits on whitespace).

**Redirection:** `>` `>>` (to file), `2>` (errors), `*>` (all streams), `2>&1` (merge error into output), `> $null` (discard). Stream numbers: 1 output, 2 error, 3 warning, 4 verbose, 5 debug, 6 information.

## Conditionals and loops

```powershell
if ($x -eq 1) { ... } elseif ($x -lt 10) { ... } else { ... }
```

Implicit boolean: `$null`, `''`, `@()`, and `0` are false; a non-empty value/array is true — so `if (Get-ChildItem $p) { }` works. Assignment inside `if` returns the value (`if ($svc = Get-Service ... ) { }`), handy but easy to confuse with `-eq`.

Loops:

```powershell
foreach ($item in $collection) { $item }     # fastest; not a pipeline
for ($i = 0; $i -lt $a.Count; $i++) { }
while ($cond) { }
do { } while ($cond)      # or until ($cond)
1..10 | ForEach-Object { $_ }                 # pipeline (see pipeline-and-objects.md)
```

`break` exits a loop; `continue` skips to the next iteration. Labels target an outer loop: `break outer`. `foreach` (keyword) is faster than the `ForEach-Object` cmdlet and reads the whole collection into memory; `ForEach-Object` streams.

## switch

```powershell
switch ($value) {
    1       { 'one' }
    'two'   { 'matched two' }
    default { 'no match' }
}
```

By default `switch` is exact, case-insensitive, and **runs every matching case** (use `break` to stop, `continue` to move to the next array element). Options: `-Wildcard`, `-Regex`, `-CaseSensitive`, `-File <path>` (matches each line). Script-block cases test a condition: `{ $_ -is [datetime] } { ... }`. On an array, each element is tested in turn; an empty `@()` runs nothing, but explicit `$null` runs `default`.

**Gotcha:** cases are coerced to strings, so an enum case must be parenthesized: `([DayOfWeek]::Monday) { }` — or use the name `'Monday'`.
