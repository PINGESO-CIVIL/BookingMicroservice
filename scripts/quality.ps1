[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('Init', 'Validate', 'Start', 'Setup', 'Coverage', 'Sonar', 'Trivy', 'Check', 'Status', 'Stop')]
    [string]$Action = 'Status'
)

$ErrorActionPreference = 'Stop'
$projectRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$stateDirectory = Join-Path $projectRoot '.quality'
$environmentFile = Join-Path $stateDirectory 'local.env'
$composeFile = Join-Path $projectRoot 'infra/quality/compose.yaml'
$reportDirectory = Join-Path $projectRoot 'reports'
$utf8 = [System.Text.UTF8Encoding]::new($false)

function New-LocalSecret {
    $bytes = New-Object byte[] 32
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $generator.GetBytes($bytes) } finally { $generator.Dispose() }
    return ([System.BitConverter]::ToString($bytes)).Replace('-', '').ToLowerInvariant()
}

function Initialize-LocalFiles {
    [System.IO.Directory]::CreateDirectory($stateDirectory) | Out-Null
    [System.IO.Directory]::CreateDirectory($reportDirectory) | Out-Null
    [System.IO.Directory]::CreateDirectory((Join-Path $reportDirectory 'sonar')) | Out-Null
    if (-not (Test-Path -LiteralPath $environmentFile)) {
        $lines = @(
            'SONAR_PORT=9000',
            ('SONAR_DB_PASSWORD=' + (New-LocalSecret)),
            ('SONAR_ADMIN_PASSWORD=Aa7!' + (New-LocalSecret)),
            'SONAR_TOKEN='
        )
        [System.IO.File]::WriteAllLines($environmentFile, $lines, $utf8)
    }
}

function Read-LocalEnvironment {
    $values = @{}
    foreach ($line in [System.IO.File]::ReadAllLines($environmentFile)) {
        if ($line -match '^([A-Z_]+)=(.*)$') { $values[$Matches[1]] = $Matches[2] }
    }
    return $values
}

function Invoke-Compose {
    param([string[]]$DockerArguments)
    & docker compose --env-file $environmentFile -f $composeFile @DockerArguments
    if ($LASTEXITCODE -ne 0) { throw "Docker Compose fallo (codigo $LASTEXITCODE)." }
}

function Assert-DockerEngine {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { throw 'Docker no esta instalado.' }
    & docker info --format '{{.ServerVersion}}' | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Inicia Docker Desktop con contenedores Linux y vuelve a ejecutar.' }
}

function Write-CoverageReports {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Go no esta instalado.' }
    $coveragePath = Join-Path $reportDirectory 'coverage.out'
    $testPath = Join-Path $reportDirectory 'go-tests.json'
    # Never let a failed run leave old reports available to the scanner.
    foreach ($path in @($coveragePath, $testPath)) {
        if (Test-Path -LiteralPath $path) { Remove-Item -LiteralPath $path }
    }
    Push-Location (Join-Path $projectRoot 'Backend')
    try {
        $moduleName = & go list -m
        if ($LASTEXITCODE -ne 0) { throw 'No se pudo resolver el modulo Go.' }
        & go vet ./...
        if ($LASTEXITCODE -ne 0) { throw 'go vet detecto errores.' }
        $testOutput = & go test -json -covermode=atomic "-coverprofile=$coveragePath" ./...
        $testExitCode = $LASTEXITCODE
        [System.IO.File]::WriteAllLines($testPath, [string[]]$testOutput, $utf8)
        if ($testExitCode -ne 0) {
            if (Test-Path -LiteralPath $coveragePath) { Remove-Item -LiteralPath $coveragePath }
            throw 'Las pruebas fallaron; no se enviara cobertura a SonarQube.'
        }
        # Go reports use module paths; the scanner runs from the repository root.
        $prefix = [regex]::Escape($moduleName.Trim() + '/')
        $coverageLines = [System.IO.File]::ReadAllLines($coveragePath) | ForEach-Object {
            $_ -replace ('^' + $prefix), 'Backend/'
        }
        [System.IO.File]::WriteAllLines($coveragePath, [string[]]$coverageLines, $utf8)
        $events = @($testOutput | ForEach-Object { $_ | ConvertFrom-Json })
        $testCount = @($events | Where-Object { $_.Action -eq 'run' -and $_.Test }).Count
        Write-Host "Pruebas ejecutadas: $testCount. Reportes generados en reports/."
        if ($testCount -eq 0) { Write-Host 'Todavia no hay pruebas de comportamiento; el reporte no acredita cobertura funcional.' }
    } finally { Pop-Location }
}

