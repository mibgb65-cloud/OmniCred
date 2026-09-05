; 更新协议 1：等待发起更新的进程退出，保留数据，失败时恢复旧程序。
Var UpdatePID
Var UpdateInstallDir
Var UpdateBackup

Function InitUpdate
  StrCpy $UpdatePID ""
  StrCpy $UpdateInstallDir $INSTDIR
  StrCpy $UpdateBackup ""
  ${GetParameters} $R0
  ClearErrors
  ${GetOptionsS} $R0 "/UPDATEPID=" $UpdatePID
  IfErrors update_init_done
  ${IfNot} ${Silent}
    SetErrorLevel 10
    Abort
  ${EndIf}
  ; PID 只允许十进制正整数，不接受混入其他参数的字符串。
  StrLen $R1 $UpdatePID
  ${If} $R1 == 0
  ${OrIf} $R1 > 10
    SetErrorLevel 10
    Abort
  ${EndIf}
  StrCpy $R2 0
  update_pid_loop:
    StrCpy $R3 $UpdatePID 1 $R2
    ${If} $R3 != "0"
    ${AndIf} $R3 != "1"
    ${AndIf} $R3 != "2"
    ${AndIf} $R3 != "3"
    ${AndIf} $R3 != "4"
    ${AndIf} $R3 != "5"
    ${AndIf} $R3 != "6"
    ${AndIf} $R3 != "7"
    ${AndIf} $R3 != "8"
    ${AndIf} $R3 != "9"
      SetErrorLevel 10
      Abort
    ${EndIf}
    IntOp $R2 $R2 + 1
    IntCmp $R2 $R1 0 update_pid_loop
  ${If} $UpdatePID == 0
    SetErrorLevel 10
    Abort
  ${EndIf}
  IfFileExists "$UpdateInstallDir\${PRODUCT_EXECUTABLE}" update_init_done
    SetErrorLevel 11
    Abort
  update_init_done:
FunctionEnd

Function PrepareUpdate
  ${If} $UpdatePID == ""
    Return
  ${EndIf}
  ; 使用进程句柄等待实际退出，避免固定延时造成 EXE、数据库或端口仍被占用。
  System::Call 'kernel32::OpenProcess(i 0x100000, i 0, i $UpdatePID) p.r0 ?e'
  Pop $2
  ${If} $0 P<> 0
    System::Call 'kernel32::WaitForSingleObject(p r0, i 60000) i.r1'
    System::Call 'kernel32::CloseHandle(p r0)'
    ${If} $1 != 0
      SetErrorLevel 12
      Abort
    ${EndIf}
  ${Else}
    ; ERROR_INVALID_PARAMETER 表示原进程已退出，其他错误不能视为退出成功。
    ${If} $2 != 87
      SetErrorLevel 12
      Abort
    ${EndIf}
  ${EndIf}
  StrCpy $UpdateBackup "$INSTDIR\${PRODUCT_EXECUTABLE}.update-backup"
  ClearErrors
  Delete "$UpdateBackup"
  ClearErrors
  Rename "$INSTDIR\${PRODUCT_EXECUTABLE}" "$UpdateBackup"
  ${If} ${Errors}
    StrCpy $UpdateBackup ""
    SetErrorLevel 13
    Abort
  ${EndIf}
FunctionEnd

Function .onInstSuccess
  ${If} $UpdateBackup != ""
    Delete "$UpdateBackup"
  ${EndIf}
FunctionEnd

Function .onInstFailed
  ${If} $UpdateBackup != ""
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    Rename "$UpdateBackup" "$INSTDIR\${PRODUCT_EXECUTABLE}"
  ${EndIf}
FunctionEnd
