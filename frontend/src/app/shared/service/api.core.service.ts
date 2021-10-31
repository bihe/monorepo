import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { AppInfo, WhoAmI } from '../models/app.info.model';
import { UploadResult } from '../models/upload.result';
import { BaseDataService } from './api.base.service';


@Injectable()
export class ApiCoreService extends BaseDataService {
  constructor (private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiBaseURL + '/api/v1/core';
  }

  getApplicationInfo(): Observable<AppInfo> {
    return this.http.get<AppInfo>(`${this.Url}/appinfo`, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getWhoAmI(): Observable<WhoAmI> {
    return this.http.get<WhoAmI>(`${this.Url}/whoami`, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  uploadFile(file:File, pass:string, initPass:string): Observable<UploadResult> {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("pass", pass);
    formData.append("initPass", initPass);

    const httpOptions = {
      headers: new HttpHeaders({
        'Cache-Control': 'no-cache',
        'Pragma': 'no-cache'
      }),
      withCredentials: true
    };

    return this.http.post<UploadResult>(`${this.Url}/upload/file`, formData, httpOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
