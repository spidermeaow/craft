param([string]$SecretsPath, [switch]$Race, [switch]$Stress)
$ErrorActionPreference = 'Stop'
$rev8Root = Split-Path -Parent $PSScriptRoot
if (-not $SecretsPath) { $SecretsPath = Join-Path $rev8Root '.local/secrets/rev5-databases.clixml' }

Push-Location $rev8Root
try {
    & (Join-Path $PSScriptRoot 'test-rev7.ps1') -SecretsPath $SecretsPath -Regression -Race:$Race
    if ($LASTEXITCODE -ne 0) { throw 'Rev.8 full regression failed.' }

    if ($Stress) {
        $rev8PreviousStress = $env:CRAFT_STRESS
        try {
            $env:CRAFT_STRESS = '1'
            & go test ./tests -run '^TestRev4HTTPStress$' -count=1 -v -timeout 300s
            if ($LASTEXITCODE -ne 0) { throw 'Rev.8 HTTP stress failed.' }
        } finally {
            $env:CRAFT_STRESS = $rev8PreviousStress
        }
    }
} finally {
    Pop-Location
}
