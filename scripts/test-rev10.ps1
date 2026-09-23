param([string]$SecretsPath, [switch]$Race, [switch]$Stress)
$ErrorActionPreference = 'Stop'
& (Join-Path $PSScriptRoot 'test-rev9.ps1') -SecretsPath $SecretsPath -Race:$Race -Stress:$Stress
if ($LASTEXITCODE -ne 0) { throw 'Rev10 regression failed' }
