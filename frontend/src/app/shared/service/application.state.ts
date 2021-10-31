import { Injectable } from "@angular/core";
import { ReplaySubject, Subject } from 'rxjs';
import { AppModules } from "src/app/app.globals";
import { AppInfo, WhoAmI } from '../models/app.info.model';
import { ModuleInfo } from '../models/module.model';
import { SearchModel } from "../models/search.model";

const ShowAmountKey = 'mydms.amount.show';

@Injectable()
export class ApplicationState {
  private progress: Subject<boolean> = new Subject();
  private appInfo: ReplaySubject<AppInfo> = new ReplaySubject();
  private admin: ReplaySubject<boolean> = new ReplaySubject();
  private modInfo: ReplaySubject<ModuleInfo> = new ReplaySubject();
  private showAmount: ReplaySubject<boolean> = new ReplaySubject();
  private requestReload: ReplaySubject<boolean> = new ReplaySubject();
  private searchInput: Subject<SearchModel> = new Subject();
  private appRoute: ReplaySubject<string> = new ReplaySubject();

  private mydmsVersion: ReplaySubject<AppInfo> = new ReplaySubject();
  private bookmarksVersion: ReplaySubject<AppInfo> = new ReplaySubject();
  private sitesVersion: ReplaySubject<AppInfo> = new ReplaySubject();

  private whoAmI: ReplaySubject<WhoAmI> = new ReplaySubject();

  private currentModule: ReplaySubject<AppModules> = new ReplaySubject();

  public setAppInfo(data: AppInfo) {
    this.appInfo.next(data);
  }

  public getAppInfo(): Subject<AppInfo> {
    return this.appInfo;
  }

  public setProgress(data: boolean) {
    this.progress.next(data);
  }

  public getProgress(): Subject<boolean> {
    return this.progress;
  }

  public isAdmin(): Subject<boolean> {
    return this.admin;
  }

  public setAdmin(data: boolean) {
    this.admin.next(data);
  }

  public setModInfo(data: ModuleInfo) {
    this.modInfo.next(data);
  }

  public getModInfo(): Subject<ModuleInfo> {
    return this.modInfo;
  }

  public getSearchInput(): Subject<SearchModel> {
    return this.searchInput;
  }

  public setSearchInput(data: SearchModel) {
    this.searchInput.next(data);
  }

  public getShowAmount(): Subject<boolean> {
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

  public getRequestReload(): Subject<boolean> {
    return this.requestReload;
  }

  public setRoute(data: string) {
    this.appRoute.next(data);
  }

  public getRoute(): Subject<string> {
    return this.appRoute;
  }


  public setWhoAmI(data: WhoAmI) {
    this.whoAmI.next(data);
  }

  public getWhoAmI(): Subject<WhoAmI> {
    return this.whoAmI;
  }




  public getMyDmsVersion(): Subject<AppInfo> {
    return this.mydmsVersion;
  }
  public setMyDmsVersion(data: AppInfo) {
    this.mydmsVersion.next(data);
  }

  public getBookmarksVersion(): Subject<AppInfo> {
    return this.bookmarksVersion;
  }
  public setBookmarksVersion(data: AppInfo) {
    this.bookmarksVersion.next(data);
  }

  public getSitesVersion(): Subject<AppInfo> {
    return this.sitesVersion;
  }
  public setSitesVersion(data: AppInfo) {
    this.sitesVersion.next(data);
  }


  public getCurrentModule(): Subject<AppModules> {
    return this.currentModule;
  }

  public setCurrentModule(data: AppModules) {
    this.currentModule.next(data);
  }


}
