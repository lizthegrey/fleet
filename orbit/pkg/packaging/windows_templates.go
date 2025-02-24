package packaging

import "text/template"

// Partially adapted from Launcher's wix XML in
// https://github.com/kolide/launcher/blob/master/pkg/packagekit/internal/assets/main.wxs.
var windowsWixTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`<?xml version="1.0" encoding="UTF-8"?>
<Wix xmlns="http://schemas.microsoft.com/wix/2006/wi" xmlns:util="http://schemas.microsoft.com/wix/UtilExtension">
  <Product
    Id="*"
    Name="Fleet osquery"
    Language="1033"
    Version="{{.Version}}"
    Manufacturer="Fleet Device Management (fleetdm.com)"
    UpgradeCode="B681CB20-107E-428A-9B14-2D3C1AFED244" >

    <Package
      Keywords='Fleet osquery'
      Description="Fleet osquery"
      InstallerVersion="500"
      Compressed="yes"
      InstallScope="perMachine"
      InstallPrivileges="elevated"
      Languages="1033" />

    <Property Id="REINSTALLMODE" Value="amus" />

    <Property Id="APPLICATIONFOLDER">
      <RegistrySearch Key="SOFTWARE\FleetDM\Orbit" Root="HKLM" Type="raw" Id="APPLICATIONFOLDER_REGSEARCH" Name="Path" />
    </Property>

    <Property Id="ARPNOREPAIR" Value="yes" Secure="yes" />
    <Property Id="ARPNOMODIFY" Value="yes" Secure="yes" />

    <MediaTemplate EmbedCab="yes" />

    <Property Id="POWERSHELLEXE">
      <RegistrySearch Id="POWERSHELLEXE"
                      Type="raw"
                      Root="HKLM"
                      Key="SOFTWARE\Microsoft\PowerShell\1\ShellIds\Microsoft.PowerShell"
                      Name="Path" />
    </Property>

    <MajorUpgrade AllowDowngrades="yes" />

    <Directory Id="TARGETDIR" Name="SourceDir">
      <Directory Id="ProgramFiles64Folder">
        <Directory Id="ORBITROOT" Name="Orbit">
          <Component Id="C_ORBITROOT" Guid="A7DFD09E-2D2B-4535-A04F-5D4DE90F3863">
            <CreateFolder>
              <PermissionEx Sddl="O:SYG:SYD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1200a9;;;BU)" />
            </CreateFolder>
          </Component>
          <Directory Id="ORBITBIN" Name="bin">
            <Directory Id="ORBITBINORBIT" Name="orbit">
              <Component Id="C_ORBITBIN" Guid="AF347B4E-B84B-4DD4-9C4D-133BE17B613D">
                <CreateFolder>
                  <PermissionEx Sddl="O:SYG:SYD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1200a9;;;BU)" />
                </CreateFolder>
                <File Source="root\bin\orbit\windows\{{ .OrbitChannel }}\orbit.exe">
                  <PermissionEx Sddl="O:SYG:SYD:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;0x1200a9;;;BU)" />
                </File>
                <Environment Id='OrbitUpdateInterval' Name='ORBIT_UPDATE_INTERVAL' Value='{{ .OrbitUpdateInterval }}' Action='set' System='yes' />
                <!--
                  ##############################################################################################
                  NOTE: We've seen some system fail to install the MSI when using Account="NT AUTHORITY\SYSTEM"
                  ##############################################################################################
                  -->
                <ServiceInstall
                  Name="Fleet osquery"
                  Account="LocalSystem"
                  ErrorControl="ignore"
                  Start="auto"
                  Type="ownProcess"
                  Description="This service runs Fleet's osquery runtime and autoupdater (Orbit)."
                  Arguments='--root-dir "[ORBITROOT]." --log-file "[System64Folder]config\systemprofile\AppData\Local\FleetDM\Orbit\Logs\orbit-osquery.log"{{ if .FleetURL }} --fleet-url "{{ .FleetURL }}"{{ end }}{{ if .FleetCertificate }} --fleet-certificate "[ORBITROOT]fleet.pem"{{ end }}{{ if .EnrollSecret }} --enroll-secret-path "[ORBITROOT]secret.txt"{{ end }}{{if .Insecure }} --insecure{{ end }}{{ if .Debug }} --debug{{ end }}{{ if .UpdateURL }} --update-url "{{ .UpdateURL }}"{{ end }}{{ if .DisableUpdates }} --disable-updates{{ end }}{{ if .Desktop }} --fleet-desktop --desktop-channel {{ .DesktopChannel }}{{ end }} --orbit-channel "{{ .OrbitChannel }}" --osqueryd-channel "{{ .OsquerydChannel }}"'
                >
                  <util:ServiceConfig
                    FirstFailureActionType="restart"
                    SecondFailureActionType="restart"
                    ThirdFailureActionType="restart"
                    ResetPeriodInDays="1"
                    RestartServiceDelayInSeconds="1"
                  />
                </ServiceInstall>
                <ServiceControl
                  Id="StartOrbitService"
                  Name="Fleet osquery"
                  Start="install"
                  Stop="both"
                  Remove="uninstall"
                />
              </Component>
            </Directory>
          </Directory>
        </Directory>
      </Directory>
    </Directory>

    <SetProperty Id="CA_UninstallOsquery"
                 Before ="CA_UninstallOsquery"
                 Sequence="execute"
                 Value='&quot;[POWERSHELLEXE]&quot; -NoLogo -NonInteractive -NoProfile -ExecutionPolicy Bypass -File "[ORBITROOT]installer_utils.ps1" -uninstallOsquery' />

    <CustomAction Id="CA_UninstallOsquery"
                  BinaryKey="WixCA"
                  DllEntry="WixQuietExec64"
                  Execute="deferred"
                  Return="check"
                  Impersonate="no" />

   <SetProperty Id="CA_RemoveOrbit"
                 Before ="CA_RemoveOrbit"
                 Sequence="execute"
                 Value='&quot;[POWERSHELLEXE]&quot; -NoLogo -NonInteractive -NoProfile -ExecutionPolicy Bypass -File "[ORBITROOT]installer_utils.ps1" -uninstallOrbit' />

    <CustomAction Id="CA_RemoveOrbit"
                  BinaryKey="WixCA"
                  DllEntry="WixQuietExec64"
                  Execute="deferred"
                  Return="check"
                  Impersonate="no" />  

    <InstallExecuteSequence>
      <Custom Action='CA_RemoveOrbit' Before='RemoveFiles'>(NOT UPGRADINGPRODUCTCODE) AND (REMOVE="ALL")</Custom> <!-- Only happens during uninstall -->
      <Custom Action='CA_UninstallOsquery' After='InstallFiles'>NOT Installed AND NOT WIX_UPGRADE_DETECTED</Custom> <!-- Only happens during first install -->
    </InstallExecuteSequence>

    <Feature Id="Orbit" Title="Fleet osquery" Level="1" Display="hidden">
      <ComponentGroupRef Id="OrbitFiles" />
      <ComponentRef Id="C_ORBITBIN" />
      <ComponentRef Id="C_ORBITROOT" />
    </Feature>

  </Product>
</Wix>
`))

