import { Component, OnInit, VERSION } from '@angular/core';
import { MatSlideToggleChange } from '@angular/material/slide-toggle';
import { MatSnackBar } from '@angular/material/snack-bar';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { ApiAppInfoService } from 'src/app/shared/service/api.appinfo.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';

@Component({
  selector: 'app-navbar',
  templateUrl: './navbar.component.html',
  styleUrls: ['./navbar.component.css']
})
export class NavbarComponent implements OnInit {
  title = 'app';
  currentRoute = '';
  navExpanded = false;
  showAmount = false;

  appData: AppInfo;
  year: number = new Date().getFullYear();
  isSiteAdmin = false;
  showSideBar = false;

  constructor(private state: ApplicationState,
    private appInfoService: ApiAppInfoService,
    private snackBar: MatSnackBar) {

      this.appInfoService.getApplicationInfo()
      .subscribe(
        data => {
          this.appData = data;
          this.appData.uiRuntime = 'angular=' + VERSION.full;
          const adminRole = this.appData.userInfo.roles.find(x => x === 'Admin');
          if (adminRole) {
            this.state.setAdmin(true);
          }
          this.state.setAppInfo(data);
        },
        error => {
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error);
        }
      );
    }

  ngOnInit() {
    this.state.getRoute().subscribe(
      data => {
        this.currentRoute = data;
      }
    );
    // get the user-role for the site-application
    // if the user is admin, the menu can be shown
    this.state.getSitesVersion().subscribe(x => {
      if (x.userInfo.roles.findIndex(r => r.toLowerCase() === 'admin') > -1) {
        this.isSiteAdmin = true;
      }
    });

    this.state.getShowAmount()
      .subscribe(
        x => {
          this.showAmount = x;
        }
      );

    this.state.getShowSideBar()
    .subscribe(
      x => {
        this.showSideBar = x;
      }
    );
  }

  isCurrentRout(route: string): boolean {
    let isCurrent = route === this.currentRoute;
    if (!isCurrent) {
      isCurrent = this.currentRoute.startsWith(route);
    }
    return isCurrent;
  }

  toggleNavbar() {
    this.navExpanded = !this.navExpanded;
  }

  showAmountToggle(event: MatSlideToggleChange) {
    console.log('Change showAmount to ' + event.checked);
    this.state.setShowAmount(event.checked);
    this.state.setRequestReload(true);
  }
}
