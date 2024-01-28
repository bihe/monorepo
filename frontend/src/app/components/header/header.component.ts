import { Component, OnDestroy, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { Subject } from 'rxjs';
import { debounceTime, distinctUntilChanged, map, mergeMap } from 'rxjs/operators';
import { AppModules } from 'src/app/app.globals';
import { ErrorModel } from 'src/app/shared/models/error.model';
import { ModuleInfo } from 'src/app/shared/models/module.model';
import { SearchModel } from 'src/app/shared/models/search.model';
import { ApiCoreService } from 'src/app/shared/service/api.core.service';
import { ApiMydmsService } from 'src/app/shared/service/api.mydms.service';
import { environment } from 'src/environments/environment';
import { AppInfo, WhoAmI } from '../../shared/models/app.info.model';
import { ApplicationState } from '../../shared/service/application.state';
import { MessageUtils } from '../../shared/utils/message.utils';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit, OnDestroy {
  appInfo: AppInfo;
  year: number = new Date().getFullYear();
  modInfo: ModuleInfo
  searchText = '';
  showSideBar = false;
  showProgress = false;
  currentRoute = '';
  currentModule = AppModules.None;
  whoAmI: WhoAmI;
  txtQueryChanged: Subject<string> = new Subject<string>();

  // all subscriptions are held in this array, on destroy all active subscriptions are unsubscribed
  subscriptions: any[];

  constructor(
    private state: ApplicationState,
    private readonly router: Router,
    private snackBar: MatSnackBar,
    private coreApi: ApiCoreService,
    private mydmsApi: ApiMydmsService) {

    this.subscriptions = [];
    this.subscriptions.push(this.txtQueryChanged.pipe(debounceTime(500), distinctUntilChanged())
      .subscribe(model => {
        this.searchText = model;

        // Call your function which calls API or do anything you would like do after a lag of 1 sec
        let m = new SearchModel();
        m.module = this.currentModule;
        m.term = this.searchText;
        this.state.setSearchInput(m);
      })
    );
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach(sub => {
      sub.unsubscribe();
    });
  }

  ngOnInit() {
    this.subscriptions.push(this.state.getRoute()
      .subscribe(
        data => {
          this.currentRoute = data;
        }
      )
    );

    this.subscriptions.push(this.state.getCurrentModule()
      .subscribe(
        data => {
          this.currentModule = data;
        }
      )
    );

    this.state.setProgress(true);
    // chain the calls so that we have an overall result/error
    // the call will "break" if an error occurs and no subsequent call will be done
    this.subscriptions.push(this.coreApi.getWhoAmI()
      .pipe(
        map(result => {
          result.authenticated = true;
          this.state.setWhoAmI(result);
          this.whoAmI = result;
        }),
        // get the application-information for the given APIs
        mergeMap(() => this.coreApi.getApplicationInfo()),
        map(result => this.state.setSitesVersion(result)),

        mergeMap(() => this.mydmsApi.getApplicationInfo()),
        map(result => this.state.setMyDmsVersion(result)),

      ).subscribe(
        result => {
          console.log(result);
        },
        error => {
          this.state.setProgress(false);
          console.log(error);
          if (error instanceof ErrorModel) {
            if (error.status == 401 || error.status == 403 || !error.status) {
              this.state.setWhoAmI(new WhoAmI());
              if (!environment.production) {
                window.location.href = 'assets/noaccess.dev.html';
                return;
              }
              window.location.href = 'assets/noaccess.html';
              return;
            }
            new MessageUtils().showError(this.snackBar, error.message);
            return;
          }
          new MessageUtils().showError(this.snackBar, error.toString());
          return;
        }
      )
    );

    this.subscriptions.push(this.state.getModInfo()
      .subscribe(
        data => {
          this.modInfo = data;
        },
        error => {
          new MessageUtils().showError(this.snackBar, error);
        }
      )
    );

    this.subscriptions.push(this.state.getAppInfo()
      .subscribe(
        x => {
          this.appInfo = x;
        }
      )
    );

    // get rid of Error: ExpressionChangedAfterItHasBeenCheckedError
    setTimeout(() => {
      this.subscriptions.push(this.state.getProgress()
        .subscribe(
          data => {
            this.showProgress = data;
          },
          error => {
            new MessageUtils().showError(this.snackBar, error);
          }
        )
      );
    });

  }

  onFieldChange(query: string) {
    this.txtQueryChanged.next(query);
  }

  isCurrentRoute(route: string): boolean {
    let isCurrent = route === this.currentRoute;
    if (!isCurrent) {
      isCurrent = this.currentRoute.startsWith(route);
    }
    return isCurrent;
  }
}
