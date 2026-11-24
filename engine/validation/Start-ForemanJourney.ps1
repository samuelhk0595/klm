param(
    [Parameter(Mandatory = $true)][string]$RunRoot,
    [string]$BrowserOrigin = 'http://127.0.0.1:5174',
    [ValidateRange(1, 30)][int]$Minutes = 15,
    [string]$KlmBinary = 'C:\Users\Samuel\AppData\Local\Temp\opencode\klm-factory-backend-check.exe',
    [string]$WerbertBinary = 'C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe'
)

# Run in a separate powershell.exe -File process. No servers start until invoked.
$ErrorActionPreference = 'Stop'
$scratch = 'C:\Users\Samuel\AppData\Local\Temp\opencode'
$RunRoot = [IO.Path]::GetFullPath($RunRoot).TrimEnd('\')
if ([IO.Path]::GetDirectoryName($RunRoot) -ne $scratch -or (Test-Path -LiteralPath $RunRoot)) {
    throw 'RunRoot must be a new direct child of the approved temporary directory.'
}
if (-not (Test-Path -LiteralPath $scratch -PathType Container)) { throw 'Temporary parent missing.' }
$origin = [Uri]$BrowserOrigin
if ($origin.Scheme -ne 'http' -or $origin.Host -notin @('127.0.0.1', 'localhost') -or
    $BrowserOrigin -ne $origin.GetLeftPart([UriPartial]::Authority)) {
    throw 'BrowserOrigin must be an exact local HTTP origin without a trailing slash.'
}
$originalAuth = Join-Path $env:USERPROFILE '.pi\agent\auth.json'
$pi = Join-Path $env:APPDATA 'npm\node_modules\@earendil-works\pi-coding-agent\dist\cli.js'
foreach ($file in @($originalAuth, $pi, $KlmBinary, $WerbertBinary)) {
    if (-not (Test-Path -LiteralPath $file -PathType Leaf)) { throw "Required file unavailable: $file" }
}
foreach ($port in @(17331, 18090)) {
    $reservation = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, $port)
    try { $reservation.Start() } finally { $reservation.Stop() }
}

$null = New-Item -ItemType Directory -Path $RunRoot
$sid = [Security.Principal.WindowsIdentity]::GetCurrent().User
$acl = [Security.AccessControl.DirectorySecurity]::new()
$acl.SetOwner($sid)
$acl.SetAccessRuleProtection($true, $false)
$rule = [Security.AccessControl.FileSystemAccessRule]::new($sid, 'FullControl', 'ContainerInherit,ObjectInherit', 'None', 'Allow')
$acl.AddAccessRule($rule)
Set-Acl -LiteralPath $RunRoot -AclObject $acl
$engineDir = Join-Path $RunRoot 'klm'
$nativeDir = Join-Path $RunRoot 'pi'
$projectDir = Join-Path $RunRoot 'project'
$werbertDir = Join-Path $RunRoot 'werbert'
$agentDir = Join-Path $projectDir '.klm\agents'
foreach ($dir in @($engineDir, $nativeDir, $projectDir, $werbertDir, $agentDir,
    (Join-Path $RunRoot 'home'), (Join-Path $RunRoot 'roaming'),
    (Join-Path $RunRoot 'local'), (Join-Path $RunRoot 'skills'))) {
    $null = New-Item -ItemType Directory -Path $dir -Force
}
$utf8 = [Text.UTF8Encoding]::new($false)
function Save-JSON([string]$Path, $Value) {
    [IO.File]::WriteAllText($Path, ($Value | ConvertTo-Json -Depth 20), $utf8)
}
function Control([string]$Command) {
    $sha = [Security.Cryptography.SHA256]::Create()
    try { $hash = $sha.ComputeHash([Text.Encoding]::UTF8.GetBytes($engineDir.ToLowerInvariant())) }
    finally { $sha.Dispose() }
    $suffix = [BitConverter]::ToString($hash, 0, 8).Replace('-', '').ToLowerInvariant()
    $pipe = [IO.Pipes.NamedPipeClientStream]::new('.', "klm-engine-$sid-$suffix", [IO.Pipes.PipeDirection]::InOut)
    try {
        $pipe.Connect(1000)
        $writer = [IO.StreamWriter]::new($pipe, $utf8, 1024, $true)
        $writer.NewLine = "`n"
        $reader = [IO.StreamReader]::new($pipe, $utf8, $false, 1024, $true)
        try {
            $writer.WriteLine($Command)
            $writer.Flush()
            $read = $reader.ReadLineAsync()
            if (-not $read.Wait(3500)) { throw 'Owned control response timed out.' }
            return $read.Result
        } finally { $reader.Dispose(); $writer.Dispose() }
    } finally { $pipe.Dispose() }
}
function Klm-JSON([string]$Method, [string]$Path, $Body, [string]$Credential = '') {
    $headers = @{ 'X-KLM-Native' = '1'; 'X-KLM-Read-Contract' = '3' }
    if ($Credential) { $headers['Authorization'] = "Bearer $Credential" }
    $args = @{ Method = $Method; Uri = "http://127.0.0.1:17331$Path"; Headers = $headers; TimeoutSec = 10 }
    if ($Method -notin @('GET', 'HEAD')) { $args.ContentType = 'application/json' }
    if ($null -ne $Body) { $args.Body = $Body | ConvertTo-Json -Depth 20 }
    # Do not propagate a response/error body that might contain pairing material.
    try { return Invoke-RestMethod @args } catch {
        $status = 'transport'
        if ($null -ne $_.Exception.Response) { $status = [string][int]$_.Exception.Response.StatusCode }
        throw "Disposable KLM $Method $Path failed (status $status)."
    }
}

