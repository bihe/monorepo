export class AppInfo {
  public userInfo: UserInfo;
  public versionInfo: VersionInfo;
  public uiRuntime: string; // dynamically added by frontend itself
}

export class UserInfo {
  public displayName: string;
  public userId: string;
  public userName: string;
  public email: string;
  public roles: string[];
}

export class VersionInfo {
  public version: string;
  public buildNumber: string;
}
