param([string]$SecretsPath, [switch]$Race, [switch]$Stress, [switch]$Live)
$ErrorActionPreference = 'Stop'
& (Join-Path $PSScriptRoot 'test-rev10.ps1') -SecretsPath $SecretsPath -Race:$Race -Stress:$Stress
if ($LASTEXITCODE -ne 0) { throw 'Rev11 regression failed' }
if ($Live) {
    $rev11Previous = $env:CRAFT_REV11_LIVE
    Push-Location (Split-Path -Parent $PSScriptRoot)
    try {
        $env:CRAFT_REV11_LIVE = '1'
        & go test ./tests -run '^TestRev11LiveGitHub$' -count=1 -v -timeout 300s
        if ($LASTEXITCODE -ne 0) { throw 'Live GitHub acceptance failed' }
    } finally {
        $env:CRAFT_REV11_LIVE = $rev11Previous
        Pop-Location
    }
}
