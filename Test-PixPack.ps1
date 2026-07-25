#Requires -Version 5.1

[CmdletBinding()]
param(
    [switch]$SkipGoChecks,
    [switch]$KeepArtifacts
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = (Get-Location).Path
$WorkDir = Join-Path $RepoRoot '.pixpack-e2e'
$Binary = Join-Path $WorkDir 'pixpack.exe'
$Results = New-Object 'System.Collections.Generic.List[object]'
$SuitePassed = $false

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ''
    Write-Host ('=' * 68) -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host ('=' * 68) -ForegroundColor Cyan
}

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [int[]]$ExpectedExitCodes = @(0),
        [string]$DisplayName = ''
    )

    if ([string]::IsNullOrWhiteSpace($DisplayName)) {
        $DisplayName = $FilePath
    }

    Write-Host ''
    Write-Host ('> ' + $DisplayName + ' ' + ($Arguments -join ' ')) -ForegroundColor DarkGray

    $CommandOutput = @(& $FilePath @Arguments 2>&1)
    $ExitCode = $LASTEXITCODE

    foreach ($Line in $CommandOutput) {
        Write-Host $Line
    }

    if ($ExpectedExitCodes -notcontains $ExitCode) {
        $ExpectedText = $ExpectedExitCodes -join ', '
        $OutputText = $CommandOutput -join [Environment]::NewLine
        throw "Unexpected exit code from $DisplayName. Expected [$ExpectedText], got [$ExitCode].`nOutput:`n$OutputText"
    }

    return [PSCustomObject]@{
        ExitCode = $ExitCode
        Output = $CommandOutput
    }
}

