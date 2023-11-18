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
