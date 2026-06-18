# Manual smoke tests for API endpoints.
$ErrorActionPreference = "Stop"
$Base = "http://127.0.0.1:8090"
$Failed = 0
$Passed = 0

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method = "GET",
        [string]$Url,
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [int[]]$ExpectStatus = @(200),
        [scriptblock]$Assert = $null
    )
    try {
        $params = @{
            Uri             = $Url
            Method          = $Method
            Headers         = $Headers
            UseBasicParsing = $true
        }
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        $resp = Invoke-WebRequest @params
        $status = [int]$resp.StatusCode
        if ($ExpectStatus -notcontains $status) {
            Write-Host "FAIL $Name - status $status, expected $($ExpectStatus -join ',')" -ForegroundColor Red
            Write-Host $resp.Content
            $script:Failed++
            return $null
        }
        $json = $null
        if ($resp.Content) {
            try { $json = $resp.Content | ConvertFrom-Json } catch {}
        }
        if ($Assert) {
            & $Assert $json $resp
        }
        Write-Host "OK   $Name ($status)" -ForegroundColor Green
        $script:Passed++
        return $json
    } catch {
        $status = 0
        $content = $_.Exception.Message
        if ($_.Exception.Response) {
            $status = [int]$_.Exception.Response.StatusCode
            try {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $content = $reader.ReadToEnd()
            } catch {}
        }
        if ($ExpectStatus -contains $status) {
            Write-Host "OK   $Name ($status)" -ForegroundColor Green
            $script:Passed++
            return $null
        }
        Write-Host "FAIL $Name - status $status" -ForegroundColor Red
        Write-Host $content
        $script:Failed++
        return $null
    }
}

Write-Host "`n=== Public / App links ===" -ForegroundColor Cyan
Test-Endpoint "GET /v1/app/links" -Url "$Base/v1/app/links" -Assert {
    param($j)
    if (-not $j.bot_username) { throw "missing bot_username" }
}

Test-Endpoint "GET /v1/users/telegram/777000/referral-link" -Url "$Base/v1/users/telegram/777000/referral-link" -Assert {
    param($j)
    if ($j.referral_link -notmatch "startapp=ref_") { throw "bad referral_link" }
}

Write-Host "`n=== Admin auth ===" -ForegroundColor Cyan
$login = Test-Endpoint "POST /api/admin/auth/login" -Method POST -Url "$Base/api/admin/auth/login" `
    -Body '{"email":"admin@test.com","password":"testpass123"}' -Assert {
    param($j)
    if (-not $j.token) { throw "missing token" }
    if (-not $j.admin.email) { throw "missing admin.email" }
}
$token = $login.token
$auth = @{ Authorization = "Bearer $token" }

Test-Endpoint "GET /api/admin/auth/me" -Url "$Base/api/admin/auth/me" -Headers $auth -Assert {
    param($j) if (-not $j.email) { throw "missing email" }
}

Test-Endpoint "POST /api/admin/auth/logout" -Method POST -Url "$Base/api/admin/auth/logout" -Headers $auth

Write-Host "`n=== Admin stats & users ===" -ForegroundColor Cyan
Test-Endpoint "GET /api/admin/stats" -Url "$Base/api/admin/stats" -Headers $auth -Assert {
    param($j)
    if ($null -eq $j.total_users) { throw "missing total_users" }
}

$users = Test-Endpoint "GET /api/admin/users" -Url "$Base/api/admin/users?page=1&per_page=20" -Headers $auth -Assert {
    param($j)
    if ($null -eq $j.data) { throw "missing data array" }
    if ($null -eq $j.total) { throw "missing total" }
}

if ($users.data.Count -gt 0) {
    $uid = $users.data[0].id
    Test-Endpoint "GET /api/admin/users/:id" -Url "$Base/api/admin/users/$uid" -Headers $auth
    $patched = Test-Endpoint "PATCH /api/admin/users/:id" -Method PATCH -Url "$Base/api/admin/users/$uid" `
        -Headers $auth -Body '{"is_active":false}' -Assert {
        param($j) if ($j.is_active -ne $false) { throw "is_active not false" }
    }
    Test-Endpoint "PATCH restore active" -Method PATCH -Url "$Base/api/admin/users/$uid" `
        -Headers $auth -Body '{"is_active":true}'
}

Test-Endpoint "GET /api/admin/users search" -Url "$Base/api/admin/users?search=test" -Headers $auth -Assert {
    param($j) if ($null -eq $j.data) { throw "missing data" }
}

Write-Host "`n=== Referrals ===" -ForegroundColor Cyan
Test-Endpoint "GET /v1/users/telegram/777000/referrals" -Url "$Base/v1/users/telegram/777000/referrals" `
    -ExpectStatus @(200, 404)

Write-Host "`n=== Register (referral link) ===" -ForegroundColor Cyan
$userJson = '{"id":999001,"first_name":"Test","username":"testref"}'
$initRaw = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("user=$userJson"))
$regBody = @{ init_data_raw = $initRaw; start_param = "ref_777000" } | ConvertTo-Json
Test-Endpoint "POST /v1/register" -Method POST -Url "$Base/v1/register" -Body $regBody `
    -ExpectStatus @(200, 409)

Write-Host "`n=== Broadcast (bot disabled -> 503 expected) ===" -ForegroundColor Cyan
Test-Endpoint "POST /api/admin/broadcast" -Method POST -Url "$Base/api/admin/broadcast" -Headers $auth `
    -Body '{"message":"test","target":"all","parse_mode":"HTML"}' -ExpectStatus @(200, 503)

Write-Host "`n=== Unauthorized ===" -ForegroundColor Cyan
Test-Endpoint "GET /api/admin/stats no token" -Url "$Base/api/admin/stats" -ExpectStatus @(401)

Write-Host "`n--- Results: $Passed passed, $Failed failed ---`n" -ForegroundColor $(if ($Failed -eq 0) { "Green" } else { "Red" })
if ($Failed -gt 0) { exit 1 }
