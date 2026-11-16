param([string]$Executable = '', [int]$Port = 3000)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
if (!$Executable) {
 $commit = & git -C $root rev-parse --short=7 HEAD
 if ($LASTEXITCODE -ne 0) { throw 'Could not read Git commit.' }
 $Executable = Join-Path $root "dist\SVNova-$commit.exe"
}
$testRoot = Join-Path $root ('.tools\acceptance-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force $testRoot | Out-Null
$data = Join-Path $testRoot 'work'
$source = Join-Path $testRoot 'source'
$bin = Join-Path $root '.tools\svn\bin'
& (Join-Path $bin 'svnadmin.exe') create $source
if ($LASTEXITCODE -ne 0) { throw 'Could not create fixture repository.' }
$fixture = Join-Path $testRoot 'fixture'
New-Item -ItemType Directory $fixture | Out-Null
Set-Content (Join-Path $fixture 'hello.txt') 'hello SVNova' -Encoding utf8
& (Join-Path $bin 'svn.exe') import $fixture ([uri]$source).AbsoluteUri -m initial --non-interactive --config-dir (Join-Path $testRoot 'svn-config')
if ($LASTEXITCODE -ne 0) { throw 'Could not import fixture content.' }
$process = Start-Process -FilePath $Executable -ArgumentList @('--work-path',$data,'--install-listen',"127.0.0.1:$Port") -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $testRoot 'stdout.log') -RedirectStandardError (Join-Path $testRoot 'stderr.log')
try {
 $base = "http://127.0.0.1:$Port"
 $session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
 $ready = $false
 for ($i=0; $i -lt 50; $i++) {
  try { $install = Invoke-WebRequest "$base/install" -WebSession $session -TimeoutSec 2; $ready=$true; break } catch { if ($process.HasExited) { throw (Get-Content (Join-Path $testRoot 'stderr.log') -Raw) }; Start-Sleep -Milliseconds 200 }
 }
 if (!$ready -or $install.Content -notmatch '首次安装') { throw 'Installation page unavailable or port already in use.' }
 $password = [guid]::NewGuid().ToString('N')
 $headers = @{Origin=$base}
 $null = Invoke-WebRequest "$base/install" -Method Post -WebSession $session -Headers $headers -Body @{name='admin';password=$password;origin=$base;listen="127.0.0.1:$Port";database=(Join-Path $data 'data\svnova.db');repositories=(Join-Path $data 'data\repositories')}
 foreach ($path in @('custom\conf\app.ini','data\svnova.db','data\secret.key','data\tools','log\svnova.log')) {
  if (!(Test-Path -LiteralPath (Join-Path $data $path))) { throw "Runtime layout missing: $path" }
 }
 if (Test-Path -LiteralPath (Join-Path $data 'config.json')) { throw 'Legacy config created.' }
 $landing = Invoke-WebRequest "$base/login" -Method Post -WebSession $session -Headers $headers -Body @{name='admin';password=$password;source='local'}
 $match = [regex]::Match($landing.Content,'name="csrf-token" content="([^"]+)"')
 if (!$match.Success) { throw 'Login did not establish a session.' }
 $headers['X-CSRF-Token']=$match.Groups[1].Value
 function Call-API([string]$Path,$Body) {
  $arguments=@{Uri="$base/api/v1/$Path";WebSession=$session;Headers=$headers}
  if ($null -ne $Body) {$arguments.Method='POST';$arguments.ContentType='application/json';$arguments.Body=($Body|ConvertTo-Json -Compress)}
  Invoke-RestMethod @arguments
 }
 $notification=Call-API 'admin/notifications' $null
 if ($notification.config.Enabled -or !$notification.config.SVNEnabled -or !$notification.config.LDAPEnabled -or $notification.configured) {throw 'Notification defaults are unsafe.'}
 $null=Call-API 'admin/notifications' @{Enabled=$false;SVNEnabled=$true;LDAPEnabled=$true;SVNFailureOnly=$true;SVNNoChange=$false;LDAPNoChange=$false;RetentionDays=90}
 $null=Call-API 'admin/ldap-schedule' @{Time='30 2 * * *'}
 $notificationPage=Invoke-WebRequest "$base/-/admin/notifications" -WebSession $session
 if ($notificationPage.Content -notmatch 'SVNFailureOnly' -or $notificationPage.Content -notmatch '发送测试消息') {throw 'Notification settings template missing from EXE.'}
 $authPage=Invoke-WebRequest "$base/-/admin/ldap" -WebSession $session
 if ($authPage.Content -notmatch '同步计划' -or $authPage.Content -notmatch [regex]::Escape('30 2 * * *')) {throw 'LDAP schedule template or saved value missing.'}
 $csrfRejected=Invoke-WebRequest "$base/api/v1/admin/notifications/test" -Method Post -WebSession $session -Headers @{Origin=$base;'X-CSRF-Token'='invalid-test-token'} -ContentType 'application/json' -Body '{}' -SkipHttpErrorCheck
 if ($csrfRejected.StatusCode -ne 403) {throw 'Notification test accepted invalid CSRF.'}
 $org=Call-API 'organizations' @{Name='acceptance';OwnerID=1}
 $repo=Call-API 'repositories' @{OrgID=$org.ID;Name='fixture';Kind='local';Source=$source;SourceStopped=$true;Interval=0}
 for ($i=0;$i -lt 40;$i++) { $repo=Call-API "repositories/$($repo.ID)" $null; if ($repo.Status -eq 'ready') {break}; if ($repo.Status -eq 'failed') { throw ((Call-API "repositories/$($repo.ID)/tasks" $null)|ConvertTo-Json) }; Start-Sleep -Milliseconds 500 }
 if ($repo.Status -ne 'ready') {throw 'Migration did not finish.'}
 $tree=Call-API "repositories/$($repo.ID)/tree?revision=1" $null
 if ($tree[0].Name -ne 'hello.txt') {throw 'Repository tree mismatch.'}
 $file=Call-API "repositories/$($repo.ID)/file?revision=1&path=hello.txt" $null
 if ($file.Content -notmatch 'hello SVNova') {throw 'Repository file mismatch.'}
 if (Test-Path $source) {throw 'Source was not moved.'}
 $conversionPage=Invoke-WebRequest "$base/repositories/$($repo.ID)?tab=settings&section=mirror" -WebSession $session
 if ($conversionPage.Content -notmatch '不逐修订比对历史内容' -or $conversionPage.Content -notmatch '快速转换为镜像') {throw 'Quick conversion template missing from EXE.'}
 $queuedConversion=Call-API "repositories/$($repo.ID)/convert-mirror" @{Remote='http://127.0.0.1:1/repository';Password='isolated-conversion-secret';FullAccess=$true;Interval=0}
 if (!$queuedConversion.queued -or !$queuedConversion.taskID -or !$queuedConversion.taskURL) {throw 'Conversion did not return a background task.'}
 $conversion=$null
 for($i=0;$i -lt 80;$i++) {
  $taskRows=Call-API "repositories/$($repo.ID)/tasks" $null
  $conversion=$taskRows|Where-Object ID -eq $queuedConversion.taskID
  if($conversion.State -eq 'failed'){break}
  Start-Sleep -Milliseconds 500
 }
 if($conversion.State -ne 'failed' -or $conversion.Kind -ne 'conversion' -or !$conversion.ActorID) {throw 'Isolated unreachable conversion did not fail safely.'}
 if($conversion.PSObject.Properties.Name -contains 'Source' -or ($conversion|ConvertTo-Json) -match 'isolated-conversion-secret') {throw 'Conversion candidate credential exposed.'}
 $repo=Call-API "repositories/$($repo.ID)" $null
 if($repo.Kind -ne 'local' -or $repo.Status -ne 'ready'){throw 'Failed conversion changed local configuration.'}
 $replacement=[guid]::NewGuid().ToString('N')
 $changed=Call-API 'user/password' @{CurrentPassword=$password;NewPassword=$replacement;ConfirmPassword=$replacement}
 if (!$changed.changed) {throw 'Password change failed.'}
 $revoked=Invoke-WebRequest "$base/api/v1/me" -WebSession $session -SkipHttpErrorCheck
 if ($revoked.StatusCode -ne 401) {throw 'Password change retained a session.'}
 $oldLogin=Invoke-WebRequest "$base/user/login" -Method Post -WebSession $session -Headers @{Origin=$base} -Body @{name='admin';password=$password;source='local'}
 if ($oldLogin.Content -notmatch '登录失败') {throw 'Old password still accepted.'}
 $newLogin=Invoke-WebRequest "$base/user/login" -Method Post -WebSession $session -Headers @{Origin=$base} -Body @{name='admin';password=$replacement;source='local'}
 $me=Invoke-RestMethod "$base/api/v1/me" -WebSession $session
 if ($me.Name -ne 'admin') {throw 'New password login failed.'}
 Stop-Process -Id $process.Id
 $process.WaitForExit()
 $process = Start-Process -FilePath $Executable -ArgumentList @('--work-path',$data) -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $testRoot 'restart-out.log') -RedirectStandardError (Join-Path $testRoot 'restart-err.log')
 $restarted = $false
 for ($i=0; $i -lt 50; $i++) {
  try { $me=Invoke-RestMethod "$base/api/v1/me" -WebSession $session -TimeoutSec 2; $restarted=$me.Name -eq 'admin'; if ($restarted) { break } } catch { if ($process.HasExited) { throw (Get-Content (Join-Path $testRoot 'restart-err.log') -Raw) }; Start-Sleep -Milliseconds 200 }
 }
 if (!$restarted) { throw 'INI restart did not preserve installation and session.' }
 $tree=Call-API "repositories/$($repo.ID)/tree?revision=1" $null
 if ($tree[0].Name -ne 'hello.txt') {throw 'Restart repository tree mismatch.'}
 $notification=Call-API 'admin/notifications' $null
 $schedule=Call-API 'admin/ldap-schedule' $null
 $deliveries=Call-API 'admin/notifications/deliveries' $null
 if ($notification.config.Enabled -or !$notification.config.SVNFailureOnly -or $schedule.Time -ne '30 2 * * *' -or $deliveries.total -ne 0) {throw 'Notification/schedule restart state mismatch or unexpected outbound work.'}
 Write-Output 'PASS: Gitea runtime layout, INI restart, single EXE extraction, installation, login, organization, local migration, browsing, password change/session revocation, notification defaults/settings/CSRF/templates, LDAP schedule persistence and quick conversion background queue/failure protection; no outbound messages queued.'
 Write-Output "Fixture and logs: $testRoot"
} finally {
 if (!$process.HasExited) {Stop-Process -Id $process.Id; $process.WaitForExit()}
}
