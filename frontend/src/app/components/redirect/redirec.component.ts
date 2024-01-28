import { Component } from '@angular/core';

@Component({
  selector: 'app-bookmark-redirect',
  template: '<span><a href="/bm">Goto Bookmarks</a></span>',
})
export class Bookmarkv2RedirectCompnent {

  constructor() {
    location.href = '/bm';
  }
}

@Component({
  selector: 'app-sites-redirect',
  template: '<span><a href="/sites">Goto Sites</a></span>',
})
export class Sitesv2RedirectCompnent {

  constructor() {
    location.href = '/sites';
  }
}
