param([switch]$Installer, [switch]$Icons, [switch]$Language, [string]$IsccPath)
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$oldCGO = $env:CGO_ENABLED
$oldGOOS = $env:GOOS
$oldGOARCH = $env:GOARCH
Push-Location $repoRoot
try {
    $env:CGO_ENABLED = '0'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    New-Item -ItemType Directory -Path (Join-Path $repoRoot 'dist') -Force | Out-Null
    & go build -trimpath -ldflags '-s -w' -o dist/craft.exe ./cmd/craft
    if ($LASTEXITCODE -ne 0) { throw 'Go build failed.' }
    & (Join-Path $PSScriptRoot 'collect-licenses.ps1')
    if ($Icons) {
        Push-Location (Join-Path $repoRoot 'craft-file-icon-theme')
        try {
            & npm exec --yes --package @vscode/vsce@4.0.0 -- vsce package --allow-missing-repository --out ../dist/craft-forge-file-icons-0.2.0.vsix
            if ($LASTEXITCODE -ne 0) { throw 'Icon theme packaging failed.' }
        } finally { Pop-Location }
    }
    if ($Language -or $Installer) {
        Push-Location (Join-Path $repoRoot 'craft-vscode')
        try {
            & npm exec --yes --package @vscode/vsce@4.0.0 -- vsce package --no-dependencies --allow-missing-repository --out ../dist/craft-language-support-0.2.2.vsix
            if ($LASTEXITCODE -ne 0) { throw 'Language extension packaging failed.' }
        } finally { Pop-Location }
    }
    if ($Installer) {
        if (-not $IsccPath) {
            $compiler = Get-Command ISCC.exe -ErrorAction SilentlyContinue
            if ($compiler) { $IsccPath = $compiler.Source }
            else {
                $candidates = @(
                    (Join-Path $repoRoot 'build/tools/inno/ISCC.exe'),
                    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
                    "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe",
                    "$env:ProgramFiles\Inno Setup 7\ISCC.exe"
                )
                $IsccPath = $candidates | Where-Object { Test-Path -LiteralPath $_ } | Select-Object -First 1
            }
        }
        if (-not $IsccPath -or -not (Test-Path -LiteralPath $IsccPath)) {
            throw 'Inno Setup compiler not found. Install Inno Setup 6/7 and pass -IsccPath <path-to-ISCC.exe>.'
        }
        & $IsccPath (Join-Path $repoRoot 'packaging/windows/craft.iss')
        if ($LASTEXITCODE -ne 0) { throw 'Installer compilation failed.' }
    }
    $files = @('dist/craft.exe')
    if ($Installer) { $files += 'dist/Craft-setup.exe' }
    if (Test-Path -LiteralPath 'dist/craft-forge-file-icons-0.2.0.vsix') { $files += 'dist/craft-forge-file-icons-0.2.0.vsix' }
    if (Test-Path -LiteralPath 'dist/craft-language-support-0.2.2.vsix') { $files += 'dist/craft-language-support-0.2.2.vsix' }
    $files | ForEach-Object {
        $hash = Get-FileHash -LiteralPath $_ -Algorithm SHA256
        '{0}  {1}' -f $hash.Hash.ToLowerInvariant(), (Split-Path -Leaf $_)
    } | Set-Content -LiteralPath 'dist/SHA256SUMS.txt' -Encoding ascii
    Write-Host 'Build complete. No tests were run.'
} finally {
    $env:CGO_ENABLED = $oldCGO
    $env:GOOS = $oldGOOS
    $env:GOARCH = $oldGOARCH
    Pop-Location
}

