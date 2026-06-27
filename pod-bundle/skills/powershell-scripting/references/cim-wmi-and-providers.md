# CIM/WMI, Files, Folders, and the Registry

System inventory via CIM/WMI, WQL, and the provider model (filesystem, registry, content, ACLs). Mostly Windows-centric.

## Contents
- CIM vs WMI
- Querying instances
- WQL
- Classes and methods
- Creating and removing instances
- The provider model
- Navigating and items
- Reading and writing content
- File encoding
- The registry
- Permissions (ACLs)

## CIM vs WMI

WMI is Windows' management database. **Use the CIM cmdlets** (`Get-CimInstance`, etc., from `CimCmdlets`), not the old `Get-WmiObject`:

- CIM cmdlets: in both Windows PowerShell and 7+ (Windows only — CIM is not on Linux/macOS), auto-convert dates, use WSMan by default (configurable to DCOM), DMTF-standard-compliant.
- `Get-WmiObject` / `*-WmiObject`: Windows PowerShell only (removed from 7+, partly usable via the compat session but methods are crippled), DCOM-only. Deprecated.

A WMI class (e.g. `Win32_Process`) defines properties and methods; instances are the live objects. Classes live in namespaces, default `root\cimv2`.

## Querying instances

```powershell
Get-CimInstance -ClassName Win32_OperatingSystem
Get-CimInstance Win32_Service -Filter "State='Running'"
Get-CimInstance Win32_UserAccount -Property Name, SID            # limit returned fields
Get-CimInstance -Query "SELECT Name, SID FROM Win32_UserAccount" # WQL
Get-CimInstance Win32_Process -ComputerName SERVER01            # or -CimSession $cim
```

`-Filter` and `-Query` take WQL. `-Property` reduces network/parse cost for wide classes.

## WQL

A SQL subset. Keywords (case-insensitive, uppercase by convention): `SELECT ... FROM <class> WHERE <condition>`.

```powershell
"SELECT * FROM Win32_Process WHERE ProcessID=$PID"
"SELECT * FROM Win32_Process WHERE Name LIKE 'pwsh%'"     # % and _ are WQL wildcards
"SELECT * FROM Win32_Directory WHERE Name='C:\\Windows'"  # backslash escapes in -Filter
```

`\` escapes characters (quotes, wildcards, itself). WQL `LIKE` uses `%` (any string) and `_` (one char). Combine conditions with `AND`/`OR`. WQL cannot filter array-valued properties; CIM dates use the `dmtf` format but CIM cmdlets present them as `[datetime]`.

## Classes and methods

```powershell
Get-CimClass Win32_Process
(Get-CimClass Win32_Process).CimClassMethods                              # Create, Terminate, ...
(Get-CimClass Win32_Process).CimClassMethods['Create'].Parameters         # In/Out args
Get-CimInstance __Namespace -Namespace root                               # child namespaces
```

Invoke a static (class) method or an instance method with `Invoke-CimMethod`; arguments go in a hashtable:

```powershell
Invoke-CimMethod -ClassName Win32_Process -MethodName Create `
    -Arguments @{ CommandLine = 'notepad.exe' }                            # returns ProcessId, ReturnValue

Get-CimInstance Win32_Process -Filter "Name='notepad.exe'" |
    Invoke-CimMethod -MethodName Terminate                                 # instance method
```

