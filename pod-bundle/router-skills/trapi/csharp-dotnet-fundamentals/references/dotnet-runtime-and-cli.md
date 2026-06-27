# .NET Runtime, SDK, and the dotnet CLI

Tooling, project structure, target frameworks, NuGet, configuration, and publishing for .NET 8.

## Contents
- Runtime, SDK, and versions
- The dotnet CLI
- Project and solution structure
- The project file (.csproj)
- Target frameworks
- Assemblies, namespaces, and NuGet
- Referencing and creating packages
- Configuration (appsettings.json)
- Publishing, AOT, and trimming
- Logging during development

## Runtime, SDK, and versions

- **Runtime** (CoreCLR + BCL): the minimum to *run* a .NET app. **SDK**: runtime + compilers (Roslyn) + tools to *build*. Install the SDK to develop.
- The C# compiler turns source into IL in an assembly; CoreCLR JIT-compiles IL to native at runtime — the same IL runs on Windows, macOS, and Linux.
- Versions are **LTS** (3 years' support; .NET 8 → Nov 2026) or **STS** (18 months). .NET 8 uses **C# 12** by default. You can install multiple SDKs side by side; the newest is used unless pinned.
- Inspect your install:
```bash
dotnet --version          # active SDK version
dotnet --list-sdks
dotnet --list-runtimes
dotnet --info             # SDK + runtimes + OS + RID
dotnet sdk check          # which installs have updates
```
You can target an older runtime while using a newer compiler: set `<TargetFramework>net8.0</TargetFramework>` with `<LangVersion>13</LangVersion>` once a newer SDK is installed. Pin the SDK for a folder with a `global.json` (`dotnet new globaljson --sdk-version 8.0.100`).

## The dotnet CLI

```bash
dotnet new console -o MyApp           # create projects from templates
dotnet new classlib -o MyLib
dotnet new xunit -o MyTests
dotnet new sln -n MySolution          # create a solution file
dotnet new list                       # show installed templates

dotnet sln add MyApp                  # add a project to the solution
dotnet add MyApp reference ../MyLib   # project-to-project reference
dotnet add MyApp package Newtonsoft.Json   # add a NuGet package

dotnet restore                        # download dependencies (build does this too)
dotnet build                          # compile (add --tl for terminal-logger output)
dotnet run                            # build + run (from the project folder)
dotnet watch                          # run with Hot Reload on file save
dotnet test                           # build + run unit tests
dotnet pack                           # build a NuGet package
dotnet publish                        # build a deployable app (Release by default in .NET 8)
dotnet clean                          # remove build outputs (bin/obj)
```
Common template short names: `console`, `classlib`, `xunit`, `nunit`, `mstest`, `sln`, `web`, `webapi`, `mvc`, `blazor`, `globaljson`, `editorconfig`.

## Project and solution structure

A **solution** (`.sln`) groups related **projects** (`.csproj`). A typical layout:
```
MySolution/
  MySolution.sln
  MyApp/        MyApp.csproj        Program.cs
  MyLib/        MyLib.csproj        *.cs
  MyTests/      MyTests.csproj      *Tests.cs
```
Each build produces `obj/` (intermediate) and `bin/` (output) folders — disposable; deleting them just forces a rebuild. The dotnet CLI runs an app from the project folder; Visual Studio runs from `bin/Debug/net8.0` — relevant when an app reads relative file paths.

## The project file (.csproj)

SDK-style, declarative, minimal:
```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>        <!-- omit for a class library -->
    <TargetFramework>net8.0</TargetFramework>
    <Nullable>enable</Nullable>          <!-- NRT analysis on -->
    <ImplicitUsings>enable</ImplicitUsings>
    <LangVersion>12</LangVersion>        <!-- optional; default matches the SDK -->
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.3" />
    <ProjectReference Include="..\MyLib\MyLib.csproj" />
    <Using Include="System.Console" Static="true" />  <!-- call WriteLine() unqualified -->
    <Using Remove="System.Threading" />               <!-- drop an implicit import -->
  </ItemGroup>
</Project>
```
`<ImplicitUsings>` auto-imports `System`, `System.Collections.Generic`, `System.IO`, `System.Linq`, `System.Net.Http`, `System.Threading`, `System.Threading.Tasks` (the Web SDK adds ASP.NET namespaces). Tune the set with `<Using Include/Remove/Static/Alias>` items.

## Target frameworks

| TFM | Use for |
|---|---|
| `net8.0` | New apps and libraries (current LTS) |
| `net6.0`, `net7.0` | Older runtimes still in support |
| `netstandard2.0` | Libraries shared with legacy .NET Framework / Xamarin (and source generators) |
| `netstandard2.1` | Libraries shared with .NET Core 3+ / Xamarin (not .NET Framework) |

Pick the lowest framework that has the APIs you need to maximize who can reference your library. A `netstandard2.0` library defaults to the C# 7.3 compiler — override with `<LangVersion>12</LangVersion>` to use modern syntax (some features still require a newer runtime, e.g. `required` needs .NET 7+).

## Assemblies, namespaces, and NuGet

- An **assembly** is the deployment unit on disk — a `.dll` (library) or `.exe` (app) containing IL. Reference an assembly to use its types; no circular references allowed.
- A **namespace** is a type's logical address (`System.Xml.Linq.XDocument`). Import with `using System.Xml.Linq;` so you can write `XDocument`. C# keywords (`string`, `int`) are aliases for BCL types (`System.String`, `System.Int32`) and need no import. Use a file-scoped namespace: `namespace MyApp.Domain;`.
- **NuGet** packages bundle assemblies; `nuget.org` is the public feed. The SDK references the `Microsoft.NET.Sdk` platform implicitly. Aliases: `using Env = System.Environment;` (rename a type), `using Tx = Texas;` (disambiguate clashing namespaces).

## Referencing and creating packages

```bash
dotnet add package Newtonsoft.Json --version 13.0.3   # pin a fixed version
```
**Pin dependencies** to a fixed version matching your target framework; avoid wildcards (`13.0.*`), `beta`, or `rc` in production — they can pull breaking changes. To produce a package, set metadata in the `.csproj` and `dotnet pack`:
```xml
<PropertyGroup>
  <GeneratePackageOnBuild>true</GeneratePackageOnBuild>
  <PackageId>Acme.StringExtensions</PackageId>   <!-- must be globally unique -->
  <Version>1.0.0</Version>
  <Authors>Your Name</Authors>
  <PackageLicenseExpression>MIT</PackageLicenseExpression>   <!-- an SPDX id -->
</PropertyGroup>
```

## Configuration (appsettings.json)

```json
{ "Logging": { "Level": "Info" }, "ConnectionStrings": { "Default": "..." } }
```
```csharp
using Microsoft.Extensions.Configuration; // add the package for non-web apps

IConfigurationRoot config = new ConfigurationBuilder()
    .SetBasePath(Directory.GetCurrentDirectory())
    .AddJsonFile("appsettings.json", optional: false, reloadOnChange: true)
    .Build();

string? level = config["Logging:Level"];
config.GetSection("Logging").Bind(myOptions);   // bind a section to an options object
```
Mark `appsettings.json` to copy to output: `<None Update="appsettings.json"><CopyToOutputDirectory>Always</CopyToOutputDirectory></None>`, because the build runs from `bin/...`.

## Publishing, AOT, and trimming

```bash
# Framework-dependent (needs .NET installed on the target):
dotnet publish -c Release

# Self-contained single file for one OS/CPU (RID):
dotnet publish -c Release -r win-x64 --self-contained /p:PublishSingleFile=true
dotnet publish -c Release -r linux-arm64 --self-contained
```
Common RIDs: `win-x64`, `win-arm64`, `linux-x64`, `linux-arm64`, `osx-x64`, `osx-arm64`. Declare them in the project with `<RuntimeIdentifiers>win-x64;linux-x64;osx-arm64</RuntimeIdentifiers>`.

Reduce self-contained size:
- **Trimming** removes unused assemblies/types/members: `<PublishTrimmed>true</PublishTrimmed>` (full) or add `<TrimMode>partial</TrimMode>` (assembly-level). Beware reflection-heavy code, which trimming can break.
- **Native AOT** compiles to native code at publish time (faster startup, smaller memory, no JIT, no runtime needed): `<PublishAot>true</PublishAot>`. Limitations: no dynamic assembly loading, no `Reflection.Emit`, requires trimming. AOT only happens at publish — debugging still uses the JIT.

Control where build artifacts go solution-wide with a `Directory.Build.props` (`dotnet new buildprops --use-artifacts`).

## Logging during development

`System.Diagnostics.Debug` writes only in Debug builds; `Trace` writes in Debug and Release. Both write to configurable trace listeners (you can add a `TextWriterTraceListener` to log to a file). Capture caller context with `[CallerMemberName]`, `[CallerFilePath]`, `[CallerLineNumber]`, and `[CallerArgumentExpression]` on optional parameters. For production, a structured logging library (Serilog, NLog) or `Microsoft.Extensions.Logging` is preferred.
