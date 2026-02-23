#Requires -Version 5.1
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$Repo       = 'flowernotfound/google-workspace-mcp-inhouse'
$BinaryName = 'google-workspace-mcp-inhouse'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\google-workspace-mcp-inhouse'
$ConfigDir  = Join-Path $HOME '.config\google-workspace-mcp-inhouse'

# --------------------------------------------------------------------------
# Detect architecture
# --------------------------------------------------------------------------
function Get-Platform {
  switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { return 'windows_amd64' }
    'x86'   {
      Write-Error 'windows/x86 is not supported.'
      exit 1
    }
    default {
      Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
      exit 1
    }
  }
}

# --------------------------------------------------------------------------
# Resolve download URL from GitHub Releases API
# --------------------------------------------------------------------------
function Resolve-DownloadUrl {
  param([string]$Platform)

  $AssetName = "${BinaryName}_${Platform}.exe"
  $ApiUrl    = "https://api.github.com/repos/${Repo}/releases/latest"

  $Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{ Accept = 'application/vnd.github+json' }
  $Asset   = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1

  if (-not $Asset) {
    Write-Error "Could not find asset '${AssetName}' in the latest release."
    exit 1
  }

  return $Asset.browser_download_url
}

# --------------------------------------------------------------------------
# Main
# --------------------------------------------------------------------------
Write-Host "Installing ${BinaryName}..."

$Platform    = Get-Platform
$DownloadUrl = Resolve-DownloadUrl -Platform $Platform

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$DestPath = Join-Path $InstallDir "${BinaryName}.exe"
$TmpPath  = [System.IO.Path]::GetTempFileName() + '.exe'

try {
  Write-Host "Downloading ${BinaryName} (${Platform})..."
  Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpPath -UseBasicParsing
  Move-Item -Force -Path $TmpPath -Destination $DestPath
} finally {
  if (Test-Path $TmpPath) { Remove-Item $TmpPath -Force }
}

# Create config directory
New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

Write-Host ""
Write-Host "v Installed to ${DestPath}"
Write-Host ""

# PATH guidance — use exact entry match to avoid false positives from substring overlap
$UserPath = [System.Environment]::GetEnvironmentVariable('PATH', 'User')
$PathEntries = if ($UserPath) { $UserPath -split ';' } else { @() }
$NormalizedInstallDir = $InstallDir.TrimEnd('\')
$AlreadyInPath = $PathEntries | Where-Object { $_.TrimEnd('\') -ieq $NormalizedInstallDir }
if (-not $AlreadyInPath) {
  $NewPath = if ($UserPath) { "${InstallDir};${UserPath}" } else { $InstallDir }
  [System.Environment]::SetEnvironmentVariable('PATH', $NewPath, 'User')
  Write-Host "Added ${InstallDir} to your PATH (User scope)."
  Write-Host "Restart your terminal to apply the change."
  Write-Host ""
}

# Next steps
Write-Host '------------------------------------------------------------'
Write-Host 'Next steps:'
Write-Host ''
Write-Host '1. Place credentials.json in the config directory:'
Write-Host "   Move-Item ~\Downloads\credentials.json ${ConfigDir}\credentials.json"
Write-Host ''
Write-Host '2. Authenticate with your Google account:'
Write-Host "   ${BinaryName}.exe auth"
Write-Host ''
Write-Host '3. Register with Claude Code:'
Write-Host "   claude mcp add ${BinaryName} ${DestPath}"
Write-Host '------------------------------------------------------------'
