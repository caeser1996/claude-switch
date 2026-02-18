#Requires -Version 5.1
<#
.SYNOPSIS
    Claude Switch installer for Windows.

.DESCRIPTION
    Downloads and installs the latest Claude Switch release.

.EXAMPLE
    irm https://raw.githubusercontent.com/sumanta-mukhopadhyay/claude-switch/main/scripts/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

$Repo = "sumanta-mukhopadhyay/claude-switch"
$BinaryName = "claude-switch"
$InstallDir = "$env:LOCALAPPDATA\Programs\claude-switch"

function Write-Info($msg) { Write-Host "i  $msg" -ForegroundColor Blue }
function Write-Ok($msg) { Write-Host "v  $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "!  $msg" -ForegroundColor Yellow }

Write-Host ""
Write-Host "  Claude Switch Installer (Windows)"
Write-Host "  -----------------------------------"
Write-Host ""

# Detect architecture
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit systems are not supported"
    exit 1
}

Write-Info "Architecture: windows_$arch"

# Get latest version
Write-Info "Fetching latest version..."
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$version = $release.tag_name
$versionNoV = $version.TrimStart("v")
Write-Info "Latest version: $version"

# Download
$downloadUrl = "https://github.com/$Repo/releases/download/$version/${BinaryName}_${versionNoV}_windows_${arch}.zip"
Write-Info "Downloading: $downloadUrl"

$tmpDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "cs-install-$(Get-Random)")
$zipPath = Join-Path $tmpDir "archive.zip"

Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

# Extract
Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

# Install
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Copy-Item (Join-Path $tmpDir "$BinaryName.exe") (Join-Path $InstallDir "$BinaryName.exe") -Force

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    Write-Info "Adding $InstallDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

# Cleanup
Remove-Item -Path $tmpDir -Recurse -Force

Write-Ok "Installed $BinaryName $version to $InstallDir"
Write-Host ""

# Verify
& (Join-Path $InstallDir "$BinaryName.exe") version

Write-Host ""
Write-Ok "Installation complete! Restart your terminal and run 'claude-switch --help' to get started."
