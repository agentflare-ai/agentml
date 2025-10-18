# agentmlx PowerShell installer
# Installs the latest version of agentmlx from Cloudflare R2
# Usage: iwr -useb ps.agentml.dev | iex

param(
    [string]$Version = "",
    [ValidateSet("latest", "next", "beta")]
    [string]$Channel = "latest",
    [switch]$NoModifyPath,
    [switch]$Help
)

# Configuration
$ReleasesBaseUrl = if ($env:AGENTMLX_RELEASES_URL) { $env:AGENTMLX_RELEASES_URL } else { "https://amlx.agentml.dev/agentmlx" }
$InstallDir = if ($env:AGENTMLX_INSTALL_DIR) { $env:AGENTMLX_INSTALL_DIR } else { Join-Path $env:USERPROFILE ".agentmlx" }
$BinDir = Join-Path $InstallDir "bin"

# Override with environment variables if set
if ($env:AGENTMLX_VERSION -and -not $Version) {
    $Version = $env:AGENTMLX_VERSION
}
if ($env:AGENTMLX_CHANNEL -and $Channel -eq "latest") {
    $Channel = $env:AGENTMLX_CHANNEL
}

# Help text
if ($Help) {
    Write-Host "agentmlx PowerShell installer"
    Write-Host ""
    Write-Host "Usage: install.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version <version>     Install specific version (without 'v' prefix)"
    Write-Host "  -Channel <channel>     Install from channel: latest, next, beta (default: latest)"
    Write-Host "  -NoModifyPath          Don't modify PowerShell profile"
    Write-Host "  -Help                  Show this help message"
    Write-Host ""
    Write-Host "Channels:"
    Write-Host "  latest                 Stable releases (v1.0.0)"
    Write-Host "  next                   Release candidates (v1.0.0-rc.1)"
    Write-Host "  beta                   Beta releases (v1.0.0-beta.1)"
    Write-Host ""
    Write-Host "Channel Fallback:"
    Write-Host "  If a channel has no releases, it automatically falls back:"
    Write-Host "  latest → next → beta"
    Write-Host ""
    Write-Host "Environment variables:"
    Write-Host "  AGENTMLX_VERSION        Version to install (without 'v' prefix)"
    Write-Host "  AGENTMLX_CHANNEL        Channel to install from (latest, next, beta)"
    Write-Host "  AGENTMLX_INSTALL_DIR    Installation directory (default: `$HOME\.agentmlx)"
    Write-Host "  AGENTMLX_RELEASES_URL   Base URL for releases"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\install.ps1                      # Install latest stable"
    Write-Host "  .\install.ps1 -Channel next        # Install latest RC"
    Write-Host "  .\install.ps1 -Version 1.0.0-rc.1  # Install specific version"
    exit 0
}

