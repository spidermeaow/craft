param([string]$SecretsPath, [switch]$Race, [switch]$Stress)
$ErrorActionPreference = 'Stop'
$rev9Root = Split-Path -Parent $PSScriptRoot

Push-Location $rev9Root
try {
    & (Join-Path $PSScriptRoot 'test-rev8.ps1') -SecretsPath $SecretsPath -Race:$Race -Stress:$Stress
    if ($LASTEXITCODE -ne 0) { throw 'Rev.8 regression failed before Rev.9 acceptance.' }

    & go test ./internal/stdlib -run '^TestRev9' -count=1 -v
    if ($LASTEXITCODE -ne 0) { throw 'Rev.9 tooling acceptance failed.' }
} finally {
    Pop-Location
}