function Invoke-Go {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)

    $null = Invoke-CheckedCommand `
        -FilePath 'go' `
        -Arguments $Arguments `
        -DisplayName 'go'
}

function Invoke-PixPack {
    param(
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [int[]]$ExpectedExitCodes = @(0)
    )

    return Invoke-CheckedCommand `
        -FilePath $Binary `
        -Arguments $Arguments `
        -ExpectedExitCodes $ExpectedExitCodes `
        -DisplayName 'pixpack'
}

function Assert-FileExists {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Description
    )

    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) {
        throw "$Description does not exist: $Path"
    }
}

function Compare-FilesExactly {
    param(
        [Parameter(Mandatory = $true)][string]$OriginalPath,
        [Parameter(Mandatory = $true)][string]$RestoredPath
    )

    Assert-FileExists -Path $OriginalPath -Description 'Original file'
    Assert-FileExists -Path $RestoredPath -Description 'Restored file'

    $OriginalItem = Get-Item -LiteralPath $OriginalPath
    $RestoredItem = Get-Item -LiteralPath $RestoredPath

    if ($OriginalItem.Length -ne $RestoredItem.Length) {
        throw "File-size mismatch. Original=$($OriginalItem.Length), restored=$($RestoredItem.Length)."
    }

    $OriginalHash = (Get-FileHash -LiteralPath $OriginalPath -Algorithm SHA256).Hash
    $RestoredHash = (Get-FileHash -LiteralPath $RestoredPath -Algorithm SHA256).Hash

    if ($OriginalHash -ne $RestoredHash) {
        throw "SHA-256 mismatch.`nOriginal: $OriginalHash`nRestored: $RestoredHash"
    }

    $FcExe = Join-Path $env:SystemRoot 'System32\fc.exe'
    & $FcExe /b $OriginalPath $RestoredPath | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw 'fc.exe reported byte differences.'
    }

    return [PSCustomObject]@{
        Bytes = $OriginalItem.Length
        SHA256 = $OriginalHash
    }
}

function Test-RoundTrip {
    param(
        [Parameter(Mandatory = $true)][string]$InputPath,
        [Parameter(Mandatory = $true)][string]$Label,
        [switch]$CheckOverwriteProtection
    )

    Write-Step "ROUND TRIP: $Label"

    $ResolvedInput = (Resolve-Path -LiteralPath $InputPath).Path
    $Extension = [System.IO.Path]::GetExtension($ResolvedInput)
    if ([string]::IsNullOrEmpty($Extension)) {
        $Extension = '.bin'
    }

    $EncodedPath = Join-Path $WorkDir ($Label + '.png')
    $RestoredPath = Join-Path $WorkDir ('restored-' + $Label + $Extension)

    # Go's standard flag parser expects flags before positional arguments.
    $null = Invoke-PixPack -Arguments @(
        'encode',
        '--overwrite',
        $ResolvedInput,
        $EncodedPath
    )

    Assert-FileExists -Path $EncodedPath -Description 'Encoded PNG'

    $null = Invoke-PixPack -Arguments @('inspect', $EncodedPath)
    $null = Invoke-PixPack -Arguments @('verify', $EncodedPath)

    $null = Invoke-PixPack -Arguments @(
        'decode',
        '--output',
        $RestoredPath,
        '--overwrite',
        $EncodedPath
    )

    Assert-FileExists -Path $RestoredPath -Description 'Decoded output'

    $Comparison = Compare-FilesExactly `
        -OriginalPath $ResolvedInput `
        -RestoredPath $RestoredPath

    $PngBytes = (Get-Item -LiteralPath $EncodedPath).Length

    $null = $Results.Add([PSCustomObject]@{
        Test = $Label
        OriginalBytes = $Comparison.Bytes
        PngBytes = $PngBytes
        SHA256 = $Comparison.SHA256.Substring(0, 16) + '...'
        Result = 'PASS'
    })

    Write-Host ''
    Write-Host "PASS: $Label restored byte-for-byte." -ForegroundColor Green
    Write-Host "Original: $($Comparison.Bytes) bytes"
    Write-Host "PNG:      $PngBytes bytes"
    Write-Host "SHA-256:  $($Comparison.SHA256)"

    if ($CheckOverwriteProtection) {
        Write-Step 'OVERWRITE PROTECTION'

        $null = Invoke-PixPack `
            -Arguments @('encode', $ResolvedInput, $EncodedPath) `
            -ExpectedExitCodes @(1)

        $null = Invoke-PixPack `
            -Arguments @('decode', '--output', $RestoredPath, $EncodedPath) `
            -ExpectedExitCodes @(1)

        Write-Host ''
        Write-Host 'PASS: Existing outputs were protected.' -ForegroundColor Green
    }

    return [PSCustomObject]@{
        EncodedPath = $EncodedPath
        RestoredPath = $RestoredPath
    }
}

try {
    Write-Step 'VALIDATING REPOSITORY'

    if (-not (Test-Path -LiteralPath (Join-Path $RepoRoot 'go.mod') -PathType Leaf)) {
        throw 'go.mod was not found. Run this script from the PixPack repository root.'
    }

    if (-not (Test-Path -LiteralPath (Join-Path $RepoRoot 'cmd\pixpack') -PathType Container)) {
        throw 'cmd\pixpack was not found. Run this script from the PixPack repository root.'
    }

    if (Test-Path -LiteralPath $WorkDir) {
        Remove-Item -LiteralPath $WorkDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $WorkDir | Out-Null

    if (-not $SkipGoChecks) {
        Write-Step 'GO CHECKS'
        Invoke-Go -Arguments @('clean', '-testcache')
        Invoke-Go -Arguments @('fmt', './...')
        Invoke-Go -Arguments @('vet', './...')
        Invoke-Go -Arguments @('test', '-count=1', './...')
        Invoke-Go -Arguments @('test', '-race', '-count=1', './...')
    }

    Write-Step 'BUILDING PIXPACK'
    Invoke-Go -Arguments @(
        'build',
        '-trimpath',
        '-o',
        $Binary,
        './cmd/pixpack'
    )

    Assert-FileExists -Path $Binary -Description 'PixPack executable'
    $null = Invoke-PixPack -Arguments @('--version')
    $null = Invoke-PixPack -Arguments @('--help')

    Write-Step 'CREATING TEST INPUTS'

    $InputsDir = Join-Path $WorkDir 'inputs'
    New-Item -ItemType Directory -Path $InputsDir | Out-Null

    $EmptyFile = Join-Path $InputsDir 'empty.bin'
    [System.IO.File]::WriteAllBytes($EmptyFile, [byte[]]@())

    $OneByteFile = Join-Path $InputsDir 'one-byte.bin'
    [System.IO.File]::WriteAllBytes($OneByteFile, [byte[]](0x41))

    $TwoByteFile = Join-Path $InputsDir 'two-byte.bin'
    [System.IO.File]::WriteAllBytes($TwoByteFile, [byte[]](0x41, 0x42))

    $ThreeByteFile = Join-Path $InputsDir 'three-byte.bin'
    [System.IO.File]::WriteAllBytes($ThreeByteFile, [byte[]](0x41, 0x42, 0x43))

    $FourByteFile = Join-Path $InputsDir 'four-byte.bin'
    [System.IO.File]::WriteAllBytes($FourByteFile, [byte[]](0x41, 0x42, 0x43, 0x44))

    $UnicodeFile = Join-Path $InputsDir 'unicode-测试.txt'
    $UnicodeText = @'
PixPack Unicode test
English: Hello
Hindi: नमस्ते
Punjabi: ਸਤ ਸ੍ਰੀ ਅਕਾਲ
Chinese: 你好
Emoji: 😭🔥💸
'@
    [System.IO.File]::WriteAllText($UnicodeFile, $UnicodeText, [System.Text.Encoding]::UTF8)

    $RandomFile = Join-Path $InputsDir 'random-1mb-plus-7.bin'
    $RandomBytes = New-Object byte[] ((1024 * 1024) + 7)
    $Rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $Rng.GetBytes($RandomBytes)
    }
    finally {
        $Rng.Dispose()
    }
    [System.IO.File]::WriteAllBytes($RandomFile, $RandomBytes)

    $ZipSourceDir = Join-Path $InputsDir 'zip-source'
    New-Item -ItemType Directory -Path $ZipSourceDir | Out-Null
    [System.IO.File]::WriteAllText(
        (Join-Path $ZipSourceDir 'hello.txt'),
        'Hello from inside a PixPack ZIP.',
        [System.Text.Encoding]::UTF8
    )
    Copy-Item -LiteralPath $FourByteFile -Destination (Join-Path $ZipSourceDir 'tiny.bin')

    $ZipFile = Join-Path $InputsDir 'sample.zip'
    Compress-Archive -Path (Join-Path $ZipSourceDir '*') -DestinationPath $ZipFile -Force

    $TestInputs = @(
        @{ Label = 'empty'; Path = $EmptyFile },
        @{ Label = 'one-byte'; Path = $OneByteFile },
        @{ Label = 'two-byte'; Path = $TwoByteFile },
        @{ Label = 'three-byte'; Path = $ThreeByteFile },
        @{ Label = 'four-byte'; Path = $FourByteFile },
        @{ Label = 'unicode'; Path = $UnicodeFile },
        @{ Label = 'zip'; Path = $ZipFile },
        @{ Label = 'random-1mb'; Path = $RandomFile }
    )

    $GoSource = Join-Path $RepoRoot 'internal\codec\encode.go'
    if (Test-Path -LiteralPath $GoSource -PathType Leaf) {
        $TestInputs += @{ Label = 'go-source'; Path = $GoSource }
    }

    $FirstTest = $true
    $RandomRoundTrip = $null

    foreach ($TestInput in $TestInputs) {
        $RoundTrip = Test-RoundTrip `
            -InputPath $TestInput.Path `
            -Label $TestInput.Label `
            -CheckOverwriteProtection:$FirstTest

        $FirstTest = $false

        if ($TestInput.Label -eq 'random-1mb') {
            $RandomRoundTrip = $RoundTrip
        }
    }

    if ($null -eq $RandomRoundTrip) {
        throw 'The random-file round-trip result was not recorded.'
    }

    Write-Step 'BUILDING PNG MUTATION HELPER'

    $HelperPath = Join-Path $WorkDir 'png-test-helper.go'
    $HelperSource = @'