# Helper functions
function Write-Info {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "Warning: " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error-Exit {
    param([string]$Message)
    Write-Host "Error: " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

# Detect platform
function Get-Platform {
    $os = "windows"

    # Detect architecture
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { "amd64" }
        "ARM64" { "arm64" }
        default {
            Write-Error-Exit "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        }
    }

    # Check if running on ARM64 emulating x64
    if ($arch -eq "amd64" -and $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
        $arch = "arm64"
        Write-Info "Detected ARM64 system, using arm64 binary"
    }

    return "${os}_${arch}"
}

# Check if a URL exists
function Test-UrlExists {
    param([string]$Url)

    try {
        $response = Invoke-WebRequest -Uri $Url -Method Head -UseBasicParsing -ErrorAction Stop
        return $response.StatusCode -eq 200
    }
    catch {
        return $false
    }
}

# Get binary URL with channel fallback
function Get-BinaryUrlWithFallback {
    param(
        [string]$Channel,
        [string]$Platform
    )

    Write-Info "Checking $Channel channel..."
    $binaryUrl = "$ReleasesBaseUrl/$Channel/agentmlx_$Platform"
    $checksumsUrl = "$ReleasesBaseUrl/$Channel/checksums.txt"

    if (Test-UrlExists $binaryUrl) {
        return @{
            BinaryUrl = $binaryUrl
            ChecksumsUrl = $checksumsUrl
        }
    }

    # Waterfall logic: latest → next → beta
    switch ($Channel) {
        "latest" {
            Write-Warn "No binary found in latest channel, trying next (rc) channel..."
            $binaryUrl = "$ReleasesBaseUrl/next/agentmlx_$Platform"
            $checksumsUrl = "$ReleasesBaseUrl/next/checksums.txt"
            if (Test-UrlExists $binaryUrl) {
                Write-Warn "Using next channel"
                return @{
                    BinaryUrl = $binaryUrl
                    ChecksumsUrl = $checksumsUrl
                }
            }

            Write-Warn "No binary found in next channel, trying beta channel..."
            $binaryUrl = "$ReleasesBaseUrl/beta/agentmlx_$Platform"
            $checksumsUrl = "$ReleasesBaseUrl/beta/checksums.txt"
            if (Test-UrlExists $binaryUrl) {
                Write-Warn "Using beta channel"
                return @{
                    BinaryUrl = $binaryUrl
                    ChecksumsUrl = $checksumsUrl
                }
            }
        }
        "next" {
            Write-Warn "No binary found in next channel, trying beta channel..."
            $binaryUrl = "$ReleasesBaseUrl/beta/agentmlx_$Platform"
            $checksumsUrl = "$ReleasesBaseUrl/beta/checksums.txt"
            if (Test-UrlExists $binaryUrl) {
                Write-Warn "Using beta channel"
                return @{
                    BinaryUrl = $binaryUrl
                    ChecksumsUrl = $checksumsUrl
                }
            }
        }
    }

    return $null
}

# Download file
function Get-File {
    param(
        [string]$Url,
        [string]$Output
    )

    try {
        Invoke-WebRequest -Uri $Url -OutFile $Output -UseBasicParsing -ErrorAction Stop
    }
    catch {
        Write-Error-Exit "Failed to download from $Url : $_"
    }
}

# Verify checksum
function Test-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumsFile,
        [string]$BinaryName
    )

    if (-not (Test-Path $ChecksumsFile)) {
        Write-Warn "Checksums file not found. Skipping verification."
        return
    }

    # Read checksums file and find the expected checksum
    $expectedSum = Get-Content $ChecksumsFile |
        Where-Object { $_ -match $BinaryName } |
        ForEach-Object { ($_ -split '\s+')[0] }

    if (-not $expectedSum) {
        Write-Warn "Could not find checksum for $BinaryName. Skipping verification."
        return
    }

    # Calculate actual checksum
    $actualSum = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()

    if ($expectedSum -ne $actualSum) {
        Write-Error-Exit "Checksum verification failed!`nExpected: $expectedSum`nActual:   $actualSum"
    }

    Write-Info "Checksum verified successfully"
}

# Setup PATH in PowerShell profile
function Add-ToPath {
    # Check if already in PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -split ';' | Where-Object { $_ -eq $BinDir }) {
        return
    }

    # Get PowerShell profile path
    $profilePath = $PROFILE.CurrentUserAllHosts

    # Create profile directory if it doesn't exist
    $profileDir = Split-Path $profilePath -Parent
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
    }

    # Check if already in profile
    if (Test-Path $profilePath) {
        $profileContent = Get-Content $profilePath -Raw
        if ($profileContent -match [regex]::Escape($BinDir)) {
            return
        }
    }

    # Add to profile
    Write-Info "Adding $BinDir to PATH in PowerShell profile"
    @"

# agentmlx
`$env:PATH = "`$env:PATH;$BinDir"
"@ | Add-Content -Path $profilePath -Encoding UTF8

    Write-Host ""
    Write-Host "Note: You may need to restart your PowerShell session or run:" -ForegroundColor Yellow
    Write-Host "  . `$PROFILE" -ForegroundColor Cyan
    Write-Host ""

    # Also add to current session's user PATH
    try {
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$BinDir", "User")
        $env:PATH = "$env:PATH;$BinDir"
    }
    catch {
        Write-Warn "Could not update user PATH environment variable. You may need to add $BinDir to PATH manually."
    }
}

