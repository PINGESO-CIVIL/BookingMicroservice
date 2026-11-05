param(
    [Parameter(Mandatory = $true)][string]$ProjectRoot,
    [Parameter(Mandatory = $true)][string]$ModuleName,
    [Parameter(Mandatory = $true)][string]$InputPath,
    [Parameter(Mandatory = $true)][string]$OutputPath
)
$ErrorActionPreference = 'Stop'
$ProjectRoot = [IO.Path]::GetFullPath($ProjectRoot)
$tests = @{}
Get-ChildItem -LiteralPath (Join-Path $ProjectRoot 'Backend') -Filter '*_test.go' -Recurse | ForEach-Object {
    $relative = $_.FullName.Substring($ProjectRoot.Length + 1).Replace('\', '/')
    $package = $ModuleName
    if ($_.DirectoryName -ne (Join-Path $ProjectRoot 'Backend')) {
        $package += '/' + $relative.Substring(8, $relative.LastIndexOf('/') - 8)
    }
    foreach ($match in [regex]::Matches([IO.File]::ReadAllText($_.FullName), '(?m)^func\s+(Test\w+)\s*\(')) {
        $tests[$package + ':' + $match.Groups[1].Value] = $relative
    }
}
$records = foreach ($line in [IO.File]::ReadAllLines($InputPath)) {
    $event = $line | ConvertFrom-Json
    if ($event.Test -and $event.Action -in @('pass', 'fail', 'skip')) {
        $key = $event.Package + ':' + $event.Test.Split('/')[0]
        if (-not $tests.ContainsKey($key)) { throw "No se encontro el archivo de la prueba $key." }
        [pscustomobject]@{ File = $tests[$key]; Name = $event.Test; Action = $event.Action; Milliseconds = [long][Math]::Round($event.Elapsed * 1000) }
    }
}
$settings = [Xml.XmlWriterSettings]::new()
$settings.Indent = $true
$settings.Encoding = [Text.UTF8Encoding]::new($false)
$writer = [Xml.XmlWriter]::Create($OutputPath, $settings)
try {
    $writer.WriteStartElement('testExecutions')
    $writer.WriteAttributeString('version', '1')
    foreach ($group in ($records | Group-Object File)) {
        $writer.WriteStartElement('file')
        $writer.WriteAttributeString('path', $group.Name)
        foreach ($record in $group.Group) {
            $writer.WriteStartElement('testCase')
            $writer.WriteAttributeString('name', $record.Name)
            $writer.WriteAttributeString('duration', [string]$record.Milliseconds)
            if ($record.Action -eq 'skip') { $writer.WriteElementString('skipped', '') }
            if ($record.Action -eq 'fail') { $writer.WriteElementString('failure', 'La prueba Go fallo; consulta go-tests.json.') }
            $writer.WriteEndElement()
        }
        $writer.WriteEndElement()
    }
    $writer.WriteEndElement()
} finally { $writer.Dispose() }
