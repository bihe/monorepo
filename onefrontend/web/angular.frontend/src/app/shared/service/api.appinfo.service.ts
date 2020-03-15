import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { AppInfo } from '../models/app.info.model';
import { BaseDataService } from './api.base.service';

@Injectable()
export class ApiAppInfoService extends BaseDataService {
  constructor (private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiUrlOne + '/appinfo';
  }

  getApplicationInfo(): Observable<AppInfo> {
    return this.http.get<AppInfo>(this.Url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
