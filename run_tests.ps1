#!/usr/bin/env pwsh
# Run all tests with coverage and display results

$coverDir = "coverage"
$coverFile = "$coverDir\coverage"

if (!(Test-Path $coverDir)) {
    New-Item -ItemType Directory -Path $coverDir | Out-Null
}

Write-Host "`n=== Running all tests with coverage ===" -ForegroundColor Cyan
go test ./... -coverprofile=$coverFile -count=1

if ($LASTEXITCODE -ne 0) {
    Write-Host "`nSome tests FAILED." -ForegroundColor Red
} else {
    Write-Host "`nAll tests PASSED." -ForegroundColor Green
}

Write-Host "`n=== Per-package coverage ===" -ForegroundColor Cyan
go tool cover -func $coverFile | Select-String -Pattern "^pickup/" | ForEach-Object {
    $line = $_.ToString()
    if ($line -match '(\S+)\s+(\d+\.\d+)%$') {
        $pct = [double]$Matches[2]
        $color = if ($pct -ge 80) { "Green" } elseif ($pct -ge 50) { "Yellow" } else { "Red" }
        Write-Host $line -ForegroundColor $color
    }
}

Write-Host "`n=== Overall coverage ===" -ForegroundColor Cyan
$total = go tool cover -func $coverFile | Select-String "total:"
Write-Host $total.ToString() -ForegroundColor Magenta
