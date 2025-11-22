# SEHAT-Chain Multi-Node Launcher for Windows
# Requires PowerShell 5.0 or higher

Write-Host "========================================" -ForegroundColor Green
Write-Host "SEHAT-Chain Multi-Node Launcher" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Store process IDs for cleanup
$script:processes = @()
$script:startedNodes = @()

# Cleanup function
function Cleanup {
    Write-Host ""
    Write-Host "Cleaning up processes..." -ForegroundColor Yellow
    foreach ($proc in $script:processes) {
        try {
            Stop-Process -Id $proc -Force -ErrorAction SilentlyContinue
        } catch {}
    }
    Write-Host "Cleanup complete" -ForegroundColor Green
}

# Register cleanup on Ctrl+C
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Cleanup }

# Create directories
Write-Host "Setting up directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "logs" | Out-Null
New-Item -ItemType Directory -Force -Path "configs" | Out-Null

# Build binary
Write-Host "Building blockchain node..." -ForegroundColor Yellow
$buildResult = go build -o sehat-chain.exe cmd\node\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Build successful" -ForegroundColor Green
Write-Host ""

# Function to start a node
function Start-Node {
    param(
        [string]$NodeId,
        [string]$Color = "Green"
    )
    
    $configPath = "configs\$NodeId.json"
    $logPath = "logs\$NodeId.log"
    
    # Start process in background and redirect output to log file
    $process = Start-Process -FilePath ".\sehat-chain.exe" `
                            -ArgumentList "-config $configPath" `
                            -RedirectStandardOutput $logPath `
                            -RedirectStandardError "logs\$NodeId-error.log" `
                            -WindowStyle Normal `
                            -PassThru
    
    $script:processes += $process.Id
    
    Write-Host "$NodeId started (PID: $($process.Id))" -ForegroundColor $Color
    Start-Sleep -Seconds 1
}

# Auto-detect and start nodes from configs folder
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Auto-detecting and Starting Nodes..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$configFiles = Get-ChildItem -Path "configs" -Filter "*.json"

if ($configFiles.Count -eq 0) {
    Write-Host "No configuration files found in 'configs' directory!" -ForegroundColor Red
    Write-Host "Please ensure .json files exist in the configs folder."
} else {
    foreach ($file in $configFiles) {
        try {
            # Read JSON to determine role and port
            $jsonContent = Get-Content $file.FullName | ConvertFrom-Json
            $fileBaseName = $file.BaseName # Filename without extension (e.g., "validator-1")
            $internalId = $jsonContent.node_id
            $port = $jsonContent.port
            
            # Determine if this node is a validator (if its ID is in the validators list)
            $isValidator = $false
            if ($jsonContent.validators) {
                foreach ($v in $jsonContent.validators) {
                    if ($v.ID -eq $internalId) {
                        $isValidator = $true
                        break
                    }
                }
            }

            # Set Color and Role Label
            if ($isValidator) {
                $roleColor = "Green"
                $roleLabel = "Validator"
            } else {
                $roleColor = "Blue"
                $roleLabel = "Light Node"
            }

            # Start the Node
            Start-Node -NodeId $fileBaseName -Color $roleColor

            # Store info for status display
            $script:startedNodes += [PSCustomObject]@{
                Name = $fileBaseName
                Port = $port
                Role = $roleLabel
                Color = $roleColor
            }
        } catch {
            Write-Host "Error parsing config for $($file.Name): $_" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "All nodes are running!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Node Status Summary:" -ForegroundColor Yellow

# Dynamic Status Display
foreach ($node in $script:startedNodes) {
    Write-Host "  $($node.Name) (Port $($node.Port)) - $($node.Role)" -ForegroundColor $node.Color
}

Write-Host ""
Write-Host "Logs available in:" -ForegroundColor Yellow
foreach ($node in $script:startedNodes) {
    Write-Host "  logs\$($node.Name).log"
}

Write-Host ""
Write-Host "View live logs example:" -ForegroundColor Yellow
if ($script:startedNodes.Count -gt 0) {
    $firstNode = $script:startedNodes[0].Name
    Write-Host "  Get-Content logs\$firstNode.log -Wait -Tail 20" -ForegroundColor Cyan
} else {
    Write-Host "  Get-Content logs\filename.log -Wait -Tail 20" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Kill all nodes:" -ForegroundColor Yellow
Write-Host "  Get-Process | Where-Object {`$_.ProcessName -eq 'sehat-chain'} | Stop-Process -Force" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C to stop all nodes" -ForegroundColor Red
Write-Host ""

# Keep script running
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Cleanup
}