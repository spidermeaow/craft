#ifndef AppVersion
  #define AppVersion "0.1.11"
#endif

[Setup]
AppId={{575098CB-94E3-4A92-AAF3-C6827C74DC6F}
AppName=Craft
AppVersion={#AppVersion}
AppPublisher=Craft
DefaultDirName={localappdata}\Programs\Craft
DefaultGroupName=Craft
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0
OutputDir=..\..\dist
OutputBaseFilename=Craft-setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ChangesEnvironment=yes
ChangesAssociations=yes
UninstallDisplayIcon={app}\craft.exe
CloseApplications=no

[Files]
Source: "..\..\roadmap\phase1\output\phase1-rev11-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev10-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev9-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\dist\craft.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\docs\*"; DestDir: "{app}\docs"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\..\roadmap\phase1\output\phase1-rev1-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev2-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev3-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev4-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev5-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev6-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev7-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\roadmap\phase1\output\phase1-rev8-status.md"; DestDir: "{app}\roadmap\phase1\output"; Flags: ignoreversion
Source: "..\..\dist\craft-language-support-0.2.2.vsix"; DestDir: "{app}\editor"; Flags: ignoreversion
Source: "..\..\craft-vscode\README.md"; DestDir: "{app}\editor"; DestName: "LANGUAGE-SUPPORT.md"; Flags: ignoreversion
Source: "..\..\dist\craft-forge-file-icons-0.2.0.vsix"; DestDir: "{app}\editor"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\..\craft-file-icon-theme\README.md"; DestDir: "{app}\editor"; Flags: ignoreversion
Source: "..\..\examples\*"; DestDir: "{app}\examples"; Excludes: ".env,.env.*,*.clixml,.local\*,dist\*"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\..\packages\*"; DestDir: "{app}\packages"; Excludes: ".env,.env.*,*.clixml,.local\*,dist\*"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Craft Terminal"; Filename: "{cmd}"; Parameters: "/K """"{app}\craft.exe"" version"""; WorkingDir: "{userdocs}"
Name: "{group}\Craft User Guide"; Filename: "{sys}\notepad.exe"; Parameters: """{app}\docs\TRY-REV11.md"""

[Registry]
Root: HKCU; Subkey: "Software\Classes\.craft\OpenWithProgids"; ValueType: string; ValueName: "Craft.Source"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\Craft.Source"; ValueType: string; ValueData: "Craft source file"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\Craft.Source\shell\open\command"; ValueType: string; ValueData: """{sys}\notepad.exe"" ""%1"""
Root: HKCU; Subkey: "Software\CraftPhase1"; Flags: uninsdeletekey

[Code]
const
  EnvironmentKey = 'Environment';
  StateKey = 'Software\CraftPhase1';

function NormalizePath(Value: String): String;
begin
  Value := Trim(Value);
  if (Length(Value) >= 2) and (Value[1] = '"') and (Value[Length(Value)] = '"') then
    Value := Copy(Value, 2, Length(Value) - 2);
  while (Length(Value) > 3) and (Value[Length(Value)] = '\') do
    Delete(Value, Length(Value), 1);
  Result := Lowercase(Value);
end;

function HasPath(PathValue, Target: String): Boolean;
var Part: String; Separator: Integer;
begin
  Result := False;
  repeat
    Separator := Pos(';', PathValue);
    if Separator = 0 then begin Part := PathValue; PathValue := ''; end
    else begin Part := Copy(PathValue, 1, Separator - 1); Delete(PathValue, 1, Separator); end;
    if NormalizePath(Part) = NormalizePath(Target) then begin Result := True; Exit; end;
  until PathValue = '';
end;

procedure CurStepChanged(CurStep: TSetupStep);
var CurrentPath: String; ExitCode: Integer;
begin
  if CurStep = ssPostInstall then begin
    if not Exec(ExpandConstant('{app}\craft.exe'), 'version', '', SW_HIDE, ewWaitUntilTerminated, ExitCode) then
      RaiseException('Craft was installed but could not start. Please reinstall.');
    if ExitCode <> 0 then
      RaiseException('Craft installation check failed. Please reinstall.');
    RegQueryStringValue(HKCU, EnvironmentKey, 'Path', CurrentPath);
    if not HasPath(CurrentPath, ExpandConstant('{app}')) then begin
      if (CurrentPath <> '') and (CurrentPath[Length(CurrentPath)] <> ';') then
        CurrentPath := CurrentPath + ';';
      if not RegWriteExpandStringValue(HKCU, EnvironmentKey, 'Path', CurrentPath + ExpandConstant('{app}')) then
        RaiseException('Could not update user PATH. Add the Craft installation directory to PATH manually.');
      RegWriteStringValue(HKCU, StateKey, 'AddedPath', ExpandConstant('{app}'));
    end;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var CurrentPath, AddedPath, NewPath, Part: String; Separator: Integer; First: Boolean;
begin
  if CurUninstallStep = usUninstall then begin
    if RegQueryStringValue(HKCU, StateKey, 'AddedPath', AddedPath) and
       RegQueryStringValue(HKCU, EnvironmentKey, 'Path', CurrentPath) then begin
      NewPath := ''; First := True;
      repeat
        Separator := Pos(';', CurrentPath);
        if Separator = 0 then begin Part := CurrentPath; CurrentPath := ''; end
        else begin Part := Copy(CurrentPath, 1, Separator - 1); Delete(CurrentPath, 1, Separator); end;
        if NormalizePath(Part) <> NormalizePath(AddedPath) then begin
          if not First then NewPath := NewPath + ';';
          NewPath := NewPath + Part; First := False;
        end;
      until CurrentPath = '';
      RegWriteExpandStringValue(HKCU, EnvironmentKey, 'Path', NewPath);
    end;
  end;
end;

function UpdateReadyMemo(Space, NewLine, MemoUserInfoInfo, MemoDirInfo, MemoTypeInfo, MemoComponentsInfo, MemoGroupInfo, MemoTasksInfo: String): String;
begin
  Result := MemoDirInfo + NewLine + NewLine +
    'Craft will be added to your user PATH. Open a NEW terminal after installation.' + NewLine +
    'Compiler, interpreter and project template are embedded. Go is not required.';
end;


