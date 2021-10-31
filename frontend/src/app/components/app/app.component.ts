import { AfterViewInit, Component, OnInit } from '@angular/core';
import { MatSlideToggleChange } from '@angular/material/slide-toggle';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ApplicationState } from 'src/app/shared/service/application.state';
import { MessageUtils } from 'src/app/shared/utils/message.utils';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit, AfterViewInit {
  title = 'app';
  currentRoute = '';
  navExpanded = false;
  showAmount = false;
  showProgress = false;

  constructor(private state: ApplicationState,private snackBar: MatSnackBar) {}

  ngOnInit() {
    this.state.getRoute().subscribe(
      data => {
        this.currentRoute = data;
      }
    );
    this.state.getShowAmount()
      .subscribe(
        x => {
          this.showAmount = x;
        }
      );
  }

  ngAfterViewInit() {
    // get rid of Error: ExpressionChangedAfterItHasBeenCheckedError
    setTimeout(() => {
      this.state.getProgress()
        .subscribe(
          data => {
            this.showProgress = data;
          },
          error => {
            new MessageUtils().showError(this.snackBar, error);
          }
        );
    },2000);
  }

  showAmountToggle(event: MatSlideToggleChange) {
    console.log('Change showAmount to ' + event.checked);
    this.state.setShowAmount(event.checked);
    this.state.setRequestReload(true);
  }
}
