import { ProblemDetail } from '../models/error.problemdetail';

export enum ErrorMode {
  Standard = 0,
  RedirectAuthFlow
}

export class Errors {
  static CheckAuth(error: any): ErrorMode {
    let e: ProblemDetail
    e = error as ProblemDetail
    if (e.status === 401 || e.status === 403) {
      return ErrorMode.RedirectAuthFlow
    }
    return ErrorMode.Standard;
  }
}
