import { Component, OnDestroy, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { AppModules } from 'src/app/app.globals';
import { AppInfo, WhoAmI } from 'src/app/shared/models/app.info.model';
import { ProblemDetail } from 'src/app/shared/models/error.problemdetail';
import { SiteInfo, UserSites } from 'src/app/shared/models/usersites.model';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiSiteService } from 'src/app/shared/service/api.sites.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-site-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class SiteHomeComponent implements OnInit, OnDestroy {

  userInfo: UserSites;
  origSites: SiteInfo[];
  appInfo: AppInfo;
  // all subscriptions are held in this array, on destroy all active subscriptions are unsubscribed
  subscriptions: any[];

  constructor(private userService: ApiSiteService,
    private snackBar: MatSnackBar,
    private state: ApplicationState,
    private router: Router,
    private moduleIndex: ModuleIndex,
    private titleService: Title
  ) {
    this.subscriptions = [];
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.Sites));
    this.state.setRoute(this.router.url);
    this.state.setCurrentModule(AppModules.Sites);
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach(sub => {
      sub.unsubscribe();
    });
  }

  ngOnInit() {
    this.titleService.setTitle('Available Apps');
    this.state.setProgress(true);

    this.subscriptions.push(this.userService.getUserInfo()
      .subscribe(
        data => {
          this.userInfo = data;
          this.origSites = this.userInfo.userSites;
          this.state.setProgress(false);
        },
        error => {
          this.state.setProgress(false);
          if (error.status == 401 || error.status == 403 || !error.status) {
            this.state.setWhoAmI(new WhoAmI());
            if (!environment.production) {
              window.location.href='assets/noaccess.dev.html';
              return;
            }
            window.location.href='assets/noaccess.html';
            return;
          }
          const errorDetail: ProblemDetail = error;
          console.log(errorDetail);
          new MessageUtils().showError(this.snackBar, errorDetail.title);
        }
      ));

    this.subscriptions.push(this.state.getSitesVersion()
      .subscribe(
        x => {
          this.appInfo = x;
        }
    ));

    this.subscriptions.push(this.state.getSearchInput()
      .subscribe(
        search => {
          this.userInfo.userSites = this.origSites;
          if (search.module === AppModules.Sites) {
            let sites = this.origSites.filter(item => {
              if (item.name.toLowerCase().indexOf(search.term) > -1) {
                return true;
              }
            })
            this.userInfo.userSites = sites;
          }
        }
      )
    );
  }


  edit() {
    this.router.navigateByUrl(AppModules.Sites + '/edit');
    return;
  }
}