# Main installation
try {
    Write-Info "Installing agentmlx..."

    # Detect platform
    $platform = Get-Platform
    Write-Info "Detected platform: $platform"

    # Get binary URLs
    $urls = $null
    if ($Version) {
        Write-Info "Version: $Version"
        $binaryName = "agentmlx_${Version}_${platform}"
        $binaryUrl = "$ReleasesBaseUrl/v$Version/$binaryName"
        $checksumsUrl = "$ReleasesBaseUrl/v$Version/checksums.txt"
        $urls = @{
            BinaryUrl = $binaryUrl
            ChecksumsUrl = $checksumsUrl
        }
    }
    else {
        $urls = Get-BinaryUrlWithFallback -Channel $Channel -Platform $platform
        if (-not $urls) {
            Write-Error-Exit "Failed to find any releases for channel: $Channel"
        }
        $binaryName = Split-Path $urls.BinaryUrl -Leaf
    }

    # Create temporary directory
    $tmpDir = Join-Path $env:TEMP "agentmlx-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download binary
        Write-Info "Downloading agentmlx..."
        $tmpBinary = Join-Path $tmpDir "agentmlx.exe"
        Get-File -Url $urls.BinaryUrl -Output $tmpBinary

        # Download checksums
        Write-Info "Downloading checksums..."
        $tmpChecksums = Join-Path $tmpDir "checksums.txt"
        try {
            Get-File -Url $urls.ChecksumsUrl -Output $tmpChecksums
        }
        catch {
            Write-Warn "Failed to download checksums"
        }

        # Verify checksum
        if (Test-Path $tmpChecksums) {
            Write-Info "Verifying checksum..."
            Test-Checksum -FilePath $tmpBinary -ChecksumsFile $tmpChecksums -BinaryName $binaryName
        }

        # Create installation directory
        if (-not (Test-Path $BinDir)) {
            New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
        }

        # Install binary
        Write-Info "Installing to $BinDir\agentmlx.exe..."
        $finalBinary = Join-Path $BinDir "agentmlx.exe"

        # Remove existing binary if it exists
        if (Test-Path $finalBinary) {
            Remove-Item $finalBinary -Force
        }

        Move-Item -Path $tmpBinary -Destination $finalBinary -Force

        # Create amlx.exe copy (PowerShell doesn't support symlinks reliably on all Windows versions)
        Write-Info "Creating amlx alias..."
        $amlxPath = Join-Path $BinDir "amlx.exe"
        if (Test-Path $amlxPath) {
            Remove-Item $amlxPath -Force
        }
        Copy-Item -Path $finalBinary -Destination $amlxPath -Force

        # Setup PATH
        if (-not $NoModifyPath) {
            Add-ToPath
        }

        # Show installed version
        Write-Host ""
        if ($Version) {
            Write-Host "==> " -ForegroundColor Green -NoNewline
            Write-Host "✓ " -ForegroundColor Green -NoNewline
            Write-Host "agentmlx v$Version installed successfully!"
        }
        else {
            try {
                $installedVersion = & $finalBinary --version 2>&1 | Select-String -Pattern '\d+\.\d+\.\d+(-[a-z]+\.\d+)?' | ForEach-Object { $_.Matches[0].Value }
                if ($installedVersion) {
                    Write-Host "==> " -ForegroundColor Green -NoNewline
                    Write-Host "✓ " -ForegroundColor Green -NoNewline
                    Write-Host "agentmlx v$installedVersion installed successfully!"
                }
                else {
                    Write-Host "==> " -ForegroundColor Green -NoNewline
                    Write-Host "✓ " -ForegroundColor Green -NoNewline
                    Write-Host "agentmlx installed successfully!"
                }
            }
            catch {
                Write-Host "==> " -ForegroundColor Green -NoNewline
                Write-Host "✓ " -ForegroundColor Green -NoNewline
                Write-Host "agentmlx installed successfully!"
            }
        }

        Write-Host ""
        Write-Host "To get started, run:"
        Write-Host "  agentmlx --help" -ForegroundColor Cyan
        Write-Host "  amlx --help" -ForegroundColor Cyan
        Write-Host ""

        if (-not $NoModifyPath) {
            $currentPath = $env:PATH -split ';'
            if (-not ($currentPath -contains $BinDir)) {
                Write-Host "Note: You may need to restart your PowerShell session or run:" -ForegroundColor Yellow
                Write-Host "  . `$PROFILE" -ForegroundColor Cyan
                Write-Host "Or manually add to your PATH:" -ForegroundColor Yellow
                Write-Host "  `$env:PATH += `";$BinDir`"" -ForegroundColor Cyan
            }
        }
    }
    finally {
        # Cleanup
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}
catch {
    Write-Error-Exit $_.Exception.Message
}
