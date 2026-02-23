$ErrorActionPreference = "Stop"

$Repo = "erfnzdeh/git-retime"
$Binary = "git-retime.exe"
$InstallDir = Join-Path $env:USERPROFILE ".git-retime"

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name
}

function Get-Architecture {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Add-ToPath {
    param([string]$Dir)

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$Dir*") {
        $newPath = "$currentPath;$Dir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$Dir"
        Write-Host "Added $Dir to user PATH."
    }
}

function Main {
    $arch = Get-Architecture
    $platform = "windows_$arch"
    Write-Host "Detected platform: $platform"

    Write-Host "Fetching latest release..."
    $version = Get-LatestVersion
    if (-not $version) {
        throw "Failed to determine latest version."
    }
    Write-Host "Latest version: $version"

    $downloadUrl = "https://github.com/$Repo/releases/download/$version/git-retime_${platform}.zip"

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $tmpFile = Join-Path $env:TEMP "git-retime-download.zip"
    $tmpDir = Join-Path $env:TEMP "git-retime-extract"

    Write-Host "Downloading $downloadUrl..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpFile

    Write-Host "Extracting..."
    if (Test-Path $tmpDir) { Remove-Item -Recurse -Force $tmpDir }
    Expand-Archive -Path $tmpFile -DestinationPath $tmpDir

    Copy-Item (Join-Path $tmpDir $Binary) (Join-Path $InstallDir $Binary) -Force

    Remove-Item $tmpFile -Force -ErrorAction SilentlyContinue
    Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue

    Add-ToPath $InstallDir

    Write-Host ""
    Write-Host "git-retime $version installed successfully."
    Write-Host "Run 'git retime HEAD~5' to get started."
}

Main
