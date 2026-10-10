param([string]$SecretsPath)
$ErrorActionPreference = 'Stop'
$rev45Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev45Root '.local/secrets/rev5-databases.clixml' }
$rev45PreviousDriver = [Environment]::GetEnvironmentVariable('CRAFT_DB_DRIVER', 'Process')
$rev45PreviousDsn = [Environment]::GetEnvironmentVariable('CRAFT_DB_DSN', 'Process')
try {
    $rev45Secrets = Import-Clixml -LiteralPath $SecretsPath
    foreach ($rev45Driver in @('mysql', 'postgres', 'sqlserver')) {
        $rev45Key = 'CRAFT_' + $rev45Driver.ToUpperInvariant() + '_DSN'
        $rev45Secret = $rev45Secrets[$rev45Key]
        if ($rev45Secret -isnot [Security.SecureString]) { throw "Missing encrypted value: $rev45Key" }
        [Environment]::SetEnvironmentVariable('CRAFT_DB_DRIVER', $rev45Driver, 'Process')
        [Environment]::SetEnvironmentVariable('CRAFT_DB_DSN', [Net.NetworkCredential]::new('', $rev45Secret).Password, 'Process')
        Push-Location (Join-Path $rev45Root 'examples/db-package')
        try {
            & (Join-Path $rev45Root 'dist/craft-rev45.exe') run
            if ($LASTEXITCODE -ne 0) { throw "DB package example failed for $rev45Driver" }
        } finally { Pop-Location }
    }
} finally {
    [Environment]::SetEnvironmentVariable('CRAFT_DB_DRIVER', $rev45PreviousDriver, 'Process')
    [Environment]::SetEnvironmentVariable('CRAFT_DB_DSN', $rev45PreviousDsn, 'Process')
    Remove-Variable rev45Secrets, rev45Secret -ErrorAction SilentlyContinue
}