function Connect-LocalSonar {
    $values = Read-LocalEnvironment
    $baseUrl = 'http://127.0.0.1:' + $values['SONAR_PORT']
    $status = Invoke-RestMethod -Uri "$baseUrl/api/system/status" -TimeoutSec 10
    if ($status.status -ne 'UP') { throw 'SonarQube aun no esta listo; consulta Status y vuelve a ejecutar Setup.' }

    function Get-BasicHeader([string]$Password) {
        $encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('admin:' + $Password))
        return @{ Authorization = 'Basic ' + $encoded }
    }
    $adminHeader = Get-BasicHeader $values['SONAR_ADMIN_PASSWORD']
    $valid = $false
    try { $valid = (Invoke-RestMethod -Uri "$baseUrl/api/authentication/validate" -Headers $adminHeader -TimeoutSec 10).valid } catch { $valid = $false }
    if (-not $valid) {
        $initialHeader = Get-BasicHeader 'admin'
        $initialValid = $false
        try { $initialValid = (Invoke-RestMethod -Uri "$baseUrl/api/authentication/validate" -Headers $initialHeader -TimeoutSec 10).valid } catch { $initialValid = $false }
        if (-not $initialValid) { throw 'No se pudo autenticar el administrador local. No se cambiaron credenciales existentes.' }
        Invoke-RestMethod -Method Post -Uri "$baseUrl/api/users/change_password" -Headers $initialHeader -Body @{
            login = 'admin'; previousPassword = 'admin'; password = $values['SONAR_ADMIN_PASSWORD']
        } -TimeoutSec 20 | Out-Null
    }

    $projectExists = $false
    try {
        Invoke-RestMethod -Uri "$baseUrl/api/components/show?component=bookingms-backend" -Headers $adminHeader -TimeoutSec 10 | Out-Null
        $projectExists = $true
    } catch {
        if ($_.Exception.Response.StatusCode -ne 404) { throw 'No se pudo comprobar el proyecto SonarQube.' }
    }
    if (-not $projectExists) {
        Invoke-RestMethod -Method Post -Uri "$baseUrl/api/projects/create" -Headers $adminHeader -Body @{
            project = 'bookingms-backend'; name = 'BookingMS Backend'; visibility = 'private'
        } -TimeoutSec 20 | Out-Null
    }

    $tokenValid = $false
    if ($values['SONAR_TOKEN']) {
        try {
            $tokenValid = (Invoke-RestMethod -Uri "$baseUrl/api/authentication/validate" -Headers @{
                Authorization = 'Bearer ' + $values['SONAR_TOKEN']
            } -TimeoutSec 10).valid
        } catch { $tokenValid = $false }
    }
    if (-not $tokenValid) {
        $tokenResult = Invoke-RestMethod -Method Post -Uri "$baseUrl/api/user_tokens/generate" -Headers $adminHeader -Body @{
            name = ('bookingms-local-' + [Guid]::NewGuid().ToString('N'))
            type = 'PROJECT_ANALYSIS_TOKEN'
            projectKey = 'bookingms-backend'
            expirationDate = [DateTime]::Today.AddDays(90).ToString('yyyy-MM-dd')
        } -TimeoutSec 20
        $lines = [System.IO.File]::ReadAllLines($environmentFile) | ForEach-Object {
            if ($_ -match '^SONAR_TOKEN=') { 'SONAR_TOKEN=' + $tokenResult.token } else { $_ }
        }
        [System.IO.File]::WriteAllLines($environmentFile, [string[]]$lines, $utf8)
    }
    Write-Host "Proyecto local preparado en $baseUrl/dashboard?id=bookingms-backend"
    Write-Host 'Credenciales locales guardadas en .quality/local.env (excluido de Git); no se muestran en consola.'
}