var windowsOsqueryEventLogTemplate = template.Must(template.New("").Option("missingkey=error").Parse(
	`<?xml version="1.0"?>
<instrumentationManifest xsi:schemaLocation="http://schemas.microsoft.com/win/2004/08/events eventman.xsd" xmlns="http://schemas.microsoft.com/win/2004/08/events" xmlns:win="http://manifests.microsoft.com/win/2004/08/windows/events" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:trace="http://schemas.microsoft.com/win/2004/08/events/trace">
	<instrumentation>
		<events>
			<provider name="FleetDM" guid="{F7740E18-3259-434F-9759-976319968900}" symbol="OsqueryWindowsEventLogProvider" resourceFileName="%systemdrive%\Program Files\Orbit\bin\osqueryd\windows\{{ .OsquerydChannel }}\osqueryd.exe" messageFileName="%systemdrive%\Program Files\Orbit\bin\osqueryd\windows\{{ .OsquerydChannel }}\osqueryd.exe">
				<events>
					<event symbol="DebugMessage" value="1" version="0" channel="osquery" level="win:Warning" task="LogMessage" opcode="MessageOpcode" template="_template_message" keywords="DebugWindowsEventLogMessage " message="$(string.osquery.event.1.message)"></event>
					<event symbol="InfoMessage" value="2" version="0" channel="osquery" level="win:Informational" task="LogMessage" opcode="MessageOpcode" template="_template_message" keywords="InfoWindowsEventLogMessage " message="$(string.osquery.event.2.message)"></event>
					<event symbol="WarningMessage" value="3" version="0" channel="osquery" level="win:Warning" task="LogMessage" opcode="MessageOpcode" template="_template_message" keywords="WarningWindowsEventLogMessage " message="$(string.osquery.event.3.message)"></event>
					<event symbol="ErrorMessage" value="4" version="0" channel="osquery" level="win:Error" task="LogMessage" opcode="MessageOpcode" template="_template_message" keywords="ErrorWindowsEventLogMessage " message="$(string.osquery.event.4.message)"></event>
					<event symbol="FatalMessage" value="5" version="0" channel="osquery" level="win:Critical" task="LogMessage" opcode="MessageOpcode" template="_template_message" keywords="FatalWindowsEventLogMessage " message="$(string.osquery.event.5.message)"></event>
				</events>
				<levels></levels>
				<tasks>
					<task name="LogMessage" symbol="WindowsEventLogMessage" value="1" eventGUID="{D3C2B9E0-4AFE-41BD-99BE-F00EE4DFEB17}"></task>
				</tasks>
				<opcodes>
					<opcode name="MessageOpcode" symbol="_opcode_message" value="10"></opcode>
				</opcodes>
				<channels>
					<channel name="osquery" chid="osquery" symbol="OsqueryWindowsEventLogChannel" type="Admin" enabled="true" message="$(string.osquery.channel.PrimaryWindowsEventLogChannel.message)"></channel>
				</channels>
				<keywords>
					<keyword name="InfoWindowsEventLogMessage" symbol="_keyword_info_message" mask="0x1"></keyword>
					<keyword name="WarningWindowsEventLogMessage" symbol="_keyword_warning_message" mask="0x2"></keyword>
					<keyword name="ErrorWindowsEventLogMessage" symbol="_keyword_error_message" mask="0x4"></keyword>
					<keyword name="FatalWindowsEventLogMessage" symbol="_keyword_fatal_message" mask="0x8"></keyword>
					<keyword name="DebugWindowsEventLogMessage" symbol="_keyword_debug_message" mask="0x10"></keyword>
				</keywords>
				<templates>
					<template tid="_template_message">
						<data name="Message" inType="win:AnsiString" outType="xs:string"></data>
						<data name="Location" inType="win:AnsiString" outType="xs:string"></data>
					</template>
				</templates>
			</provider>
		</events>
	</instrumentation>
	<localization>
		<resources culture="en-US">
			<stringTable>
				<string id="osquery.event.5.message" value="Fatal error"></string>
				<string id="osquery.event.4.message" value="Error"></string>
				<string id="osquery.event.3.message" value="Warning"></string>
				<string id="osquery.event.2.message" value="Information"></string>
				<string id="osquery.event.1.message" value="Debug"></string>
				<string id="osquery.channel.PrimaryWindowsEventLogChannel.message" value="osquery"></string>
				<string id="level.Warning" value="Warning"></string>
				<string id="level.Verbose" value="Verbose"></string>
				<string id="level.Informational" value="Information"></string>
				<string id="level.Error" value="Error"></string>
				<string id="level.Critical" value="Critical"></string>
			</stringTable>
		</resources>
	</localization>
</instrumentationManifest>
`))

