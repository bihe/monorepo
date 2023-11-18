import { Component } from '@angular/core';

@Component({
    selector: 'app-bookmark-redirect',
    template: '<span></span>',
})
export class Bookmarkv2RedirectCompnent {

    constructor() {
        location.href = '/bm';
    }
}
