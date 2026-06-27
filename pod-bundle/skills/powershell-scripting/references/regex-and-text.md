# Regular Expressions and Text

PowerShell's regex operators, `$matches`, capture groups, the `[regex]` class, and parsing recipes.

## Contents
- The four regex operators
- $matches and capture groups
- Named groups
- -replace
- -split
- switch -Regex
- The [regex] type
- Long patterns and comments
- Common recipes

PowerShell uses .NET regular expressions throughout. Regex operators are **case-insensitive by default**; prefix with `c` for case-sensitive (`-cmatch`, `-creplace`, `-csplit`).

## The four regex operators

```powershell
'The cow jumped'  -match    'cow'        # True; populates $matches
'The cow jumped'  -notmatch 'pig'        # True
'abababab'        -replace  'a', 'c'     # 'cbcbcbcb'
'a1b2c3'          -split    '[0-9]'      # @('a','b','c','')
```

`-match`/`-notmatch` test and return a boolean for a scalar. **On an array they return matching elements** (like other comparison operators) and do **not** set `$matches`:

```powershell
'one','two','three' -match 'e'           # @('one','three')
```

## $matches and capture groups

After a successful scalar `-match`, `$matches` is a hashtable. `$matches[0]` is the whole match; numbered keys hold positional capture groups (parentheses):

```powershell
'Group one, Group two' -match 'Group (.*), Group (.*)'
$matches[0]   # 'Group one, Group two'
$matches[1]   # 'one'
$matches[2]   # 'two'
```

## Named groups

`(?<name>...)` captures into a named key — far more readable than numbers:

```powershell
if ('2026-06-13' -match '(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})') {
    "$($matches.year) / $($matches.month) / $($matches.day)"
}
```

`(?:...)` is a non-capturing group (group for alternation/quantifying without filling `$matches`).

## -replace

```powershell
'abababab' -replace 'a'                          # 'bbbb' (no replacement = remove)
'value1,value2,value3' -replace '(.*),(.*),(.*)', '$3,$2,$1'   # reverse via groups
```

- Form: `<input> -replace <pattern>, <replacement>`.
- In the replacement, `$1`/`$2` (or `${name}`) reference capture groups; `$$` is a literal `$`. **Use single quotes** for the replacement so PowerShell doesn't expand `$1` as a variable.
- A **script block** replacement (computed per match) — `$_` is a `[System.Text.RegularExpressions.Match]`:
  ```powershell
  'Process: 0', "Process: $PID" -replace '\d+', { (Get-Process -Id $_.Value).Name }
  ```
- Differs from the `.Replace()` string method: `-replace` is regex-based and case-insensitive; `.Replace('a','b')` is literal and case-sensitive.

## -split

```powershell
'1,2,3,4,5' -split ','                  # all separators
'1,2,3,4,5' -split ',', 2               # max 2 parts: '1', '2,3,4,5'
'1,2,3,4,5' -split ',', -2              # split from the right: '1,2,3,4', '5'
-split "a`tb   c"                        # unary: split on runs of whitespace
'a?b?c' -split 'b?', 0, 'SimpleMatch'   # literal (non-regex) split
```

Form: `<input> -split <pattern>, <max>, <options>`. Options include `SimpleMatch`, `RegexMatch`, `IgnoreCase`, `Multiline`, `Singleline`, `IgnorePatternWhitespace`, `ExplicitCapture`. Combine `-split` with destructuring: `$user, $domain = $upn -split '@'`.

## switch -Regex

Match a value (or each line of a `-File`) against multiple patterns; `$matches` is set per case:

```powershell
switch -Regex ('cat') {
    '^c'      { 'starts with c' }
    '^.{3}$'  { 'three chars' }
    't$'      { 'ends with t' }       # all three run (switch runs every match)
}
switch -Regex -File .\log.txt {
    'ERROR (?<code>\d+)' { "error $($matches.code)" }
}
```

## The [regex] type

For reuse, options, or richer APIs, use the .NET class directly:

```powershell
$rx = [regex]::new('(?<key>\w+)=(?<val>\w+)', 'IgnoreCase')
$rx.Matches('a=1 b=2') | ForEach-Object { $_.Groups['key'].Value }
[regex]::Match('a=1', '(\w+)=(\w+)').Groups[2].Value      # '1'
[regex]::Matches($text, $pattern)                          # all matches
[regex]::Escape('C:\temp\*.txt')                           # escape regex metachars
[regex]::Replace($text, $pattern, { param($m) $m.Value.ToUpper() })
```

`[regex]::Escape()` is the way to safely embed user input into a pattern. (For wildcard, not regex, patterns use `[System.Management.Automation.WildcardPattern]::Escape()`.)

## Long patterns and comments

Use the `(?x)` inline modifier (free-spacing) to ignore literal whitespace and allow `#` comments — split a gnarly pattern across lines. Or assemble with `-join @( ... )`:

```powershell
$pattern = -join @(
    '^(?:(?:\+|00)\d{2})?[ -]*'   # country code
    '(?:\(?0\)?[ -]*)?'           # trunk prefix
    '([138]\d{1,3}|20)[ -]*'      # area code
    '(\d{3,4})[ -]*(\d{3,4})$'    # subscriber number
)
$numbers -replace $pattern, '+44 $1 $2 $3'
```

## Common recipes

```powershell
# Extract all matches of a pattern from text
[regex]::Matches($text, '\b\d{1,3}(?:\.\d{1,3}){3}\b').Value     # IPv4-ish

# Validate format (anchored)
$code -match '^[A-Z]{2}\d{4}$'

# Parse key=value lines into a hashtable
$h = @{}
Get-Content config.ini | ForEach-Object {
    if ($_ -match '^\s*(?<k>[^=]+?)\s*=\s*(?<v>.*)$') { $h[$matches.k] = $matches.v }
}

# Trim/normalize whitespace
$s -replace '\s+', ' '

# Strip ANSI/control or non-printable chars
$s -replace '[^\x20-\x7E]'
```

Prefer `-match`/`-replace` with anchors (`^`, `$`) and explicit quantifiers; test patterns against representative inputs. For fixed-string find/replace, the literal `.Replace()` method or `-replace [regex]::Escape($literal)` avoids accidental metacharacter interpretation.
