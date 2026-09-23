$ErrorActionPreference = 'Stop'
$licenseRoot = Split-Path -Parent $PSScriptRoot
$licenseDestination = Join-Path $licenseRoot 'docs/third-party'
New-Item -ItemType Directory -Path $licenseDestination -Force | Out-Null
Push-Location $licenseRoot
try {
    $licenseModules = & go list -deps -f '{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}' ./cmd/craft
    if ($LASTEXITCODE -ne 0) { throw 'Unable to enumerate CLI dependencies.' }
    $licenseIndex = @('# Bundled Go dependency notices', '', 'Generated from the modules used by cmd/craft. License texts are copied unmodified from the pinned module cache. The CLI includes these dependencies; database servers are not bundled.', '', 'The unmodified MPL-2.0 MySQL driver source is available as the [v1.10.1 source archive](https://proxy.golang.org/github.com/go-sql-driver/mysql/@v/v1.10.1.zip). Source for every pinned module is available through its module path and Go module proxy; see go.mod/go.sum in the Craft source repository.', '', '| Module | Version | Notices |', '| --- | --- | --- |')
    foreach ($licenseLine in ($licenseModules | Where-Object { $_ -and $_ -notlike 'craft|*' } | Sort-Object -Unique)) {
        $licenseParts = $licenseLine -split '\|', 3
        $licenseSlug = $licenseParts[0] -replace '[^a-zA-Z0-9.-]', '_'
        $licenseFiles = @(Get-ChildItem -LiteralPath $licenseParts[2] -File | Where-Object { $_.Name -match '^(LICENSE|COPYING|NOTICE|AUTHORS|PATENTS)(\..*)?$' })
        if ($licenseFiles.Count -eq 0) { throw ('No license found for ' + $licenseParts[0]) }
        $licenseLinks = @()
        foreach ($licenseFile in $licenseFiles) {
            $licenseOutputName = $licenseSlug + '-' + $licenseFile.Name + '.txt'
            Copy-Item -LiteralPath $licenseFile.FullName -Destination (Join-Path $licenseDestination $licenseOutputName) -Force
            $licenseLinks += '[' + $licenseFile.Name + '](' + $licenseOutputName + ')'
        }
        $licenseIndex += '| ' + $licenseParts[0] + ' | ' + $licenseParts[1] + ' | ' + ($licenseLinks -join ', ') + ' |'
    }
    Set-Content -LiteralPath (Join-Path $licenseDestination 'README.md') -Value $licenseIndex -Encoding utf8
} finally { Pop-Location }
