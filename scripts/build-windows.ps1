$ErrorActionPreference = 'Stop'

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

wails build -platform windows/amd64

$Binary = Get-ChildItem -Path "$Root\build\bin" -Filter "custos.exe" | Select-Object -First 1
if (-not $Binary) {
    $Binary = Get-ChildItem -Path "$Root\build\bin" -Filter "*.exe" | Where-Object { $_.Name -notmatch 'installer' } | Select-Object -First 1
}
if (-not $Binary) {
    throw "built binary not found in $Root\build\bin"
}

$OutDir = Join-Path $Root 'dist\windows'
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
Copy-Item $Binary.FullName (Join-Path $OutDir 'custos.exe') -Force
Write-Output (Join-Path $OutDir 'custos.exe')