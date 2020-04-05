import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { Globals } from 'src/app/app.globals';
import { AppInfo } from 'src/app/shared/models/app.info.model';
import { ProblemDetail } from 'src/app/shared/models/error.problemdetail';
import { UserSites } from 'src/app/shared/models/usersites.model';
import { ModuleIndex, ModuleName } from 'src/app/shared/moduleIndex';
import { ApiSiteService } from 'src/app/shared/service/api.sites.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { ErrorMode, Errors } from 'src/app/shared/utils/errors';
import { MessageUtils } from 'src/app/shared/utils/message.utils';

@Component({
  selector: 'app-site-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class SiteHomeComponent implements OnInit {

  userInfo: UserSites;
  appInfo: AppInfo;

  constructor(private userService: ApiSiteService,
    private snackBar: MatSnackBar,
    private state: ApplicationState,
    private router: Router,
    private moduleIndex: ModuleIndex,
    private titleService: Title
  ) {
    this.state.setModInfo(this.moduleIndex.getModuleInfo(ModuleName.Sites));
    this.state.setRoute(this.router.url);
  }

  ngOnInit() {
    this.titleService.setTitle('Available Apps');
    this.userService.getUserInfo()
      .subscribe(
        data => {
          this.userInfo = data;
        },
        error => {
          if (Errors.CheckAuth(error) === ErrorMode.RedirectAuthFlow) {
            window.location.reload();
            return;
          }
          const errorDetail: ProblemDetail = error;
          this.state.setProgress(false);
          console.log(errorDetail);
          new MessageUtils().showError(this.snackBar, errorDetail.title);
        }
      );

      this.state.getSitesVersion()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      );
  }


  edit() {
    this.router.navigateByUrl(Globals.SitesPath + '/edit');
    return;
  }
}
