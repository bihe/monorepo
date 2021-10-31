import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { catchError, timeout } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { AppInfo } from '../models/app.info.model';
import { BookmarkModel, BookmarkPathsModel, BoomarkSortOrderModel } from '../models/bookmarks.model';
import { ListResult, Result } from '../models/result.model';
import { BaseDataService } from './api.base.service';


@Injectable()
export class ApiBookmarksService extends BaseDataService {
  constructor (private http: HttpClient) {
    super();
  }

  private get Url(): string {
    return environment.apiBaseURL + '/api/v1/bookmarks';
  }

  getApplicationInfo(): Observable<AppInfo> {
    return this.http.get<AppInfo>(`${this.Url}/appinfo`, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getBookmarksForPath(path: string): Observable<ListResult<BookmarkModel[]>> {
    const url = `${this.Url}/bypath?path=${path}`;
    return this.http.get<ListResult<BookmarkModel[]>>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getBookmarkFolderByPath(path: string): Observable<Result<BookmarkModel>> {
    const url = `${this.Url}/folder?path=${path}`;
    return this.http.get<Result<BookmarkModel>>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getBookmarksByName(name: string): Observable<ListResult<BookmarkModel[]>> {
    const url = `${this.Url}/byname?name=${name}`;
    return this.http.get<ListResult<BookmarkModel[]>>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getMostVisitedBookmarks(num: number): Observable<ListResult<BookmarkModel[]>> {
    const url = `${this.Url}/mostvisited/${num}`;
    return this.http.get<ListResult<BookmarkModel[]>>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  fetchBookmarkById(id: string): Observable<BookmarkModel> {
    const url = `${this.Url}/${id}`;
    return this.http.get<BookmarkModel>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  createBookmark(model: BookmarkModel): Observable<Result<string>> {
    return this.http.post<Result<string>>(this.Url, model, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  deleteBookmarkById(id: string): Observable<Result<string>> {
    const url = `${this.Url}/${id}`;
    return this.http.delete<Result<string>>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  updateBookmark(model: BookmarkModel): Observable<Result<string>> {
    return this.http.put<Result<string>>(this.Url, model, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  updateBookmarksSortOrder(model: BoomarkSortOrderModel): Observable<Result<string>> {
    const url = `${this.Url}/sortorder`;
    return this.http.put<Result<string>>(url, model, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }

  getAllPaths(): Observable<BookmarkPathsModel> {
    const url = `${this.Url}/allpaths`;
    return this.http.get<BookmarkPathsModel>(url, this.RequestOptions)
      .pipe(
        timeout(this.RequestTimeOutDefault),
        catchError(this.handleError)
      );
  }
}