package main

import (
    "fmt"
    "image"
    "image/draw"
    "image/png"
    "os"
)

func fail(format string, args ...any) {
    fmt.Fprintf(os.Stderr, format+"\n", args...)
    os.Exit(1)
}

func savePNG(path string, img image.Image) {
    file, err := os.Create(path)
    if err != nil {
        fail("create output: %v", err)
    }

    if err := png.Encode(file, img); err != nil {
        _ = file.Close()
        fail("encode PNG: %v", err)
    }

    if err := file.Close(); err != nil {
        fail("close output: %v", err)
    }
}

func corrupt(inputPath, outputPath string) {
    file, err := os.Open(inputPath)
    if err != nil {
        fail("open input: %v", err)
    }

    img, err := png.Decode(file)
    _ = file.Close()
    if err != nil {
        fail("decode input PNG: %v", err)
    }

    bounds := img.Bounds()
    converted := image.NewNRGBA(bounds)
    draw.Draw(converted, bounds, img, bounds.Min, draw.Src)

    totalPixels := bounds.Dx() * bounds.Dy()
    if totalPixels == 0 {
        fail("image has no pixels")
    }

    index := 100
    if index >= totalPixels {
        index = totalPixels / 2
    }

    x := bounds.Min.X + index%bounds.Dx()
    y := bounds.Min.Y + index/bounds.Dx()
    pixel := converted.NRGBAAt(x, y)
    pixel.R ^= 1
    converted.SetNRGBA(x, y, pixel)

    savePNG(outputPath, converted)
}

