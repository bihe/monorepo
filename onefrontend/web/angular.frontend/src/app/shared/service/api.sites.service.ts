import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { AppInfo } from '../models/app.info.model';
import { SiteInfo, UserSites } from '../models/usersites.model';
import { BaseDataService } from './api.base.service';


@Injectable()
export class ApiSiteService extends BaseDataService {
  constructor (private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiUrlSites + '/api/v1/sites';
  }

  getApplicationInfo(): Observable<AppInfo> {
    return this.http.get<AppInfo>(environment.apiUrlSites + '/api/v1/appinfo', this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getUserInfo(): Observable<UserSites> {
    return this.http.get<UserSites>(this.Url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  saveUserInfo(payload: SiteInfo[]): Observable<string> {
    return this.http.post<string>(this.Url, payload, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
