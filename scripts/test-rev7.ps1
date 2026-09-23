param([string]$SecretsPath, [switch]$Regression, [switch]$Race)
$ErrorActionPreference = 'Stop'
$rev7Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev7Root '.local/secrets/rev5-databases.clixml' }
$rev7Keys = @('CRAFT_MYSQL_DSN', 'CRAFT_POSTGRES_DSN', 'CRAFT_SQLSERVER_DSN')
$rev7Previous = @{}
foreach ($rev7Key in $rev7Keys) { $rev7Previous[$rev7Key] = [Environment]::GetEnvironmentVariable($rev7Key, 'Process') }
try {
    $rev7Secrets = Import-Clixml -LiteralPath $SecretsPath
    foreach ($rev7Key in $rev7Keys) {
        if ($rev7Secrets[$rev7Key] -isnot [Security.SecureString]) { throw "Missing encrypted value: $rev7Key" }
        $rev7Value = [Net.NetworkCredential]::new('', $rev7Secrets[$rev7Key]).Password
        if ([string]::IsNullOrWhiteSpace($rev7Value)) { throw "Empty value: $rev7Key" }
        [Environment]::SetEnvironmentVariable($rev7Key, $rev7Value, 'Process')
    }
    Push-Location $rev7Root
    try {
        $rev7Args = @('test', './tests', './internal/stdlib', '-run', '^TestRev7', '-count=1', '-v', '-timeout', '300s')
        if ($Regression) { $rev7Args = @('test', './...', '-count=1', '-v', '-timeout', '600s') }
        if ($Race) { $rev7Args += '-race' }
        & go @rev7Args
        $rev7Exit = $LASTEXITCODE
    } finally { Pop-Location }
} finally {
    foreach ($rev7Key in $rev7Keys) { [Environment]::SetEnvironmentVariable($rev7Key, $rev7Previous[$rev7Key], 'Process') }
    Remove-Variable rev7Secrets, rev7Previous, rev7Value -ErrorAction SilentlyContinue
}
exit $rev7Exit
