export class UserSites {
  public editable: boolean;
  public user: string;
  public userSites: SiteInfo[];
}

export class SiteInfo {
  public name: string;
  public url: string;
  public permissions: string[];
}

