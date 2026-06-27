# Remoting and Remote Management

Running commands on remote machines via WinRM or SSH, sessions, argument passing, CIM sessions, and pitfalls.

## Contents
- Enabling remoting
- Invoke-Command
- Enter-PSSession (interactive)
- Persistent sessions
- Passing data: ArgumentList vs using:
- Capturing per-host failures
- SSH-based remoting
- The double-hop problem
- CIM sessions
- Just Enough Administration

## Enabling remoting

Remoting is client–server. The **server** (the target receiving the request) runs the WinRM service and must be enabled; the **client** needs no special config beyond PowerShell.

```powershell
Enable-PSRemoting                          # admin; configures WinRM + firewall (Private/Domain)
Enable-PSRemoting -SkipNetworkProfileCheck # allow limited access on Public profiles
```

Run it in *both* Windows PowerShell and PowerShell 7 if you use both. If client and server are the same machine, PowerShell must run elevated (or use `-EnableNetworkAccess`).

## Invoke-Command

Runs a script block (non-interactively) on one or many hosts:

```powershell
Invoke-Command -ComputerName SERVER01 -ScriptBlock { Get-Process }
Invoke-Command -ComputerName web01, web02, db01 -ScriptBlock { Get-Service dnscache }
```

- With an array of names it runs **in parallel**, up to `-ThrottleLimit` (default 32).
- Output gains `PSComputerName`, `PSShowComputerName`, `RunspaceId` so you can attribute each result to its host (these appear only when the call is remote, even for `localhost`).
- The script block can be arbitrarily complex; build a `[PSCustomObject]` to return structured results.
- Everything the block calls must exist *on the remote host*. You can reuse a local function's body as the block via the function provider, but only if it has no local dependencies:
  ```powershell
  Invoke-Command ${function:Get-NetInfo} -ComputerName web01
  ```

## Enter-PSSession (interactive)

Opens an interactive remote console — for ad-hoc work, not scripts:

```powershell
Enter-PSSession -ComputerName SERVER01
# ... commands run on SERVER01 ...
Exit-PSSession
```

## Persistent sessions

`Invoke-Command`/`Enter-PSSession` with `-ComputerName` create and tear down a session each call. Create one explicitly to reuse it (faster, preserves remote state):

```powershell
$s = New-PSSession -ComputerName SERVER01 -Credential (Get-Credential)
Invoke-Command -Session $s -ScriptBlock { $x = 1 }
Invoke-Command -Session $s -ScriptBlock { $x }    # state persists: returns 1
Get-PSSession
Remove-PSSession $s
```

`Import-PSSession`/`Export-PSSession` proxy remote commands into the local session (implicit remoting). Disconnected sessions (`Disconnect-PSSession`/`Connect-PSSession`) survive a client disconnect. `Copy-Item -ToSession`/`-FromSession` transfers files over a session.

## Passing data: ArgumentList vs using:

**`ArgumentList`** binds positionally to a `param()` in the block — awkward for optional/switch params:

```powershell
Invoke-Command -ComputerName SERVER01 -ArgumentList 'C', 'GB' -ScriptBlock {
    param ($Name, $Units)
    Get-PSDrive $Name | ForEach-Object { [math]::Round($_.Free / "1$Units", 2) }
}
```

**`using:` scope modifier** (preferred — clearer, handles any variable) reads client-side variables inside the block:

```powershell
$name = 'C'; $units = 'GB'
Invoke-Command -ComputerName SERVER01 -ScriptBlock {
    Get-PSDrive $using:name | ForEach-Object { [math]::Round($_.Free / "1$using:units", 2) }
}
```

The variables must exist before the call. `using:` also works with `Start-Job`, `Start-ThreadJob`, and `ForEach-Object -Parallel`. Script blocks are serialized to the remote host, so closures and live object references don't survive — only the data does.

## Capturing per-host failures

When targeting many hosts, don't pre-test with ping (`Test-Connection` tests ICMP, not WinRM, and isn't parallel). Let `Invoke-Command` run and collect failures via `-ErrorVariable` (not `try`/`-ErrorAction Stop`, which would abort on the first failure):

```powershell
$ok = Invoke-Command -ComputerName $hosts -ScriptBlock { Get-Service dnscache } `
    -ErrorAction SilentlyContinue -ErrorVariable failed
$failed | Select-Object @{ n='ComputerName'; e='TargetObject' },
                        @{ n='Error'; e={ $_.ToString() } }
```

## SSH-based remoting

PowerShell 7 can remote over SSH — no certificates, works to/from Linux. Not available in Windows PowerShell.

```powershell
Enter-PSSession -HostName linux01 -UserName me -SSHTransport -KeyFilePath ~/.ssh/id_rsa
Invoke-Command -HostName linux01 -UserName me -SSHTransport -ScriptBlock { uname -a }
New-PSSession -HostName linux01 -UserName me -SSHTransport
```

Server setup: install OpenSSH, and add a `Subsystem powershell /usr/bin/pwsh -sshs -NoLogo -NoProfile` line to `sshd_config` (path to `pwsh` varies), then restart sshd. Key auth uses the standard `~/.ssh/authorized_keys` flow. Interactive full-screen apps (vi, etc.) don't work over PS remoting because traffic is serialized.

## The double-hop problem

When you remote to host A and A tries to reach host B (a file share, AD, SQL), the second hop fails — the user's credentials can't be delegated implicitly (Kerberos). Symptom: "Access denied" on the remote resource only.

```powershell
Invoke-Command -ComputerName WEB01 { Get-Content \\FS01\share\file.txt }  # 2nd hop fails
```

Solutions, best to worst:
- **Kerberos constrained delegation** (resource-based) — the secure, recommended fix; configured in AD, outside the script.
- **Passing explicit credentials** into the block — simple, works, but exposes the credential on the remote host:
  ```powershell
  $cred = Get-Credential
  Invoke-Command -ComputerName FS01 { Get-ADUser -Filter * -Credential $using:cred }
  ```
- **CredSSP** — sends credentials in clear text; **not secure**, avoid for routine automation. Requires `Enable-WSManCredSSP -Role Client -DelegateComputer X` (client) and `-Role Server` (server). Disable afterward.

## CIM sessions

CIM/WMI work (and modules like `NetAdapter`, `Storage`) use CIM sessions, which also ride WinRM. Windows only.

```powershell
$cim = New-CimSession -ComputerName SERVER01        # WSMAN protocol
$cim = New-CimSession -SessionOption (New-CimSessionOption -Protocol DCOM)  # DCOM fallback
Get-Disk -CimSession $cim
Get-CimInstance Win32_OperatingSystem -CimSession $cim
Get-CimSession; Remove-CimSession $cim
```

Reuse one session across many CIM calls. Use DCOM when WinRM isn't configured on the target. Find CIM-capable commands with `Get-Command -ParameterName CimSession`.

## Just Enough Administration (JEA)

JEA constrains a remoting endpoint to a fixed set of commands/parameters run under a virtual account, so delegated operators get least-privilege access. Defined by a session configuration (`.pssc`, registered with `Register-PSSessionConfiguration`) plus role-capability files (`.psrc`) listing visible cmdlets/functions. Use it to expose a narrow, audited management surface instead of full admin.
