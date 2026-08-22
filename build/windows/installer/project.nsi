Unicode true

!include "wails_tools.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"
VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true
SetCompressor /SOLID lzma

!include "MUI.nsh"
!include "nsDialogs.nsh"
!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING
!define MUI_LANGDLL_REGISTRY_ROOT HKCU
!define MUI_LANGDLL_REGISTRY_KEY "Software\OmniCred"
!define MUI_LANGDLL_REGISTRY_VALUENAME "Installer Language"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
Page custom DataDirectoryPageCreate DataDirectoryPageLeave
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "SimpChinese"

LangString DATA_PAGE_TITLE ${LANG_ENGLISH} "Choose data storage location"
LangString DATA_PAGE_TITLE ${LANG_SIMPCHINESE} "选择数据存储位置"
LangString DATA_PAGE_SUBTITLE ${LANG_ENGLISH} "Choose where OmniCred stores its local database."
LangString DATA_PAGE_SUBTITLE ${LANG_SIMPCHINESE} "选择 OmniCred 本地数据库的保存位置。"
LangString DATA_PAGE_LABEL ${LANG_ENGLISH} "Data folder:"
LangString DATA_PAGE_LABEL ${LANG_SIMPCHINESE} "数据文件夹："
LangString DATA_PAGE_NOTE ${LANG_ENGLISH} "The database file omnicred.db will be created in this folder. Existing settings are never overwritten."
LangString DATA_PAGE_NOTE ${LANG_SIMPCHINESE} "数据库文件 omnicred.db 将保存在此文件夹中。安装器不会覆盖已有设置。"
LangString DATA_PAGE_EXISTING ${LANG_ENGLISH} "Existing data settings were found. This installation will keep the current location. You can migrate the database later from OmniCred Settings."
LangString DATA_PAGE_EXISTING ${LANG_SIMPCHINESE} "检测到已有数据设置。本次安装将保留当前位置；安装完成后可在 OmniCred 设置页面迁移数据库。"
LangString DATA_PAGE_BROWSE ${LANG_ENGLISH} "Browse..."
LangString DATA_PAGE_BROWSE ${LANG_SIMPCHINESE} "浏览..."
LangString DATA_PAGE_DIALOG ${LANG_ENGLISH} "Choose the OmniCred data folder"
LangString DATA_PAGE_DIALOG ${LANG_SIMPCHINESE} "选择 OmniCred 数据文件夹"
LangString DATA_PAGE_REQUIRED ${LANG_ENGLISH} "Choose a data storage folder."
LangString DATA_PAGE_REQUIRED ${LANG_SIMPCHINESE} "请选择数据存储文件夹。"
LangString DATA_PAGE_CREATE_FAILED ${LANG_ENGLISH} "The selected folder could not be created. Choose another folder."
LangString DATA_PAGE_CREATE_FAILED ${LANG_SIMPCHINESE} "无法创建所选文件夹，请选择其他位置。"

Var DataDirectory
Var DataDirectoryInput
Var DataDirectoryBrowseButton
Var ExistingDataConfig

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"

!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif

ShowInstDetails show

Function .onInit
  !insertmacro MUI_LANGDLL_DISPLAY
  !insertmacro wails.checkArchitecture
  SetRegView 64
  StrCpy $DataDirectory "$APPDATA\OmniCred"
  StrCpy $ExistingDataConfig "0"
  IfFileExists "$APPDATA\OmniCred\config.json" 0 +2
    StrCpy $ExistingDataConfig "1"
  ReadRegStr $0 HKCU "Software\OmniCred" "InstallerDatabasePath"
  ${If} $0 != ""
    ${GetParent} "$0" $DataDirectory
  ${EndIf}
FunctionEnd

Function DataDirectoryPageCreate
  !insertmacro MUI_HEADER_TEXT "$(DATA_PAGE_TITLE)" "$(DATA_PAGE_SUBTITLE)"
  nsDialogs::Create 1018
  Pop $0
  ${If} $0 == error
    Abort
  ${EndIf}

  ${If} $ExistingDataConfig == "1"
    ${NSD_CreateLabel} 0 0 100% 48u "$(DATA_PAGE_EXISTING)"
    Pop $0
    nsDialogs::Show
    Return
  ${EndIf}

  ${NSD_CreateLabel} 0 0 100% 12u "$(DATA_PAGE_LABEL)"
  Pop $0
  ${NSD_CreateFileRequest} 0 18u 78% 13u "$DataDirectory"
  Pop $DataDirectoryInput
  ${NSD_CreateBrowseButton} 80% 18u 20% 13u "$(DATA_PAGE_BROWSE)"
  Pop $DataDirectoryBrowseButton
  ${NSD_OnClick} $DataDirectoryBrowseButton DataDirectoryBrowse
  ${NSD_CreateLabel} 0 43u 100% 30u "$(DATA_PAGE_NOTE)"
  Pop $0

  nsDialogs::Show
FunctionEnd

Function DataDirectoryBrowse
  Pop $0
  ${NSD_GetText} $DataDirectoryInput $DataDirectory
  nsDialogs::SelectFolderDialog "$(DATA_PAGE_DIALOG)" "$DataDirectory"
  Pop $0
  ${If} $0 != error
    StrCpy $DataDirectory $0
    ${NSD_SetText} $DataDirectoryInput $DataDirectory
  ${EndIf}
FunctionEnd

Function DataDirectoryPageLeave
  ${If} $ExistingDataConfig == "1"
    Return
  ${EndIf}

  ${NSD_GetText} $DataDirectoryInput $DataDirectory
  ${If} $DataDirectory == ""
    MessageBox MB_OK|MB_ICONEXCLAMATION "$(DATA_PAGE_REQUIRED)"
    Abort
  ${EndIf}

  GetFullPathName $DataDirectory "$DataDirectory"
  ClearErrors
  CreateDirectory "$DataDirectory"
  IfErrors 0 +3
    MessageBox MB_OK|MB_ICONEXCLAMATION "$(DATA_PAGE_CREATE_FAILED)"
    Abort
FunctionEnd

Function un.onInit
  !insertmacro MUI_UNGETLANGUAGE
FunctionEnd

Section
  !insertmacro wails.setShellContext
  !insertmacro wails.webview2runtime
  SetOutPath $INSTDIR
  !insertmacro wails.files
  IfFileExists "$APPDATA\OmniCred\config.json" data_location_done
  WriteRegStr HKCU "Software\OmniCred" "InstallerDatabasePath" "$DataDirectory\omnicred.db"
  data_location_done:
  CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
  CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
  !insertmacro wails.associateFiles
  !insertmacro wails.associateCustomProtocols
  !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
  !insertmacro wails.setShellContext
  RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"
  RMDir /r $INSTDIR
  Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
  Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
  !insertmacro wails.unassociateFiles
  !insertmacro wails.unassociateCustomProtocols
  !insertmacro wails.deleteUninstaller
SectionEnd