func ordinary(outputPath string) {
    img := image.NewNRGBA(image.Rect(0, 0, 16, 16))

    for y := 0; y < 16; y++ {
        for x := 0; x < 16; x++ {
            offset := img.PixOffset(x, y)
            img.Pix[offset+0] = uint8(x * 16)
            img.Pix[offset+1] = uint8(y * 16)
            img.Pix[offset+2] = uint8((x + y) * 8)
            img.Pix[offset+3] = 255
        }
    }

    savePNG(outputPath, img)
}

func main() {
    if len(os.Args) < 3 {
        fail("usage: helper <corrupt|ordinary> [input] <output>")
    }

    switch os.Args[1] {
    case "corrupt":
        if len(os.Args) != 4 {
            fail("usage: helper corrupt <input.png> <output.png>")
        }
        corrupt(os.Args[2], os.Args[3])
    case "ordinary":
        if len(os.Args) != 3 {
            fail("usage: helper ordinary <output.png>")
        }
        ordinary(os.Args[2])
    default:
        fail("unknown mode: %s", os.Args[1])
    }
}
'@
    [System.IO.File]::WriteAllText($HelperPath, $HelperSource, [System.Text.Encoding]::UTF8)

    Write-Step 'CORRUPTION DETECTION'

    $CorruptedPng = Join-Path $WorkDir 'corrupted-payload.png'
    Invoke-Go -Arguments @(
        'run',
        $HelperPath,
        'corrupt',
        $RandomRoundTrip.EncodedPath,
        $CorruptedPng
    )

    Assert-FileExists -Path $CorruptedPng -Description 'Corrupted PNG'
    $null = Invoke-PixPack -Arguments @('verify', $CorruptedPng) -ExpectedExitCodes @(4)
    Write-Host ''
    Write-Host 'PASS: A changed payload pixel caused checksum failure.' -ForegroundColor Green

    Write-Step 'ORDINARY PNG REJECTION'

    $OrdinaryPng = Join-Path $WorkDir 'ordinary.png'
    Invoke-Go -Arguments @('run', $HelperPath, 'ordinary', $OrdinaryPng)

    Assert-FileExists -Path $OrdinaryPng -Description 'Ordinary PNG'
    $null = Invoke-PixPack -Arguments @('inspect', $OrdinaryPng) -ExpectedExitCodes @(3)
    $null = Invoke-PixPack -Arguments @('verify', $OrdinaryPng) -ExpectedExitCodes @(3)
    Write-Host ''
    Write-Host 'PASS: An ordinary PNG was rejected.' -ForegroundColor Green

    Write-Step 'RESULTS'

    $Results | Format-Table Test, OriginalBytes, PngBytes, SHA256, Result -AutoSize | Out-Host

    Write-Host ''
    Write-Host 'ALL PIXPACK TESTS PASSED!' -ForegroundColor Green
    Write-Host ('Artifacts directory: ' + $WorkDir)

    $SuitePassed = $true
}
catch {
    Write-Host ''
    Write-Host 'PIXPACK TEST SUITE FAILED' -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ('Artifacts preserved at: ' + $WorkDir)
    exit 1
}
finally {
    if ($SuitePassed -and -not $KeepArtifacts) {
        Write-Host ''
        Write-Host 'Successful test artifacts were kept for inspection.' -ForegroundColor DarkGray
        Write-Host 'Run with -KeepArtifacts for the same behavior; delete .pixpack-e2e manually when done.' -ForegroundColor DarkGray
    }
}

exit 0