function Invoke-SonarAnalysis {
    Assert-DockerEngine
    $values = Read-LocalEnvironment
    if (-not $values['SONAR_TOKEN']) { throw 'Ejecuta Start y Setup antes de analizar con SonarQube.' }
    Write-CoverageReports
    $taskReport = Join-Path $reportDirectory 'sonar/report-task.txt'
    if (Test-Path -LiteralPath $taskReport) { Remove-Item -LiteralPath $taskReport }
    Invoke-Compose -DockerArguments @('run', '--rm', '--no-deps', 'scanner')
}

function Invoke-TrivyAnalysis {
    Assert-DockerEngine
    $reportPath = Join-Path $reportDirectory 'trivy.json'
    if (Test-Path -LiteralPath $reportPath) { Remove-Item -LiteralPath $reportPath }
    Invoke-Compose -DockerArguments @('run', '--rm', '--no-deps', 'trivy', 'fs', '--config', '/workspace/trivy.yaml',
        '--exit-code', '0', '--severity', 'UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL',
        '--ignorefile=', '--secret-config=', '--output', '/workspace/reports/trivy.json', '/workspace')
    if (-not (Test-Path -LiteralPath $reportPath)) { throw 'Trivy no genero el reporte esperado.' }
    $report = Get-Content -LiteralPath $reportPath -Raw | ConvertFrom-Json
    if (-not $report.SchemaVersion) { throw 'El reporte Trivy no tiene un formato valido.' }
    $blockingCount = 0
    $findingCount = 0
    foreach ($result in $report.Results) {
        foreach ($category in @('Vulnerabilities', 'Misconfigurations', 'Secrets')) {
            $property = $result.PSObject.Properties[$category]
            if ($null -eq $property) { continue }
            foreach ($finding in $property.Value) {
                $findingCount++
                if ($category -eq 'Secrets' -or $finding.Severity -in @('HIGH', 'CRITICAL')) { $blockingCount++ }
            }
        }
    }
    Write-Host "Trivy: $findingCount hallazgos; $blockingCount bloqueantes. Reporte: reports/trivy.json"
    if ($blockingCount -gt 0) { throw 'Trivy detecto hallazgos bloqueantes. Revisa el reporte; no se aplicaron exclusiones.' }
}

Push-Location $projectRoot
try {
    Initialize-LocalFiles
    switch ($Action) {
        'Init' { Write-Host 'Configuracion local inicializada sin sobrescribir credenciales existentes.' }
        'Validate' { Invoke-Compose -DockerArguments @('config', '--quiet') }
        'Start' {
            Assert-DockerEngine
            Invoke-Compose -DockerArguments @('up', '-d', '--wait', '--wait-timeout', '300', 'database', 'sonarqube')
        }
        'Setup' { Connect-LocalSonar }
        'Coverage' { Write-CoverageReports }
        'Sonar' { Invoke-SonarAnalysis }
        'Trivy' { Invoke-TrivyAnalysis }
        'Check' {
            # Run both analyzers even if one fails, but never return success on failure.
            $failures = @()
            try { Invoke-SonarAnalysis } catch { $failures += 'SonarQube'; Write-Warning $_.Exception.Message }
            try { Invoke-TrivyAnalysis } catch { $failures += 'Trivy'; Write-Warning $_.Exception.Message }
            if ($failures.Count -gt 0) { throw ('Verificaciones fallidas: ' + ($failures -join ', ')) }
        }
        'Status' { Invoke-Compose -DockerArguments @('ps') }
        'Stop' { Invoke-Compose -DockerArguments @('stop') }
    }
} finally { Pop-Location }
