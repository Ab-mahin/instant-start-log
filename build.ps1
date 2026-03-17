# build.ps1 — Pull, build, and deploy mahin CLI
# Usage: pwsh build.ps1

param(
    [string]$BinDir
)

$ErrorActionPreference = "Stop"
$BinaryName = "mahin"

# Detect OS
$IsWindows_ = ($PSVersionTable.PSEdition -eq "Desktop") -or ($IsWindows -eq $true)
$IsMac_ = ($IsMacOS -eq $true)

# Set default bin directory based on OS
if ($BinDir) {
    $DestDir = $BinDir
} elseif ($IsWindows_) {
    $DestDir = "E:\bin-run"
} elseif ($IsMac_) {
    $DestDir = "/usr/local/bin"
} else {
    # Linux
    $DestDir = "/usr/local/bin"
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  🔧 Mahin CLI — Build & Deploy" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Step 1: Git pull
Write-Host "📥 Pulling latest changes..." -ForegroundColor Yellow
git pull
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Git pull failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Pull complete" -ForegroundColor Green
Write-Host ""

# Step 2: Tidy modules
Write-Host "📦 Tidying Go modules..." -ForegroundColor Yellow
go mod tidy
Write-Host "✅ Modules ready" -ForegroundColor Green
Write-Host ""

# Step 3: Build
Write-Host "🔨 Building binary..." -ForegroundColor Yellow
if ($IsWindows_) {
    $BinaryFile = "$BinaryName.exe"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
} else {
    $BinaryFile = $BinaryName
    if ($IsMac_) {
        $env:GOOS = "darwin"
        # Detect architecture
        $Arch = & uname -m
        if ($Arch -eq "arm64") {
            $env:GOARCH = "arm64"
        } else {
            $env:GOARCH = "amd64"
        }
    } else {
        $env:GOOS = "linux"
        $env:GOARCH = "amd64"
    }
}

go build -ldflags="-s -w" -o $BinaryFile .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Built: $BinaryFile" -ForegroundColor Green
Write-Host ""

# Step 4: Deploy to bin directory
Write-Host "📁 Deploying to: $DestDir" -ForegroundColor Yellow

# Create destination directory if it doesn't exist
if (-not (Test-Path $DestDir)) {
    New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
    Write-Host "   Created directory: $DestDir" -ForegroundColor Gray
}

$DestPath = Join-Path $DestDir $BinaryFile
Copy-Item -Path $BinaryFile -Destination $DestPath -Force

# On Mac/Linux, ensure it's executable
if (-not $IsWindows_) {
    chmod +x $DestPath
}

Write-Host "✅ Deployed to: $DestPath" -ForegroundColor Green
Write-Host ""

# Step 5: Clean up local binary
Remove-Item -Path $BinaryFile -Force -ErrorAction SilentlyContinue
Write-Host "🧹 Cleaned local build artifact" -ForegroundColor Gray

# Step 6: Verify
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  ✅ Done! Run 'mahin' to use." -ForegroundColor Green
Write-Host "  📍 Binary at: $DestPath" -ForegroundColor Gray
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Verify it runs
& $DestPath version 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "  🎉 Verified — mahin is ready!" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Binary deployed but 'mahin version' didn't run." -ForegroundColor Yellow
    Write-Host "  Make sure $DestDir is in your PATH." -ForegroundColor Yellow
}
