param([string]$SecretsPath)
$ErrorActionPreference = 'Stop'
$rev5Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev5Root '.local/secrets/rev5-databases.clixml' }
$rev5Keys = @('CRAFT_MYSQL_DSN', 'CRAFT_POSTGRES_DSN', 'CRAFT_SQLSERVER_DSN')
$rev5Previous = @{}
$rev5Secrets = Import-Clixml -LiteralPath $SecretsPath
try {
    foreach ($rev5Key in $rev5Keys) {
        $rev5Previous[$rev5Key] = [Environment]::GetEnvironmentVariable($rev5Key, 'Process')
    }
    foreach ($rev5Key in $rev5Keys) {
        if ($rev5Secrets[$rev5Key] -isnot [Security.SecureString]) { throw "Missing encrypted value: $rev5Key" }
        [Environment]::SetEnvironmentVariable($rev5Key, [Net.NetworkCredential]::new('', $rev5Secrets[$rev5Key]).Password, 'Process')
    }
    Push-Location $rev5Root
    try {
        & go test ./tests ./internal/stdlib -run '^TestRev5' -count=1 -v -timeout 180s
        $rev5Exit = $LASTEXITCODE
    } finally { Pop-Location }
} finally {
    foreach ($rev5Key in $rev5Keys) {
        [Environment]::SetEnvironmentVariable($rev5Key, $rev5Previous[$rev5Key], 'Process')
    }
    Remove-Variable rev5Secrets, rev5Previous -ErrorAction SilentlyContinue
}
exit $rev5Exit
