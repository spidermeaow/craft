param([ValidateSet('mysql','postgres','sqlserver')][string]$Driver = 'mysql', [string]$SecretsPath)
$ErrorActionPreference = 'Stop'
$rev6Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev6Root '.local/secrets/rev5-databases.clixml' }
$rev6Keys = @('CRAFT_DB_DRIVER','CRAFT_DB_DSN','CRAFT_DB_RUN_ID','CRAFT_HTTP_ADDRESS')
$rev6Previous = @{}
foreach ($rev6Key in $rev6Keys) { $rev6Previous[$rev6Key] = [Environment]::GetEnvironmentVariable($rev6Key, 'Process') }
try {
    $rev6Secrets = Import-Clixml -LiteralPath $SecretsPath
    $rev6Secret = $rev6Secrets['CRAFT_' + $Driver.ToUpperInvariant() + '_DSN']
    if ($rev6Secret -isnot [Security.SecureString]) { throw 'Missing encrypted database DSN' }
    $env:CRAFT_DB_DRIVER = $Driver
    $env:CRAFT_DB_DSN = [Net.NetworkCredential]::new('', $rev6Secret).Password
    $env:CRAFT_DB_RUN_ID = [Guid]::NewGuid().ToString('N')
    $env:CRAFT_HTTP_ADDRESS = '127.0.0.1:8080'
    Push-Location (Join-Path $rev6Root 'examples/rev6-http-db')
    try {
        & (Join-Path $rev6Root 'dist/craft.exe') run
        $rev6Exit = $LASTEXITCODE
    } finally { Pop-Location }
} finally {
    foreach ($rev6Key in $rev6Keys) { [Environment]::SetEnvironmentVariable($rev6Key, $rev6Previous[$rev6Key], 'Process') }
    Remove-Variable rev6Secrets, rev6Secret, rev6Previous -ErrorAction SilentlyContinue
}
exit $rev6Exit
