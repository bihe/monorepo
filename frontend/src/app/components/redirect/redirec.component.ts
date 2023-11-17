import { Component } from '@angular/core';

@Component({
    selector: 'app-bookmark-redirect',
})
export class Bookmarkv2RedirectCompnent {

    constructor() {
        location.href = '/bm';
    }
}
