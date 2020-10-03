import { HttpErrorResponse, HttpHeaders } from '@angular/common/http';
import { throwError } from 'rxjs';

export class BaseDataService {

  protected get RequestTimeOutDefault(): number { return 1000 * 60 * 1; }
  protected get RequestTimeOutLongRunning(): number { return 1000 * 60 * 10; }

  protected handleError (error: HttpErrorResponse | any) {
    let errorRaised: any;
    if (error instanceof HttpErrorResponse) {
      try {
        errorRaised = error.error;
      } catch (exception) {
        errorRaised = error.toString();
      }
    } else {
      errorRaised = error.message ? error.message : error.toString();
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
