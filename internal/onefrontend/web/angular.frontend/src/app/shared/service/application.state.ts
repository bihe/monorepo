import { Injectable } from "@angular/core";
import { ReplaySubject } from 'rxjs';
import { AppInfo } from '../models/app.info.model';
import { ModuleInfo } from '../models/module.model';

const ShowAmountKey = 'mydms.amount.show';

@Injectable()
export class ApplicationState {
  private progress: ReplaySubject<boolean> = new ReplaySubject();
  private appInfo: ReplaySubject<AppInfo> = new ReplaySubject();
  private admin: ReplaySubject<boolean> = new ReplaySubject();
  private modInfo: ReplaySubject<ModuleInfo> = new ReplaySubject();
  private showAmount: ReplaySubject<boolean> = new ReplaySubject();
  private requestReload: ReplaySubject<boolean> = new ReplaySubject();
  private searchInput: ReplaySubject<string> = new ReplaySubject();
  private appRoute: ReplaySubject<string> = new ReplaySubject();
  private showSideBar: ReplaySubject<boolean> = new ReplaySubject();
  private mydmsVersion: ReplaySubject<AppInfo> = new ReplaySubject();
  private bookmarksVersion: ReplaySubject<AppInfo> = new ReplaySubject();
  private sitesVersion: ReplaySubject<AppInfo> = new ReplaySubject();

  public setAppInfo(data: AppInfo) {
    this.appInfo.next(data);
  }

  public getAppInfo(): ReplaySubject<AppInfo> {
    return this.appInfo;
  }

  public setProgress(data: boolean) {
    this.progress.next(data);
  }

  public getProgress(): ReplaySubject<boolean> {
    return this.progress;
  }

  public isAdmin(): ReplaySubject<boolean> {
    return this.admin;
  }

  public setAdmin(data: boolean) {
    this.admin.next(data);
  }

  public setModInfo(data: ModuleInfo) {
    this.modInfo.next(data);
  }

  public getModInfo(): ReplaySubject<ModuleInfo> {
    return this.modInfo;
  }

  public getSearchInput(): ReplaySubject<string> {
    return this.searchInput;
  }

  public setSearchInput(data: string) {
    this.searchInput.next(data);
  }

  public getShowAmount(): ReplaySubject<boolean> {
    const str = localStorage.getItem(ShowAmountKey);
    if (!str || str === undefined) {
      this.showAmount.next(false);
      return this.showAmount;
    }
    const parsed = JSON.parse(str);
    let val = false;
    if (parsed === true || parsed === 1 || parsed === "1" || parsed === "true" || parsed === "TRUE") {
      val = true;
    }
    this.showAmount.next(val);
    return this.showAmount;
  }

  public setShowAmount(show: boolean) {
    localStorage.setItem(ShowAmountKey, JSON.stringify(show));
    this.showAmount.next(show);
  }

  public setRequestReload(data: boolean) {
    this.requestReload.next(data);
  }

  public getRequestReload(): ReplaySubject<boolean> {
    return this.requestReload;
  }

  public setRoute(data: string) {
    this.appRoute.next(data);
  }

  public getRoute(): ReplaySubject<string> {
    return this.appRoute;
  }

  getMyDmsVersion(): ReplaySubject<AppInfo> {
    return this.mydmsVersion;
  }
  setMyDmsVersion(data: AppInfo) {
    this.mydmsVersion.next(data);
  }

  getBookmarksVersion(): ReplaySubject<AppInfo> {
    return this.bookmarksVersion;
  }
  setBookmarksVersion(data: AppInfo) {
    this.bookmarksVersion.next(data);
  }

  getSitesVersion(): ReplaySubject<AppInfo> {
    return this.sitesVersion;
  }
  setSitesVersion(data: AppInfo) {
    this.sitesVersion.next(data);
  }

  getShowSideBar(): ReplaySubject<boolean> {
    return this.showSideBar;
  }
  setShowSideBar(data: boolean) {
    this.showSideBar.next(data);
  }
}