`ReturnValue` of `0` means success; other codes are documented per class (PowerShell doesn't translate them). `In` params are inputs; `Out` params come back on the result object.

## Creating and removing instances

```powershell
New-CimInstance -ClassName Win32_Environment -Property @{ Name='X'; VariableValue='1'; UserName='<SYSTEM>' }
Get-CimInstance Win32_Share -Filter "Name='temp'" | Remove-CimInstance
```

Get associated objects with `Get-CimAssociatedInstance` (e.g. the session that owns a process, the partitions on a disk).

## The provider model

Providers expose data as drive-like hierarchies sharing one command set. Built-in everywhere: `Alias:`, `Env:`, `Function:`, `Variable:`, `FileSystem` (`C:`, `/`). Windows adds `Registry` (`HKLM:`, `HKCU:`), `Certificate` (`Cert:`), `WSMan:`. Modules can add more.

```powershell
Get-PSProvider                  # available providers
Get-PSDrive                     # drives across all providers
```

Which commands a provider supports depends on the interfaces it implements: navigation (`Get-Item`, `Get-ChildItem`, `Set-Location`), content (`Get-/Set-/Add-/Clear-Content`), properties (`*-ItemProperty`), security (`Get-/Set-Acl`).

## Navigating and items

```powershell
Set-Location HKLM:\Software             # cd into any provider container
Push-Location C:\Windows; Pop-Location  # location stack
$PWD                                    # current location (any provider)

Get-Item C:\Windows                     # one item
Get-ChildItem C:\Windows -Recurse -File -Filter *.dll
New-Item .\logs -ItemType Directory
New-Item .\f.txt -ItemType File -Force  # Force overwrites/creates parents
Copy-Item / Move-Item / Rename-Item / Remove-Item
Test-Path C:\Windows                    # exists?
Test-Path C:\Windows -PathType Container # is it a folder?
Join-Path $root child                   # build cross-platform paths
```

Use `Join-Path`, `$env:`, and `[System.IO.Path]::PathSeparator` rather than hard-coded separators. `~` is the home directory. `-LiteralPath` skips wildcard interpretation.

## Reading and writing content

```powershell
Get-Content file.txt                    # array of lines (with PS* note properties)
Get-Content file.txt -Raw               # whole file as one string (much faster)
[System.IO.File]::ReadAllLines("$pwd\file.txt")   # fastest; needs a full path
Set-Content file.txt -Value 'overwrite'
Add-Content log.txt -Value 'append'
```

**Pipeline self-overwrite trap:** `Get-Content f | ... | Set-Content f` fails (file still open for reading). Force a full read first with parentheses: `(Get-Content f) | ... | Set-Content f`.

## File encoding

`-Encoding` matters. PowerShell 7 defaults to **UTF-8 without BOM**; Windows PowerShell 5.1 defaults vary (often UTF-16LE for `Out-File`/`Set-Content`, ASCII for `Add-Content`). Always specify when output must be consumed elsewhere: `Set-Content f.txt -Encoding utf8`. Values: `utf8`, `utf8BOM`, `utf8NoBOM`, `unicode` (UTF-16LE), `ascii`, etc.

## The registry (Windows)

Keys are containers (items); values are item *properties*:

```powershell
Get-ChildItem HKLM:\Software             # subkeys
Get-ItemProperty HKCU:\Environment       # all values under a key
Get-ItemPropertyValue HKCU:\Environment -Name Path
Set-ItemProperty HKCU:\Environment -Name MyVar -Value 'x'   # creates REG_SZ if new
New-ItemProperty HKCU:\Environment -Name Expand -Value '%USERNAME%' -PropertyType ExpandString
Remove-ItemProperty HKCU:\Environment -Name MyVar
New-Item HKCU:\Software\MyApp            # new key
Test-Path HKLM:\Software\Microsoft
```

`Set-ItemProperty` infers the type for new values (Int32→REG_DWORD, String→REG_SZ, String[]→REG_MULTI_SZ, Byte[]→REG_BINARY); use `New-ItemProperty -PropertyType` for an explicit type (e.g. `ExpandString`). 32-bit values live under `Wow6432Node` from a 64-bit process. A common idiom: enumerate the `Uninstall` keys to detect installed software.

## Permissions (ACLs, Windows)

`Get-Acl`/`Set-Acl` work on the FileSystem, Registry, and Certificate providers. The security descriptor holds a **DACL** (access — `.Access` in PowerShell) and a **SACL** (audit — `.Audit`, needs admin and `-Audit`).

```powershell
$acl = Get-Acl C:\Temp\data
$acl.Access                                  # the ACEs
$rule = [System.Security.AccessControl.FileSystemAccessRule]::new(
    'DOMAIN\User', 'Modify', 'ContainerInherit,ObjectInherit', 'None', 'Allow')
$acl.AddAccessRule($rule)
$acl.SetAccessRuleProtection($true, $false)  # block inheritance, don't copy parent ACEs
Set-Acl C:\Temp\data -AclObject $acl
```

ACL type is `DirectorySecurity` (folder), `FileSecurity` (file), or `RegistrySecurity` (key). ACEs come from inheritance or are explicit; `SetAccessRuleProtection` controls inheritance. For less ceremony, the community `NtfsSecurity` module wraps these .NET calls.