$klm = $null
$werbert = $null
$cleanup = [ordered]@{ klm_exit_observed = $false; werbert_exit_observed = $false; native_auth_removed = $false; original_auth_unchanged = $false }
$authHash = (Get-FileHash -LiteralPath $originalAuth -Algorithm SHA256).Hash
$authCopy = Join-Path $nativeDir 'auth.json'
$agentPath = Join-Path $agentDir 'journey-worker.toml'
$deadline = [DateTime]::UtcNow.AddMinutes($Minutes)
try {
    [IO.File]::WriteAllBytes($authCopy, [IO.File]::ReadAllBytes($originalAuth))
    $sourceUID = [Guid]::NewGuid().ToString('N')
    $workerPrompt = [IO.File]::ReadAllText((Join-Path $PSScriptRoot 'foreman-journey-agent.md'))
    if ($workerPrompt.Contains("'''")) { throw 'Fixture prompt cannot contain a TOML literal delimiter.' }
    $agent = "source_uid = '$sourceUID'`nname = 'Journey Worker'`nenabled = true`ndefault_harness = 'pi'`nmodel = 'openai-codex/gpt-6-astra'`neffort = 'low'`nskills = []`nmcp_connections = []`nprompt = '''`n$workerPrompt`n'''`n"
    [IO.File]::WriteAllText($agentPath, $agent, $utf8)
    $agentHash = (Get-FileHash -LiteralPath $agentPath -Algorithm SHA256).Hash

    # Child-only environment: preserve executable PATH but discard owning-agent controls.
    foreach ($entry in @(Get-ChildItem Env:)) {
        if ($entry.Name -match '^(KLM_|WERBERT_|OPENCODE_|CODEX_|PI_|OPENAI_|ANTHROPIC_|XDG_)') {
            [Environment]::SetEnvironmentVariable($entry.Name, $null, 'Process')
        }
    }
    $env:HOME = Join-Path $RunRoot 'home'
    $env:USERPROFILE = $env:HOME
    $env:APPDATA = Join-Path $RunRoot 'roaming'
    $env:LOCALAPPDATA = Join-Path $RunRoot 'local'
    $env:TEMP = $RunRoot
    $env:TMP = $RunRoot
    $env:PI_CODING_AGENT_DIR = $nativeDir
    $env:KLM_PI_BIN = $pi
    $env:KLM_OPENCODE_BIN = Join-Path $RunRoot 'unavailable-opencode.exe'
    $env:KLM_CODEX_BIN = Join-Path $RunRoot 'unavailable-codex.exe'
    $env:KLM_USER_SKILLS_DIR = Join-Path $RunRoot 'skills'
    $env:KLM_DEV_DATA_DIR = $engineDir
    $env:KLM_API_BIND_HOST = '127.0.0.1'
    # KLM_DEV_DISABLE_CLIENT_AUTH is intentionally absent: pairing is exercised.
    $klm = Start-Process -FilePath $KlmBinary -ArgumentList '__worker' -WorkingDirectory $RunRoot -PassThru -WindowStyle Hidden -RedirectStandardOutput (Join-Path $RunRoot 'klm.out.log') -RedirectStandardError (Join-Path $RunRoot 'klm.err.log')
    $null = $klm.Handle
    $startupDeadline = [DateTime]::UtcNow.AddSeconds(60)
    $discovery = $null
    $discoveryAttempts = 0
    $discoveryDiagnostic = 'Protected control has not responded yet.'
    while ([DateTime]::UtcNow -lt $startupDeadline) {
        if ($klm.HasExited) { throw 'Disposable KLM exited; inspect its owned logs.' }
        $discoveryAttempts++
        try {
            $line = Control 'discover'
            if ($line -eq 'Unknown control request.') {
                $discoveryDiagnostic = 'Protected control rejected discover; verify LF framing and protocol support.'
            } else {
                try {
                    $candidate = $line | ConvertFrom-Json -ErrorAction Stop
                    $discoveryDiagnostic = "Protected discovery returned ready=$($candidate.ready), api_port=$($candidate.api_port), protocol=$($candidate.runner_protocol)."
                    if ($candidate.ready) { $discovery = $candidate; break }
                } catch { $discoveryDiagnostic = 'Protected discovery returned invalid JSON; verify framing/protocol.' }
            }
        } catch { $discoveryDiagnostic = "Protected pipe connect/read failed ($($_.Exception.GetType().Name)); inspect the owned KLM log." }
        Start-Sleep -Milliseconds 250
    }
    Save-JSON (Join-Path $RunRoot 'discovery-diagnostics.json') @{
        attempts = $discoveryAttempts; diagnostic = $discoveryDiagnostic; framing = 'LF'
    }
    if ($null -eq $discovery -or $discovery.api_port -ne '17331' -or $discovery.runner_protocol -ne '1') {
        throw "Disposable KLM protected discovery did not become compatible/ready. $discoveryDiagnostic"
    }
    $unauthenticatedStatus = 0
    $setupError = $null
    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri 'http://127.0.0.1:17331/api/state' -TimeoutSec 5
        $unauthenticatedStatus = [int]$response.StatusCode
    } catch {
        if ($null -ne $_.Exception.Response) { $unauthenticatedStatus = [int]$_.Exception.Response.StatusCode }
    }
    if ($unauthenticatedStatus -ne 401) { throw 'Disposable KLM did not enforce paired-client authentication.' }
    # Pair a short-lived setup client only to register the disposable project.
    # The browser still exercises its own explicit Werbert Connect afterward.
    $pair = Control 'pair'
    if (-not $pair.StartsWith('pair ')) { throw 'Disposable setup pairing unavailable.' }
    $paired = Klm-JSON 'POST' '/api/auth/pair' @{ code = $pair.Substring(5); name = 'Disposable journey setup' }
    try {
        $state = Klm-JSON 'GET' '/api/state' $null $paired.credential
        $availablePi = @($state.harnesses | Where-Object { $_.id -eq 'pi' -and $_.available -and $_.compatible -and $_.installedVersion -eq '0.85.1' })
        if ($availablePi.Count -ne 1) { throw 'Pinned Pi is not ready; do not install or substitute a harness.' }
        $project = Klm-JSON 'POST' '/api/projects' @{ name = 'Disposable Foreman journey'; folder = $projectDir; icon = '' } $paired.credential
    } catch { $setupError = $_ } finally {
        try { $null = Klm-JSON 'DELETE' ("/api/auth/devices/" + $paired.clientId) $null $paired.credential }
        catch {
            if ($null -eq $setupError) { $setupError = $_ }
            else { $cleanup.setup_revoke_error = 'Setup-client revocation also failed; original setup error retained.' }
        }
        $paired = $null
        $pair = $null
    }
    if ($null -ne $setupError) { throw $setupError }

    $env:WERBERT_MODE = 'foreman'
    $env:WERBERT_LISTEN = '127.0.0.1:18090'
    $env:WERBERT_DATA_ROOT = $werbertDir
    $env:WERBERT_ORIGINS = $BrowserOrigin
    $env:WERBERT_HOSTS = '127.0.0.1:18090'
    $env:WERBERT_KLM_PORT = '17331'
    $env:WERBERT_KLM_DATA_DIR = $engineDir
    $werbert = Start-Process -FilePath $WerbertBinary -WorkingDirectory $RunRoot -PassThru -WindowStyle Hidden -RedirectStandardOutput (Join-Path $RunRoot 'werbert.out.log') -RedirectStandardError (Join-Path $RunRoot 'werbert.err.log')
    $null = $werbert.Handle
    $startupDeadline = [DateTime]::UtcNow.AddSeconds(20)
    $healthy = $false
    while ([DateTime]::UtcNow -lt $startupDeadline) {
        if ($werbert.HasExited) { throw 'Disposable Werbert exited; inspect its owned logs.' }
        try {
            $health = Invoke-RestMethod -Uri 'http://127.0.0.1:18090/healthz' -TimeoutSec 2
            if ($health.mode -eq 'foreman') { $healthy = $true; break }
        } catch { }
        Start-Sleep -Milliseconds 250
    }
    if (-not $healthy) { throw 'Disposable Werbert did not become ready.' }
    Save-JSON (Join-Path $RunRoot 'ready.json') ([ordered]@{
        engine_id = $discovery.engine_id; project_id = $project.id; project_path = $projectDir
        browser_url = "$BrowserOrigin/foreman"; backend_url = 'http://127.0.0.1:18090'
        agent = 'journey-worker'; source_uid = $sourceUID; agent_sha256 = $agentHash
        harness = 'pi'; version = '0.85.1'; model = 'openai-codex/gpt-6-astra'; effort = 'low'
        unauthenticated_state_status = $unauthenticatedStatus
        expires_at = $deadline.ToString('o'); klm_pid = $klm.Id; werbert_pid = $werbert.Id
    })
    "Disposable backends ready. Browser handoff: $(Join-Path $RunRoot 'ready.json')"
    "Create $(Join-Path $RunRoot 'stop.request') to stop early; maximum lifetime: $Minutes minutes."
    while ([DateTime]::UtcNow -lt $deadline -and -not (Test-Path -LiteralPath (Join-Path $RunRoot 'stop.request'))) {
        if ($klm.HasExited -or $werbert.HasExited) { throw 'An owned backend exited before the journey ended.' }
        Start-Sleep -Milliseconds 500
    }
} finally {
    # Seal connector delivery first. Windows termination is not a graceful Werbert close.
    try {
        if ($null -ne $werbert) {
            if (-not $werbert.HasExited) { $werbert.Kill() }
            $cleanup.werbert_exit_observed = $werbert.WaitForExit(10000)
        }
    } catch { $cleanup.werbert_stop_error = 'Owned Werbert termination failed; inspect its retained PID/handle.' }
    try {
        if ($null -ne $klm) {
            if (-not $klm.HasExited) {
                try { $null = Control 'stop' } catch { $cleanup.control_stop_error = 'Owned control stop unavailable.' }
                $cleanup.klm_exit_observed = $klm.WaitForExit(75000)
                if (-not $cleanup.klm_exit_observed) {
                    $klm.Kill()
                    $cleanup.forced_klm_termination = $true
                    $cleanup.klm_exit_observed = $klm.WaitForExit(10000)
                }
            } else { $cleanup.klm_exit_observed = $true }
        }
    } catch { $cleanup.klm_stop_error = 'Owned KLM termination failed; inspect its retained PID/handle.' }
    if (Test-Path -LiteralPath $authCopy) { Remove-Item -LiteralPath $authCopy -Force }
    $cleanup.native_auth_removed = -not (Test-Path -LiteralPath $authCopy)
    $cleanup.original_auth_unchanged = (Get-FileHash -LiteralPath $originalAuth -Algorithm SHA256).Hash -eq $authHash
    if (Test-Path -LiteralPath $agentPath) {
        $cleanup.saved_agent_unchanged = (Get-FileHash -LiteralPath $agentPath -Algorithm SHA256).Hash -eq $agentHash
    }
    Save-JSON (Join-Path $RunRoot 'cleanup.json') $cleanup
    if (-not $cleanup.original_auth_unchanged) { throw 'Original native auth changed during the window; inspect concurrent use before attributing cause.' }
    'Owned backend window ended. Read cleanup.json; retained data/logs are private validation artifacts.'
}
