import { HttpErrorResponse, HttpHeaders } from '@angular/common/http';
import { throwError } from 'rxjs';
import { ErrorModel } from '../models/error.model';

export class BaseDataService {

  protected get RequestTimeOutDefault(): number { return 1000 * 60 * 1; }
  protected get RequestTimeOutLongRunning(): number { return 1000 * 60 * 10; }

  protected handleError (error: HttpErrorResponse | any) {
    let errorRaised = new ErrorModel();
    if (error instanceof HttpErrorResponse) {
      try {
        errorRaised.message = `'${error.statusText}' for url ${error.url}`;
        errorRaised.status = error.status;
        errorRaised.statusText = error.statusText;
      } catch (exception) {
        errorRaised.message = error.toString();
      }
    } else {
      errorRaised.message = error.message ? error.message : error.toString();
    }
    return throwError(errorRaised);
  }

  protected get RequestOptions() {
    const httpOptions = {
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'Cache-Control': 'no-cache',
        'Pragma': 'no-cache'
      }),
      withCredentials: true
    };
    return httpOptions;
  }
}