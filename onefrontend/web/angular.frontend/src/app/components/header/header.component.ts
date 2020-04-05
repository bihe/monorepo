import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { DomSanitizer } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { debounceTime } from 'rxjs/operators';
import { ModuleInfo } from 'src/app/shared/models/module.model';
import { ApiBookmarksService } from 'src/app/shared/service/api.bookmarks.service';
import { ApiMydmsService } from 'src/app/shared/service/api.mydms.service';
import { ApiSiteService } from 'src/app/shared/service/api.sites.service';
import { AppInfo } from '../../shared/models/app.info.model';
import { ApplicationState } from '../../shared/service/application.state';
import { MessageUtils } from '../../shared/utils/message.utils';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit {
  appInfo: AppInfo;
  year: number = new Date().getFullYear();
  modInfo: ModuleInfo
  searchText = '';

  constructor(
    private state: ApplicationState,
    private snackBar: MatSnackBar,
    private sanitizer: DomSanitizer,
    private siteApi: ApiSiteService,
    private bookmarkApi: ApiBookmarksService,
    private mydmsApi: ApiMydmsService,
    private router: Router) {
  }

  ngOnInit() {
    

    this.state.getModInfo()
      .subscribe(
        data => {
          this.modInfo = data;
        },
        error => {
          new MessageUtils().showError(this.snackBar, error);
        }
      );

    this.state.getSearchInput().pipe(
      debounceTime(500))
      .subscribe(x => {
        if (this.searchText !== x) {
          this.searchText = x;
        }
      });

    this.state.getAppInfo()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      );

    this.siteApi.getApplicationInfo().subscribe(
      result => {
        this.state.setSitesVersion(result);
      }
    );
    this.bookmarkApi.getApplicationInfo().subscribe(
      result => {
        this.state.setBookmarksVersion(result);
      }
    );
    this.mydmsApi.getApplicationInfo().subscribe(
      result => {
        this.state.setMyDmsVersion(result);
      }
    );
  }

  onSearch(searchText: string) {
    this.state.setSearchInput(searchText);
  }
}
