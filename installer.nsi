!define PRODUCT_NAME "AutoLoginSZCU"
!define PRODUCT_SHORT_NAME "AutoLoginSZCU"
!define PRODUCT_VERSION "3.0"
!define PRODUCT_BUILD_VERSION "3.0.0.0"
!define PRODUCT_PUBLISHER "Tianli"
!define PRODUCT_COPYRIGHT "2023 Tianli."
!define PRODUCT_WEB_SITE "https://www.tianli0.top/"
!define PRODUCT_UNINSTALL_KEY "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"

Var AutoLogin.Header.SubText
Var AutoLogin.Component.MainText
Var AutoLogin.Component.InstallationType
Var AutoLogin.Component.InstallationTypeText
Var AutoLogin.Component.ComponentListTitleText
Var AutoLogin.Component.ComponentList
Var AutoLogin.Component.DescriptionTitleText
Var AutoLogin.Component.DescriptionText
Var AutoLogin.Component.DiskSizeText
Var AutoLogin.ExePath

Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "AutoLogin_Setup.exe"
Icon "source\favicon.ico"
InstallDir "C:\${PRODUCT_SHORT_NAME}"

VIAddVersionKey ProductName "${PRODUCT_NAME} Install Program"
VIAddVersionKey ProductVersion "${PRODUCT_VERSION}"
VIAddVersionKey Comments "${PRODUCT_NAME}"
VIAddVersionKey CompanyName "${PRODUCT_PUBLISHER}"
VIAddVersionKey LegalCopyright "${PRODUCT_COPYRIGHT}"
VIAddVersionKey FileVersion "${PRODUCT_BUILD_VERSION}"
VIAddVersionKey FileDescription "${PRODUCT_NAME} "
VIProductVersion "${PRODUCT_BUILD_VERSION}"

!include "MUI.nsh"

Caption "${PRODUCT_NAME} "
BrandingText /TRIMLEFT "${PRODUCT_NAME} ${PRODUCT_VERSION}"

!define MUI_ABORTWARNING
!define MUI_ICON "source\favicon.ico"

!define MUI_UNICON "source\favicon.ico"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP_STRETCH NoStretchNoCropNoAlign
!define MUI_HEADERIMAGE_RIGHT

!insertmacro MUI_PAGE_WELCOME
!define MUI_LICENSEPAGE_CHECKBOX
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!define MUI_PAGE_HEADER_TEXT "Choose Components"
!define MUI_COMPONENTSPAGE_TEXT_COMPLIST "Select the components you want to install."
!define MUI_PAGE_HEADER_SUBTEXT "choose components"
!define MUI_PAGE_CUSTOMFUNCTION_SHOW ComponentsPageShow
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
ShowInstDetails hide
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "AutoLoginSZCU"
!define MUI_FINISHPAGE_RUN_FUNCTION FinishRun
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_COMPONENTS
ShowUnInstDetails nevershow
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Function ComponentsPageShow
    FindWindow $0 "#32770" "" $HWNDPARENT
    GetDlgItem $AutoLogin.Header.SubText $HWNDPARENT 1038
    ShowWindow $AutoLogin.Header.SubText ${SW_HIDE}
    GetDlgItem $AutoLogin.Component.MainText $0 1006
    ShowWindow $AutoLogin.Component.MainText ${SW_HIDE}
    GetDlgItem $AutoLogin.Component.InstallationTypeText $0 1021
    GetDlgItem $AutoLogin.Component.InstallationType $0 1017
    ShowWindow $AutoLogin.Component.InstallationTypeText ${SW_HIDE}
    ShowWindow $AutoLogin.Component.InstallationType ${SW_HIDE}
    GetDlgItem $AutoLogin.Component.DescriptionTitleText $0 1042
    ShowWindow $AutoLogin.Component.DescriptionTitleText ${SW_HIDE}
    GetDlgItem $AutoLogin.Component.ComponentListTitleText $0 1022
    GetDlgItem $AutoLogin.Component.ComponentList $0 1032
    GetDlgItem $AutoLogin.Component.DiskSizeText $0 1023
    GetDlgItem $AutoLogin.Component.DescriptionText $0 1043
    System::Call "User32::SetWindowPos(i $AutoLogin.Component.ComponentListTitleText, i 0, i 0, i 0, i 270, i 13, i 0)"
    System::Call "User32::SetWindowPos(i $AutoLogin.Component.ComponentList, i 0, i 0, i 18, i 270, i 176, i 0)"
    System::Call "User32::SetWindowPos(i $AutoLogin.Component.DescriptionText, i 0, i 285, i 16, i 164, i 176, i 0)"
    System::Call "User32::SetWindowPos(i $AutoLogin.Component.DiskSizeText, i 0, i 0, i 210, i 0, i 0, i 1)"
FunctionEnd

Function FinishRun
    SetOutPath "$INSTDIR"
    ExecShell "" "$INSTDIR\Autologin.exe"
FunctionEnd

Section -AutoLogin
    SetOutPath "$INSTDIR"
    SetOverwrite ifnewer
    WriteUninstaller "$INSTDIR\uninstall.exe"
    File "Autologin.exe"
    SetOutPath "$INSTDIR\resources"
    File "/oname=uninstallerIcon.ico" "${MUI_ICON}"
    StrCpy $AutoLogin.ExePath "$INSTDIR\Autologin.exe"
    CreateShortCut "$DESKTOP\${PRODUCT_NAME}.lnk" "$AutoLogin.ExePath"
    IfFileExists "$SMPROGRAMS\${PRODUCT_NAME}" +2 0
        CreateDirectory "$SMPROGRAMS\${PRODUCT_NAME}"
    CreateShortCut "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk" "$AutoLogin.ExePath"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "DisplayName" "$(^Name)"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "InstallDir" "$INSTDIR"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\resources\uninstallerIcon.ico"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"
    WriteRegStr HKLM "${PRODUCT_UNINSTALL_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${PRODUCT_NAME}" "$AutoLogin.ExePath"
SectionEnd

Section -Uninstall
    Delete "$INSTDIR\uninstall.exe"
    Delete "$INSTDIR\repair.file"
    Delete "$INSTDIR\Autologin.exe"
    RMDir /r "$INSTDIR\bin"
    Delete "$INSTDIR\resources\uninstallerIcon.ico"
    RMDir "$INSTDIR\resources"
    Delete "$DESKTOP\${PRODUCT_NAME}.lnk"
    Delete "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk"
    RMDir "$SMPROGRAMS\${PRODUCT_NAME}"
    RMDir "$INSTDIR"
    DeleteRegKey HKLM "${PRODUCT_UNINSTALL_KEY}"
    DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "${PRODUCT_NAME}"
    SetAutoClose true
SectionEnd