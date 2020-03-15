import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { UserSites } from 'src/app/shared/models/usersites.model';
import { ApiSiteService } from 'src/app/shared/service/api.sites.service';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';

@Component({
  selector: 'app-sites-edit',
  templateUrl: './edit.component.html',
  styleUrls: ['./edit.component.css']
})
export class EditSitesComponent implements OnInit {

  public jsonPayload: string;
  public isEditable = false;

  constructor(private userService: ApiSiteService,
    private state: ApplicationState,
    private snackBar: MatSnackBar,
    private router: Router,
  ) {}

  ngOnInit() {

    this.userService.getUserInfo()
      .subscribe(
        data => {
          this.isEditable = data.editable;
          if (!data.editable) {
            new MessageUtils().showError(this.snackBar, 'No permission to edit data!');
            this.router.navigateByUrl('/home');
            return;
          }

          this.jsonPayload = this.jsonify(data);
        },
        error => {
          console.log('Error: ' + error);
          new MessageUtils().showError(this.snackBar, error);
        }
      );
  }

  public save() {
    console.log('Save the JSON payload!');

    this.state.setProgress(true);

    let user: UserSites;
    user = JSON.parse(this.jsonPayload);
    this.userService.saveUserInfo(user.userSites)
      .subscribe(
        data => {
          console.log('saved!');
          this.state.setProgress(false);
          new MessageUtils().showSuccess(this.snackBar, "success!");
        },
        error => {
          this.state.setProgress(false);
          console.log(error.detail);
          new MessageUtils().showError(this.snackBar, error.detail);
        }
      );
  }

  private jsonify(data: UserSites) {
    return JSON.stringify(data, null, 4);
  }
}
