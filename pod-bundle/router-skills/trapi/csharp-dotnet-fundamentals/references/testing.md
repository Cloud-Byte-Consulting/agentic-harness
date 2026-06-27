# Unit Testing in .NET

Writing and running unit tests with xUnit (the common default) and NUnit, plus structure, assertions, parameterized tests, fixtures, and mocking.

## Contents
- Setup
- Anatomy of a test (Arrange–Act–Assert)
- xUnit
- NUnit
- Assertions
- Parameterized tests
- Setup/teardown and fixtures
- Testing exceptions and async
- Mocking dependencies
- Running tests
- Practices

## Setup

Keep tests in their own project that references the code under test:
```bash
dotnet new classlib -o CalculatorLib && dotnet sln add CalculatorLib
dotnet new xunit -o CalculatorLib.Tests && dotnet sln add CalculatorLib.Tests
dotnet add CalculatorLib.Tests reference CalculatorLib
```
(Use `dotnet new nunit` for NUnit, or `dotnet new mstest` for MSTest.) The test project references the test framework and `Microsoft.NET.Test.Sdk` plus a runner/adapter, all wired up by the template.

Unit testing exercises the smallest unit (a method) in isolation, mocking its dependencies. It catches bugs early and cheaply, and supports refactoring and TDD. Other levels: integration, system, performance, load, user-acceptance.

## Anatomy of a test (Arrange–Act–Assert)

```csharp
[Fact]
public void Add_TwoAndTwo_ReturnsFour()
{
    // Arrange
    Calculator calc = new();
    double a = 2, b = 2, expected = 4;

    // Act
    double actual = calc.Add(a, b);

    // Assert
    Assert.Equal(expected, actual);
}
```
Write several tests per unit: typical inputs, boundary/extreme values, and deliberately invalid inputs (to verify error handling). Name tests so the intent is clear (`Method_Scenario_ExpectedResult`).

## xUnit

```csharp
using Xunit;
using CalculatorLib;

namespace CalculatorLib.Tests;

public class CalculatorTests
{
    [Fact]                                   // a single, parameterless test
    public void Subtract_FiveMinusThree_ReturnsTwo()
    {
        var calc = new Calculator();
        Assert.Equal(2, calc.Subtract(5, 3));
    }
}
```
- `[Fact]` marks a test method. `[Theory]` + data attributes parameterize it.
- The test class is instantiated **fresh for every test method**, so use the constructor for per-test setup and implement `IDisposable.Dispose()` for per-test teardown (no `[SetUp]`/`[TearDown]` attributes).

## NUnit

```csharp
using NUnit.Framework;

[TestFixture]
public class CalculatorTests
{
    private Calculator _calc = null!;

    [SetUp]    public void Setup()    => _calc = new Calculator();   // before each test
    [TearDown] public void Teardown() { /* clean up after each test */ }

    [Test]
    public void Add_TwoAndTwo_ReturnsFour()
        => Assert.That(_calc.Add(2, 2), Is.EqualTo(4));
}
```
NUnit uses `[Test]`, `[TestFixture]`, `[SetUp]`/`[TearDown]` (per test) and `[OneTimeSetUp]`/`[OneTimeTearDown]` (per fixture), and the constraint-based `Assert.That(actual, Is.EqualTo(expected))`. xUnit and NUnit do essentially the same job; xUnit (built by NUnit's original authors) is the common modern default and integrates cleanly with `dotnet test`.

## Assertions

xUnit:
```csharp
Assert.Equal(expected, actual);
Assert.NotEqual(a, b);
Assert.True(condition);  Assert.False(condition);
Assert.Null(x);  Assert.NotNull(x);
Assert.Contains(item, collection);
Assert.Empty(collection);
Assert.Equal(3.14159, value, precision: 2);     // floating-point tolerance
Assert.IsType<Employee>(person);
Assert.Throws<ArgumentException>(() => calc.Divide(1, 0));
```
NUnit constraint model: `Assert.That(actual, Is.EqualTo(expected))`, `Is.True`, `Is.Null`, `Does.Contain(...)`, `Is.EqualTo(3.14).Within(0.01)`, `Throws.TypeOf<ArgumentException>()`.

## Parameterized tests

xUnit `[Theory]`:
```csharp
[Theory]
[InlineData(2, 3, 5)]
[InlineData(-1, 1, 0)]
[InlineData(0, 0, 0)]
public void Add_VariousInputs(int a, int b, int expected)
    => Assert.Equal(expected, new Calculator().Add(a, b));
```
Use `[MemberData]` / `[ClassData]` for richer/computed data sets. NUnit equivalents: `[TestCase(2, 3, 5)]` and `[TestCaseSource]`.

## Setup/teardown and fixtures

- **xUnit**: constructor = setup, `Dispose` = teardown, both per test. Share expensive state across tests in a class with a *class fixture* (`IClassFixture<TFixture>`), or across classes with a *collection fixture*.
- **NUnit**: `[SetUp]`/`[TearDown]` run per test; `[OneTimeSetUp]`/`[OneTimeTearDown]` run once per fixture.

## Testing exceptions and async

```csharp
// Exception (xUnit):
var ex = Assert.Throws<ArgumentOutOfRangeException>(() => account.Deposit(-5));
Assert.Equal("amount", ex.ParamName);

// Async: make the test method async Task and await
[Fact]
public async Task FetchAsync_ReturnsContent()
{
    var result = await service.FetchAsync(url);
    Assert.NotEmpty(result);
}

// Async-throwing:
await Assert.ThrowsAsync<HttpRequestException>(() => service.FetchAsync(badUrl));
```

## Mocking dependencies

To isolate the unit under test, replace its collaborators with test doubles. Define dependencies behind interfaces, then inject a fake. Hand-written fake:
```csharp
public interface IClock { DateTime UtcNow { get; } }
public sealed class FixedClock(DateTime now) : IClock { public DateTime UtcNow => now; }

[Fact]
public void Expiry_UsesInjectedClock()
{
    var sut = new TokenService(new FixedClock(new DateTime(2026, 1, 1)));
    Assert.False(sut.IsExpired(token));
}
```
Or a mocking library (Moq, NSubstitute) for behavior verification:
```csharp
var repo = new Mock<IUserRepository>();
repo.Setup(r => r.FindById(1)).Returns(new User("Ada"));
var sut = new UserService(repo.Object);
Assert.Equal("Ada", sut.GetName(1));
repo.Verify(r => r.FindById(1), Times.Once);
```

## Running tests

```bash
dotnet test                                   # build + run all tests in the solution/project
dotnet test --filter "FullyQualifiedName~Calculator"   # subset by name
dotnet test --collect:"XPlat Code Coverage"   # collect coverage
```
Visual Studio (Test Explorer) and VS Code (C# Dev Kit Testing view) discover and run tests with a UI and per-test debugging. Build the test project before tests appear.

## Practices

- One logical assertion per test where practical; clear, behavior-describing names.
- Test the public contract, not private internals.
- Cover happy path, boundaries, and invalid input; assert on exception type and key details.
- Keep tests fast, deterministic, and independent (no shared mutable state, no ordering dependencies). Inject time, randomness, and I/O so they're controllable.
