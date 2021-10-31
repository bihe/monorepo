import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { UserSites } from '../models/usersites.model';
import { BaseDataService } from './api.base.service';


@Injectable()
export class ApiSiteService extends BaseDataService {
  constructor (private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiBaseURL + '/api/v1/sites';
  }

  getUserInfo(): Observable<UserSites> {
    return this.http.get<UserSites>(this.Url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  saveUserInfo(payload: UserSites): Observable<string> {
    return this.http.post<string>(this.Url, payload, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
