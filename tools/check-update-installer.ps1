param([string]$MakeNSIS = "${env:ProgramFiles(x86)}\NSIS\makensis.exe")

$ErrorActionPreference = 'Stop'
$workspace = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$testRoot = Join-Path $workspace ('.cache\update-installer-' + [Guid]::NewGuid().ToString('N'))
[IO.Directory]::CreateDirectory($testRoot) | Out-Null
$updateInclude = Join-Path $workspace 'build\windows\installer\update.nsh'

# 测试安装器只写工作区内的占位文件，不触及正式安装目录、注册表和用户数据。
$source = @'
Unicode true
RequestExecutionLevel user
SilentInstall silent
!include "LogicLib.nsh"
!include "FileFunc.nsh"
!define PRODUCT_EXECUTABLE "fixture.exe"
!include /CHARSET=UTF8 "__INCLUDE__"
Name "OmniCred Update Test"
OutFile "__OUTPUT__"
Function .onInit
  Call InitUpdate
FunctionEnd
Section
  Call PrepareUpdate
  FileOpen $0 "$INSTDIR\fixture.exe" w
  FileWrite $0 "new-version"
  FileClose $0
  ${If} ${FileExists} "$INSTDIR\force-failure"
    SetErrorLevel 20
    Abort
  ${EndIf}
SectionEnd
'@
$source = $source.Replace('__INCLUDE__', $updateInclude).Replace('__OUTPUT__', (Join-Path $testRoot 'installer.exe'))
$sourcePath = Join-Path $testRoot 'fixture.nsi'
[IO.File]::WriteAllText($sourcePath, $source, [Text.UTF8Encoding]::new($false))
& $MakeNSIS /V2 $sourcePath
if ($LASTEXITCODE -ne 0) { throw '测试安装器编译失败' }

function Invoke-InstallerCase([string]$Name, [bool]$Failure, [bool]$AlreadyExited = $false) {
  $target = Join-Path $testRoot ($Name + ' 中文 空格')
  [IO.Directory]::CreateDirectory($target) | Out-Null
  $executable = Join-Path $target 'fixture.exe'
  [IO.File]::WriteAllText($executable, 'old-version')
  [IO.File]::WriteAllText((Join-Path $target 'data.db'), 'test-data-preserve')
  if ($Failure) { [IO.File]::WriteAllText((Join-Path $target 'force-failure'), '') }
  $parent = Start-Process -FilePath (Get-Process -Id $PID).Path -ArgumentList '-NoProfile', '-Command', 'Start-Sleep -Seconds 3' -WindowStyle Hidden -PassThru
  if ($AlreadyExited) { $parent.WaitForExit() }
  $installer = Start-Process -FilePath (Join-Path $testRoot 'installer.exe') -ArgumentList "/S /UPDATEPID=$($parent.Id) /D=$target" -WindowStyle Hidden -PassThru
  Start-Sleep -Milliseconds 500
  if (-not $AlreadyExited -and [IO.File]::ReadAllText($executable) -ne 'old-version') { throw "$Name : 未等待原进程退出" }
  if (-not $installer.WaitForExit(15000)) { throw "$Name : 安装器未退出" }
  $parent.WaitForExit()
  $expected = if ($Failure) { 'old-version' } else { 'new-version' }
  if ([IO.File]::ReadAllText($executable) -ne $expected) { throw "$Name : 升级或回滚结果错误" }
  if ([IO.File]::ReadAllText((Join-Path $target 'data.db')) -ne 'test-data-preserve') { throw "$Name : 数据被修改" }
  if ($Failure -and $installer.ExitCode -eq 0) { throw "$Name : 失败未返回错误码" }
  if (-not $Failure -and $installer.ExitCode -ne 0) { throw "$Name : 升级失败，退出码 $($installer.ExitCode)" }
  if (Test-Path -LiteralPath ($executable + '.update-backup')) { throw "$Name : 备份未清理" }
  Write-Output "$Name : 等待退出、中文空格路径、文件结果及数据保留检查通过"
}

Invoke-InstallerCase 'success' $false
Invoke-InstallerCase 'rollback' $true
Invoke-InstallerCase 'already-exited' $false $true
$invalid = Start-Process -FilePath (Join-Path $testRoot 'installer.exe') -ArgumentList '/S /UPDATEPID=invalid' -WindowStyle Hidden -PassThru
if (-not $invalid.WaitForExit(5000) -or $invalid.ExitCode -eq 0) { throw '无效更新参数未被拒绝' }
Write-Output '无效参数检查通过'
