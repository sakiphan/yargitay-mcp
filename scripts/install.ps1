# SPDX-License-Identifier: AGPL-3.0-only
param(
  [Parameter(Mandatory=$true)][ValidateSet('codex','claude','claude-desktop','gemini')][string]$Client,
  [ValidatePattern('^v\d+\.\d+\.\d+$')][string]$Version = 'v0.3.1'
)
$ErrorActionPreference = 'Stop'
$platformArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
$arch = switch ($platformArch) { 'x64' {'amd64'} 'arm64' {'arm64'} default {throw 'Desteklenmeyen işlemci.'} }
$asset = "yargitay-mcp_windows_$arch.exe"
$base = "https://github.com/sakiphan/yargitay-mcp/releases/download/$Version"
$workDir = Join-Path ([System.IO.Path]::GetTempPath()) ('yargitay-install-' + [guid]::NewGuid())
$null = New-Item -ItemType Directory -Path $workDir
try {
  $download = Join-Path $workDir $asset
  $checksums = Join-Path $workDir 'checksums.txt'
  Invoke-WebRequest "$base/$asset" -OutFile $download
  Invoke-WebRequest "$base/checksums.txt" -OutFile $checksums
  $matching = @(Get-Content $checksums | Where-Object { $_ -match ('^[a-fA-F0-9]{64}\s+' + [regex]::Escape($asset) + '$') })
  if ($matching.Count -ne 1) { throw 'Geçerli checksum bulunamadı.' }
  $expected = ($matching[0] -split '\s+')[0]
  if ((Get-FileHash -Algorithm SHA256 $download).Hash -ne $expected) { throw 'SHA-256 uyuşmuyor.' }
  $dest = Join-Path $env:LOCALAPPDATA 'Programs\yargitay-mcp'
  $null = New-Item -ItemType Directory -Force -Path $dest
  $target = Join-Path $dest 'yargitay-mcp.exe'
  if (Test-Path $target) {
    if ((Get-Item $target).Attributes -band [System.IO.FileAttributes]::ReparsePoint) { throw 'Hedef bağlantı; kurulum durduruldu.' }
    $backupTarget = Join-Path $dest 'yargitay-mcp.previous.exe'
    if ((Test-Path $backupTarget) -and ((Get-Item $backupTarget).Attributes -band [System.IO.FileAttributes]::ReparsePoint)) { throw 'Yedek hedefi bağlantı; kurulum durduruldu.' }
    Copy-Item $target $backupTarget -Force
  }
  Move-Item $download $target -Force
  & $target install --client $Client
  if ($LASTEXITCODE -ne 0) { throw 'İstemci ayarı eklenemedi; mevcut ayarlar korunuyor.' }
  & $target version
  Write-Output "Kuruldu: $target — istemcinizi yeniden başlatın."
} finally {
  # Only files created in this unique temporary directory are removed.
  Get-ChildItem -LiteralPath $workDir -File | Remove-Item -Force
  Remove-Item -LiteralPath $workDir
}
