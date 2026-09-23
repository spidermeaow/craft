param([string]$SecretsPath, [switch]$Regression, [switch]$Race)
$ErrorActionPreference = 'Stop'
$rev6Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev6Root '.local/secrets/rev5-databases.clixml' }
$rev6Keys = @('CRAFT_MYSQL_DSN', 'CRAFT_POSTGRES_DSN', 'CRAFT_SQLSERVER_DSN')
$rev6Previous = @{}
foreach ($rev6Key in $rev6Keys) { $rev6Previous[$rev6Key] = [Environment]::GetEnvironmentVariable($rev6Key, 'Process') }
try {
    $rev6Secrets = Import-Clixml -LiteralPath $SecretsPath
    foreach ($rev6Key in $rev6Keys) {
        if ($rev6Secrets[$rev6Key] -isnot [Security.SecureString]) { throw "Missing encrypted value: $rev6Key" }
        $rev6Value = [Net.NetworkCredential]::new('', $rev6Secrets[$rev6Key]).Password
        if ([string]::IsNullOrWhiteSpace($rev6Value)) { throw "Empty value: $rev6Key" }
        [Environment]::SetEnvironmentVariable($rev6Key, $rev6Value, 'Process')
    }
    Push-Location $rev6Root
    try {
        $rev6Args = @('test', './tests', './internal/stdlib', '-run', '^TestRev6', '-count=1', '-v', '-timeout', '600s')
        if ($Regression) { $rev6Args = @('test', './...', '-count=1', '-v', '-timeout', '600s') }
        if ($Race) { $rev6Args += '-race' }
        & go @rev6Args
        $rev6Exit = $LASTEXITCODE
    } finally { Pop-Location }
} finally {
    foreach ($rev6Key in $rev6Keys) { [Environment]::SetEnvironmentVariable($rev6Key, $rev6Previous[$rev6Key], 'Process') }
    Remove-Variable rev6Secrets, rev6Previous, rev6Value -ErrorAction SilentlyContinue
}
exit $rev6Exit
