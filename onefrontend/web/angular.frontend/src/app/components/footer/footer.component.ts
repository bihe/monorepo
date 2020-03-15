import { Component, OnInit, VERSION } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { AppInfo } from '../../shared/models/app.info.model';
import { ApiAppInfoService } from '../../shared/service/api.appinfo.service';
import { ApplicationState } from '../../shared/service/application.state';
import { MessageUtils } from '../../shared/utils/message.utils';

@Component({
  selector: 'app-footer',
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.css']
})
export class FooterComponent implements OnInit {

  appData: AppInfo;
  year: number = new Date().getFullYear();

  constructor(
    private appInfoService: ApiAppInfoService,
    private appState: ApplicationState,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit(): void {
    this.appInfoService.getApplicationInfo()
      .subscribe(
        data => {
          this.appData = data;
          this.appData.uiRuntime = 'angular=' + VERSION.full;
          const adminRole = this.appData.userInfo.roles.find(x => x === 'Admin');
          if (adminRole) {
            this.appState.setAdmin(true);
          }
          this.appState.setAppInfo(data);
        },
        error => {
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error);
        }
      );
  }
}
