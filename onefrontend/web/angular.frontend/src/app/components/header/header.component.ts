import { Component, OnInit } from '@angular/core';
import { MatSlideToggleChange } from '@angular/material/slide-toggle';
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

  menuVisible = false;
  showProgress = false;
  appInfo: AppInfo;
  year: number = new Date().getFullYear();
  modInfo: ModuleInfo
  showAmount = false;
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
    this.state.getProgress()
      .subscribe(
        data => {
          this.showProgress = data;
        },
        error => {
          new MessageUtils().showError(this.snackBar, error);
        }
      );

    this.state.getModInfo()
      .subscribe(
        data => {
          this.modInfo = data;
        },
        error => {
          new MessageUtils().showError(this.snackBar, error);
        }
      );

    this.state.getShowAmount()
      .subscribe(
        x => {
          this.showAmount = x;
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

  toggleMenu(visible: boolean) {
    this.menuVisible = visible;
  }

  menuTransform() {
    if (this.menuVisible) {
      return this.sanitizer.bypassSecurityTrustStyle('translateX(0)');
    } else {
      return this.sanitizer.bypassSecurityTrustStyle('translateX(-110%)');
    }
  }

  showAmountToggle(event: MatSlideToggleChange) {
    console.log('Change showAmount to ' + event.checked);
    this.state.setShowAmount(event.checked);
    this.state.setRequestReload(true);
  }

  onSearch(searchText: string) {
    this.state.setSearchInput(searchText);
  }
}
