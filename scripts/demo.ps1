[CmdletBinding()]
param([ValidateSet('Start', 'Status', 'Stop')][string]$Action = 'Status')
$ErrorActionPreference = 'Stop'
$projectRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
Push-Location $projectRoot
try {
    switch ($Action) {
        'Start' { & docker compose up --build -d --wait }
        'Status' { & docker compose ps -a }
        'Stop' { & docker compose stop }
    }
    if ($LASTEXITCODE -ne 0) { throw "Docker Compose fallo (codigo $LASTEXITCODE)." }
} finally { Pop-Location }
