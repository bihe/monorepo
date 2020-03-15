import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, distinctUntilChanged, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { AppInfo } from '../models/app.info.model';
import { MyDmsDocument } from '../models/document.model';
import { DocumentResult } from '../models/document.result.model';
import { StringResult } from '../models/result.model';
import { SearchResult } from '../models/simple.search.result';
import { BaseDataService } from './api.base.service';


@Injectable()
export class ApiMydmsService extends BaseDataService {
  private readonly SEARCH_DOCUMENTS: string = '/documents/search';
  private readonly APP_INFO_URL: string = '/appinfo';
  private readonly SAVE_DOCUMENTS_URL: string = '/documents/';
  private readonly LOAD_DOCUMENT_URL: string = '/documents/%ID%';

  private readonly SEARCH_SENDERS_URL: string = '/documents/senders/search';
  private readonly SEARCH_TAGS_URL: string = '/documents/tags/search';

  constructor(private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiUrlMydms + '/api/v1';
  }

  getApplicationInfo(): Observable<AppInfo> {
    return this.http.get<AppInfo>(this.Url + this.APP_INFO_URL, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  searchDocuments(title: string, pageSize: number, skipEntries: number): Observable<DocumentResult> {
    const searchUrl = 'title=%TITLE%&limit=%LIMIT%&skip=%SKIP%';
    if (title && title !== '') {
      title = encodeURIComponent(title);
    }
    let url = this.Url + this.SEARCH_DOCUMENTS + '?' + searchUrl.replace('%TITLE%', title || '');
    if (!pageSize) {
      url = url.replace('%LIMIT%', '');
    } else {
      url = url.replace('%LIMIT%', pageSize.toString());
    }
    if (!skipEntries) {
      url = url.replace('%SKIP%', '');
    } else {
      url = url.replace('%SKIP%', skipEntries.toString());
    }

    return this.http.get<DocumentResult>(url, this.RequestOptions)
      .pipe(
        distinctUntilChanged(),
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  searchSenders(name: string): Observable<SearchResult> {
    const searchUrl = 'name=%NAME%';
    const url = this.Url + this.SEARCH_SENDERS_URL + '?' + searchUrl.replace('%NAME%', name || '');

    return this.http.get<SearchResult>(url, this.RequestOptions)
      .pipe(
        distinctUntilChanged(),
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  searchTags(name: string): Observable<SearchResult> {
    const searchUrl = 'name=%NAME%';
    const url = this.Url + this.SEARCH_TAGS_URL + '?' + searchUrl.replace('%NAME%', name || '');

    return this.http.get<SearchResult>(url, this.RequestOptions)
      .pipe(
        distinctUntilChanged(),
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  saveDocument(document: MyDmsDocument): Observable<StringResult> {
    return this.http.post<StringResult>(this.Url + this.SAVE_DOCUMENTS_URL, JSON.stringify(document), this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );

  }

  getDocument(id: string): Observable<MyDmsDocument> {
    const url = this.LOAD_DOCUMENT_URL.replace('%ID%', id || '-1');

    return this.http.get<MyDmsDocument>(this.Url + url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  deleteDocument(id: string): Observable<StringResult> {
    const url = this.LOAD_DOCUMENT_URL.replace('%ID%', id || '-1');

    return this.http.delete<StringResult>(this.Url + url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