var windowsPSInstallerUtils = template.Must(template.New("").Option("missingkey=error").Parse(
	`param(
  [switch] $uninstallOsquery = $false,
  [switch] $uninstallOrbit = $false,
  [switch] $stopOrbit = $false,
  [switch] $help = $false
)

$ErrorActionPreference = "SilentlyContinue"

$code = @"
using Microsoft.Win32;
using System;
using System.IO;
using System.Runtime.InteropServices;
public class RegistryUtils
{

    [DllImport("kernel32.dll", SetLastError = true)]
    internal static extern IntPtr GetCurrentProcess();

    [DllImport("advapi32.dll", ExactSpelling = true, SetLastError = true)]
    internal static extern bool AdjustTokenPrivileges(IntPtr htok, bool disall,
        ref TokPriv1Luid newst, int len, IntPtr prev, IntPtr relen);

    [DllImport("advapi32.dll", ExactSpelling = true, SetLastError = true)]
    internal static extern bool OpenProcessToken(IntPtr h, int acc, ref IntPtr phtok);

    [DllImport("advapi32.dll", SetLastError = true)]
    internal static extern bool LookupPrivilegeValue(string host, string name, ref long pluid);

    [DllImport("advapi32.dll", SetLastError = true)]
    static extern int RegLoadKey(UInt32 hKey, String lpSubKey, String lpFile);

    [DllImport("advapi32.dll", SetLastError = true)]
    static extern int RegUnLoadKey(UInt32 hKey, string lpSubKey);

    [StructLayout(LayoutKind.Sequential, Pack = 1)]
    internal struct TokPriv1Luid
    {
        public int Count;
        public long Luid;
        public int Attr;
    }

    internal const int SE_PRIVILEGE_ENABLED = 0x00000002;
    internal const int SE_PRIVILEGE_DISABLED = 0x00000000;
    internal const int TOKEN_QUERY = 0x00000008;
    internal const int TOKEN_ADJUST_PRIVILEGES = 0x00000020;

    public static bool EnablePrivilege(string privilege, bool disable)
    {
        TokPriv1Luid tp;
        IntPtr htok = IntPtr.Zero;

        if (!OpenProcessToken(GetCurrentProcess(), TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, ref htok))
        {
            Console.WriteLine("EnablePrivilege - Failed to obtain handle to primary access token");
            return false;
        }

        tp.Count = 1;
        tp.Luid = 0;

        if (disable)
        {
            tp.Attr = SE_PRIVILEGE_DISABLED;
        }
        else
        {
            tp.Attr = SE_PRIVILEGE_ENABLED;
        }

        if (!LookupPrivilegeValue(null, privilege, ref tp.Luid))
        {
            Console.WriteLine("EnablePrivilege - Failed to lookup privilege {0}", privilege);
            return false;
        }

        if (!AdjustTokenPrivileges(htok, false, ref tp, 0, IntPtr.Zero, IntPtr.Zero))
        {
            Console.WriteLine("EnablePrivilege - Failed to modify privilege {0}", privilege);
            return false;
        }

        return true;
    }


    public static void LoadUsersHives()
    {
        try
        {
            if (EnablePrivilege("SeRestorePrivilege", false) && EnablePrivilege("SeBackupPrivilege", false))
            {
                string usersRoot = System.Environment.GetEnvironmentVariable("SystemDrive") + "\\users\\";
                string[] userDir = System.IO.Directory.GetDirectories(usersRoot, "*", System.IO.SearchOption.TopDirectoryOnly);

                foreach (string fullDirectoryName in userDir)
                {
                    string userDirName = new DirectoryInfo(fullDirectoryName).Name;
                    if (userDirName.Length == 0)
                    {
                        continue;
                    }

                    string targetFile = fullDirectoryName + "\\NTUSER.DAT";
                    if (!File.Exists(targetFile))
                    {
                        continue;
                    }

                    const uint HKEY_USERS = 0x80000003;
                    int retRegLoadKey = RegLoadKey(HKEY_USERS, userDirName, targetFile);
                    Console.WriteLine("RegLoadKey result was {0}", retRegLoadKey);
                }
            }
        }
        catch
        {

        }
    }

    public static void UnloadUsersHives()
    {
        try
        {
            if (EnablePrivilege("SeRestorePrivilege", false) && EnablePrivilege("SeBackupPrivilege", false))
            {
                string usersRoot = System.Environment.GetEnvironmentVariable("SystemDrive") + "\\users\\";
                string[] userDir = System.IO.Directory.GetDirectories(usersRoot, "*", System.IO.SearchOption.TopDirectoryOnly);

                foreach (string fullDirectoryName in userDir)
                {
                    var userDirName = new DirectoryInfo(fullDirectoryName).Name;
                    if (userDirName.Length == 0)
                    {
                        continue;
                    }

                    const uint HKEY_USERS = 0x80000003;
                    int retRegUnLoadKey = RegUnLoadKey(HKEY_USERS, userDirName);
                    Console.WriteLine("RegUnLoadKey result was {0}", retRegUnLoadKey);
                }
            }
        }
        catch
        {

        }
    }

    public static void RemoveOsqueryInstallationFromUserHives()
    {
        try
        {
            LoadUsersHives();

            foreach (var username in Registry.Users.GetSubKeyNames())
            {
                string targetKey = username + "\\SOFTWARE\\Microsoft\\Installer\\Products\\";
                RegistryKey rootKey = Registry.Users.OpenSubKey(targetKey, writable: true);

                if (rootKey != null)
                {
                    foreach (var productEntry in rootKey.GetSubKeyNames())
                    {
                        RegistryKey productKey = rootKey.OpenSubKey(productEntry);
                        if (productKey != null)
                        {
                            if ((string)productKey.GetValue("ProductName") == "osquery")
                            {
                                productKey.Close();
                                rootKey.DeleteSubKeyTree(productEntry);
                            }
                        }
                    }
                }
            }

            UnloadUsersHives();
        }
        catch (Exception ex)
        {
            Console.WriteLine("There was an exception when removing osquery installer from user hives: {0}", ex.ToString());
        }
    }
}
"@

Add-Type -TypeDefinition $code -Language CSharp

function Test-Administrator  
{  
    [OutputType([bool])]
    param()
    process {
        [Security.Principal.WindowsPrincipal]$user = [Security.Principal.WindowsIdentity]::GetCurrent();
        return $user.IsInRole([Security.Principal.WindowsBuiltinRole]::Administrator);
    }
}

function Do-Help {
  $programName = (Get-Item $PSCommandPath ).Name
  
  Write-Host "Usage: $programName (-uninstallOsquery|-uninstallOrbit|-stopOrbit|-help)" -foregroundcolor Yellow
  Write-Host ""
  Write-Host "  Only one of the following options can be used. Using multiple will result in "
  Write-Host "  options being ignored."
  Write-Host "    -uninstallOsquery         Uninstall Osquery"
  Write-Host "    -uninstallOrbit           Uninstall Orbit"
  Write-Host "    -stopOrbit                Stop Orbit"
  Write-Host ""
  Write-Host "    -help                     Shows this help screen"
  
  Exit 1
}

#Stops Osquery service and related processes
function Stop-Osquery {

  $kServiceName = "osqueryd"
  
  # Stop Service
  Stop-Service -Name $kServiceName
  Start-Sleep -Milliseconds 1000

  # Ensure that no process left running
  Get-Process -Name $kServiceName | Stop-Process -Force
}

#Stops Orbit service and related processes
function Stop-Orbit {

  # Stop Service
  Stop-Service -Name "Fleet osquery"
  Start-Sleep -Milliseconds 1000

  # Ensure that no process left running
  Get-Process -Name "orbit" | Stop-Process -Force
  Get-Process -Name "osqueryd" | Stop-Process -Force
  Get-Process -Name "fleet-desktop" | Stop-Process -Force
  Start-Sleep -Milliseconds 1000
}

#Revove Orbit footprint from registry and disk
function Force-Remove-Orbit {

  #Stoping Orbit
  Stop-Orbit

  #Remove Service
  $service = Get-WmiObject -Class Win32_Service -Filter "Name='Fleet osquery'"
  $service.delete()    

  #Removing Program files entries
  $targetPath = $Env:Programfiles + "\\Orbit"
  Remove-Item -LiteralPath $targetPath -Force -Recurse

  #Remove HKLM registry entries
  Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" -Recurse |  Where-Object {($_.ValueCount -gt 0)} | ForEach-Object {
    
    # Filter for osquery entries
    $properties = Get-ItemProperty $_.PSPath  |  Where-Object {($_.DisplayName -eq "Fleet osquery")}     
    if ($properties) {
     
      #Remove Registry Entries
      $regKey = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\" + $_.PSChildName
      Get-Item $regKey | Remove-Item -Force

      return
    }
  }   
}


#Revove Osquery footprint from registry and disk
function Force-Remove-Osquery {

  #Stoping Osquery
  Stop-Osquery

  #Remove Service
  $service = Get-WmiObject -Class Win32_Service -Filter "Name='osqueryd'"
  $service.delete() | Out-Null

  #Remove HKLM registry entries and disk footprint
  Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" -Recurse |  Where-Object {($_.ValueCount -gt 0)} | ForEach-Object {
    
    # Filter for osquery entries
    $properties = Get-ItemProperty $_.PSPath  |  Where-Object {($_.DisplayName -eq "osquery")}     
    if ($properties) {

      #Remove files from osquery location 
      if ($properties.InstallLocation){
        Remove-Item -LiteralPath $properties.InstallLocation -Force -Recurse
      }
      
      #Remove Registry Entries
      $regKey = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\" + $_.PSChildName
      Get-Item $regKey | Remove-Item -Force

      return
    }
  }   

  #Remove user entries if present
  [RegistryUtils]::RemoveOsqueryInstallationFromUserHives()
}

function Graceful-Product-Uninstall($productName) {

  if (!$productName) {
    Write-Host "Product name should be provided" -foregroundcolor Yellow
    return $false
  }

  # Grabbing the location of msiexec.exe
  $targetBinPath = Resolve-Path "$env:windir\system32\msiexec.exe"
  if (!(Test-Path $targetBinPath)) {
    Write-Host "msiexec.exe cannot be located." -foregroundcolor Yellow
    return $false
  }

  # Creating a COM instance of the WindowsInstaller.Installer COM object
  $Installer = New-Object -ComObject WindowsInstaller.Installer
  if (!$Installer) {
    Write-Host "There was a problem retrieving the installed packages." -foregroundcolor Yellow
    return $false
  }

  # Enumerating the installed packages
  $ProductEnumFlag = 7 #installed packaged enumeration flag
  $InstallerProducts = $Installer.ProductsEx("", "", $ProductEnumFlag); 
  if (!$InstallerProducts) {
    Write-Host "Installed packages cannot be retrieved." -foregroundcolor Yellow
    return $false
  }

  # Iterating over the installed packages results and checking for osquery package
  ForEach ($Product in $InstallerProducts) {

      $ProductCode = $null
      $VersionString = $null
      $ProductPath = $null

      try {
          $ProductCode = $Product.ProductCode()
          $VersionString = $Product.InstallProperty("VersionString")
          $ProductPath = $Product.InstallProperty("ProductName")
      }
      catch { }

      if ($ProductPath -like $productName) {
        Write-Host "Graceful uninstall of $ProductPath version $VersionString."  -foregroundcolor Cyan
        $InstallProcess = Start-Process $targetBinPath -ArgumentList "/quiet /x $ProductCode" -PassThru -Verb RunAs -Wait
        if ($InstallProcess.ExitCode -eq 0) {
          return $true
        } else {
          Write-Host "There was an error uninstalling osquery. Error code was: $($InstallProcess.ExitCode)." -foregroundcolor Yellow
          return $false
        }
      }
  }

  return $false
}


function Main {
  # Is Administrator check
  if (-not (Test-Administrator)) {
    Write-Host "Please run this script with Admin privileges!" -foregroundcolor Red
    Exit -1
  }

  # Help commands
  if ($help) {
    Do-Help
    Exit -1
  } 
   
  if ($uninstallOsquery) {
    Write-Host "About to uninstall Osquery." -foregroundcolor Yellow

    Stop-Osquery
  
    if (Graceful-Product-Uninstall("osquery")) {
      Write-Host "Osquery was gracefully uninstalled." -foregroundcolor Cyan
    } else {
      Force-Remove-Osquery
      Write-Host "Osquery was uninstalled." -foregroundcolor Cyan
    }
    Exit 0

  } elseif ($uninstallOrbit) {
    Write-Host "About to uninstall Orbit." -foregroundcolor Yellow

    Stop-Orbit
  
    if (Graceful-Product-Uninstall("Fleet osquery")) {
      Force-Remove-Orbit
      Write-Host "Orbit was gracefully uninstalled." -foregroundcolor Cyan
    } else {
      Force-Remove-Orbit
      Write-Host "Orbit was uninstalled." -foregroundcolor Cyan
    }
    Exit 0
  } elseif ($stopOrbit) {
    Write-Host "About to stop Orbit and remove it from system." -foregroundcolor Yellow

    Stop-Orbit

    Write-Host "Orbit was stopped." -foregroundcolor Cyan
    Exit 0
  } else {
    Write-Host "Invalid option selected: please see -help for usage details." -foregroundcolor Red
    Do-Help
    Exit -1

  }
}

$null = Main
`))
